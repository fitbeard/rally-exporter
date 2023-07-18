package rally

import (
	"fmt"
	"os/exec"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"

	"github.com/fitbeard/rally-exporter/models"
)

// TaskLabels has the common labels used by every task.
var TaskLabels = []string{"title"}

// PeriodicRunner defines the structure of the collector
type PeriodicRunner struct {
	prometheus.Collector

	CloudName string
	ExecTime  int
	TaskCount int

	TaskDuration *prometheus.Desc
	TaskSLADesc  *prometheus.Desc
	TaskTime     *prometheus.Desc
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

		count := strconv.Itoa(runner.TaskCount)

		log.Info("Deleting last Rally task")
        // This is horrible. Should be rewritten in native golang.
		precmd := exec.Command("/delete-tasks.sh", count)
		preoutput, err := precmd.CombinedOutput()
		if err != nil {
			log.Error("Failed to delete last Rally task:")
		} else {
			log.Info("Deleted last Rally task:")
		}
		fmt.Println(string(preoutput))

		log.Info("Starting Rally run at ", currentTime.String())
		cmd := exec.Command("rally", "task", "start", "/conf/tasks.yml", "--task-args-file", "/conf/arguments.yml")
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Error("Failed test, output below:")
		} else {
			log.Info("Successful run, output below:")
		}
		fmt.Println(string(output))

		minutes := runner.ExecTime
		time.Sleep(time.Duration(minutes) * time.Minute)
	}
}

func (runner *PeriodicRunner) createDeployment() {

	precmd := exec.Command("rally", "deployment", "destroy", runner.CloudName)
	preoutput, err := precmd.CombinedOutput()
	if err != nil {
		log.Warn("There is no Rally deployment named '", runner.CloudName, "' to destroy.")
	} else {
		log.Info("Successfully destroyed Rally deployment named '", runner.CloudName, "'.")
	}
	fmt.Println(string(preoutput))

	cmd := exec.Command("rally", "deployment", "create", "--filename", "/conf/deployment.yml", "--name", runner.CloudName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(output))
		log.Fatal("Failed to install Rally deployment: ", err)
	}
}

// Describe provides all the descriptions of all the information for the collector
func (runner *PeriodicRunner) Describe(ch chan<- *prometheus.Desc) {
	ch <- runner.TaskDuration
	ch <- runner.TaskSLADesc
	ch <- runner.TaskTime
}

func getLatestTask(db *gorm.DB) (*models.Task, error) {
	task := &models.Task{}
	err := db.Not("status", []string{"running"}).Last(task).Error

	return task, err
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
	}

	ch <- prometheus.MustNewConstMetric(
		runner.TaskTime,
		prometheus.GaugeValue,
		float64(task.UpdatedAt.Unix()),
	)
}
