package rally

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"

	"github.com/KouroshVivan/rally-exporter/models"
)

// TaskLabels has the common labels used by every task.
var TaskLabels = []string{"name"}

// PeriodicRunner defines the structure of the collector
type PeriodicRunner struct {
	prometheus.Collector

	CloudName       string
	ExecTime        int
	TaskCount       int
	FailedTaskCount int

	TaskDuration       *prometheus.Desc
	AtomicTaskDuration *prometheus.Desc
	TaskSLADesc        *prometheus.Desc
	TaskTime           *prometheus.Desc
}

// NewPeriodicRunner creates a rally runner which executes every
// `execTime` minutes after last run.
func NewPeriodicRunner(cloudName string, execTime int, taskCount int) *PeriodicRunner {
	return &PeriodicRunner{
		CloudName: cloudName,
		ExecTime:  execTime,
		TaskCount: taskCount,
		TaskDuration: prometheus.NewDesc(
			"rally_task_duration",
			"Rally task duration",
			TaskLabels, nil,
		),
		AtomicTaskDuration: prometheus.NewDesc(
			"rally_subtask_duration",
			"Rally subtask duration",
			append(TaskLabels, "step"), nil,
		),
		TaskSLADesc: prometheus.NewDesc(
			"rally_task_passed",
			"Rally task passed",
			TaskLabels, nil,
		),
		TaskTime: prometheus.NewDesc(
			"rally_task_time",
			"Rally last run time",
			[]string{}, nil,
		),
	}
}

// Run starts the PeriodicRunner which runs Rally every
// `execTime` minutes after last run.
func (runner *PeriodicRunner) Run() {
	runner.createDeployment()

	for {
		currentTime := time.Now()

		log.Info("Remove old tasks")
		runner.removeOldTasks(runner.TaskCount)

		log.Info("Starting Rally run at ", currentTime.Format(time.DateTime))
		cmd := exec.Command("rally", "task", "start", "/conf/tasks.yml", "--task-args-file", "/conf/arguments.yml")
		if _, err := cmd.CombinedOutput(); err != nil {
			log.Error("Failed test")
		} else {
			log.Info("Successful run")
		}

		minutes := runner.ExecTime
		time.Sleep(time.Duration(minutes) * time.Minute)
	}
}

func (runner *PeriodicRunner) createDeployment() {

	precmd := exec.Command("rally", "deployment", "destroy", runner.CloudName)
	if _, err := precmd.CombinedOutput(); err != nil {
		log.Warn("There is no Rally deployment named '", runner.CloudName, "' to destroy")
	} else {
		log.Info("Successfully destroyed Rally deployment named '", runner.CloudName, "'")
	}

	cmd := exec.Command("rally", "deployment", "create", "--filename", "/conf/deployment.yml", "--name", runner.CloudName)
	if output, err := cmd.CombinedOutput(); err != nil {
		fmt.Println(string(output))
		log.Fatal("Failed to install Rally deployment: ", err)
	}
}
func (runner *PeriodicRunner) removeOldTasks(taskcount int) {

	if out, err := exec.Command("rally", "task", "list", "--uuids-only").Output(); err != nil {
		log.Warn("Cannot list tasks: ", err)
	} else {
		taskUUIDs := strings.Split(string(out), "\n")
		log.Info("Found ", len(taskUUIDs), " tasks (", taskcount, " to keep)")
		if len(taskUUIDs) > taskcount {
			taskUUIDs = taskUUIDs[:len(taskUUIDs)-taskcount]
			if _, err := exec.Command("rally", "task", "delete", "--uuid", strings.Join(taskUUIDs, " ")).Output(); err != nil {
				log.Warn("Cannot delete tasks: ", err)
			} else {
				log.Info("Sucessfully removed ", taskUUIDs, " tasks")
			}
		}
	}

}

// Describe provides all the descriptions of all the information for the collector
func (runner *PeriodicRunner) Describe(ch chan<- *prometheus.Desc) {
	ch <- runner.AtomicTaskDuration
	ch <- runner.TaskDuration
	ch <- runner.TaskSLADesc
	ch <- runner.TaskTime
}

func getLatestTask(db *gorm.DB) (*models.Task, error) {
	task := &models.Task{}
	err := db.Not("status", []string{"running"}).Last(task).Error

	return task, err
}

func (runner PeriodicRunner) GetTasks() string {
	db, err := gorm.Open("sqlite3", "/home/rally/.rally/rally.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	tasks := []models.Task{}
	if err := db.Find(&tasks).Error; err != nil {
		log.Fatal(err)
	}

	htmlResponse := "<h1>Tasks</h1>\n"
	for _, task := range tasks {
		htmlResponse += fmt.Sprintf("<li>%s %s %f %s <a href=\"/%s\">report</a></li>\n", task.UUID, task.CreatedAt.Format(time.DateTime), task.TaskDuration, task.Status, task.UUID)
	}

	return htmlResponse
}

func (runner PeriodicRunner) GenReport(TaskUUID string) ([]byte, error) {
	return exec.Command("rally", "task", "report", "--html", "--uuid", TaskUUID).Output()
}

// Collect grabs all the data from the Rally database
func (runner *PeriodicRunner) Collect(ch chan<- prometheus.Metric) {
	db, err := gorm.Open("sqlite3", "/home/rally/.rally/rally.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	task, err := getLatestTask(db)
	if err != nil {
		log.Error(err)
		return
	}

	subtasks := []models.Subtask{}
	if err := db.Where(&models.Subtask{TaskUUID: task.UUID}).Find(&subtasks).Error; err != nil {
		log.Fatal(err)
	}

	for _, subtask := range subtasks {
		ch <- prometheus.MustNewConstMetric(
			runner.TaskDuration,
			prometheus.GaugeValue,
			subtask.Duration, subtask.Title,
		)

		passSLA := float64(0)
		if subtask.PassSLA && subtask.Status == "finished" {
			passSLA = float64(1)
		}
		ch <- prometheus.MustNewConstMetric(
			runner.TaskSLADesc,
			prometheus.GaugeValue,
			passSLA, subtask.Title,
		)

		// Generate atomics action metrics
		workloads := []models.Workload{}
		if err := db.Where(&models.Workload{SubtaskUUID: subtask.UUID}).Find(&workloads).Error; err != nil {
			log.Fatal(err)
		}
		for _, workload := range workloads {
			StatisticsData := workload.Statistics
			StatisticsJSON := models.StatisticsFormat{}
			if err := json.Unmarshal([]byte(StatisticsData), &StatisticsJSON); err != nil {
				log.Fatal(err)
			}
			for _, Atomic := range StatisticsJSON.Durations.Atomics {
				ch <- prometheus.MustNewConstMetric(
					runner.AtomicTaskDuration,
					prometheus.GaugeValue,
					Atomic.Data.Avg, subtask.Title, Atomic.Name,
				)
			}
		}
	}

	ch <- prometheus.MustNewConstMetric(
		runner.TaskTime,
		prometheus.GaugeValue,
		float64(task.UpdatedAt.Unix()),
	)
}
