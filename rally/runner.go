package rally

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"time"

	"github.com/gophercloud/utils/openstack/clientconfig"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"

	"opendev.org/vexxhost/rally-exporter/models"
)

// TaskLabels has the common labels used by every task.
var TaskLabels = []string{"title"}

// PeriodicRunner defines the structure of the collector
type PeriodicRunner struct {
	prometheus.Collector

	CloudName string
	TaskFile  string

	TaskDuration *prometheus.Desc
	TaskSLADesc  *prometheus.Desc
	TaskTime     *prometheus.Desc
}

// NewPeriodicRunner creates a rally runner which executes every 5 minutes
func NewPeriodicRunner(cloudName string, taskFile string) *PeriodicRunner {
	return &PeriodicRunner{
		CloudName: cloudName,
		TaskFile:  taskFile,
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

// Run starts the PeriodicRunner which runs Rally every 5 minutes.
func (runner *PeriodicRunner) Run() {
	runner.createDeployment()

	for {
		log.Info("Starting Rally run...")
		cmd := exec.Command("rally", "task", "start", runner.TaskFile)
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Error("Failed test, output below:")
		} else {
			log.Info("Successful run, output below:")
		}
		fmt.Println(string(output))

		time.Sleep(5 * time.Minute)
	}
}

func (runner *PeriodicRunner) createDeployment() {
	opts := &clientconfig.ClientOpts{Cloud: runner.CloudName}
	cloud, err := clientconfig.GetCloudFromYAML(opts)
	if err != nil {
		log.Fatal(err)
	}

	deployment := &Deployment{
		OpenStackDeployment: OpenStackDeployment{
			AuthURL:      cloud.AuthInfo.AuthURL,
			RegionName:   cloud.RegionName,
			EndpointType: cloud.EndpointType,
			Users: []OpenStackUser{
				OpenStackUser{
					Username:          cloud.AuthInfo.Username,
					Password:          cloud.AuthInfo.Password,
					UserDomainName:    cloud.AuthInfo.UserDomainName,
					ProjectName:       cloud.AuthInfo.ProjectName,
					ProjectDomainName: cloud.AuthInfo.ProjectDomainName,
				},
			},
		},
	}

	b, err := json.Marshal(deployment)
	if err != nil {
		log.Fatal(err)
	}

	file, err := ioutil.TempFile("/tmp", "deployment")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(file.Name())
	_, err = file.Write(b)
	if err != nil {
		log.Fatal(err)
	}

	cmd := exec.Command("rally", "deployment", "create", "--filename", file.Name(), "--name", runner.CloudName)
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
	db, err := gorm.Open("sqlite3", "/home/rally/data/rally.db")
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
