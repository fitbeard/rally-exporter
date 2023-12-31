package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/fitbeard/rally-exporter/rally"
)

func main() {
	var (
		listenAddress = kingpin.Flag(
			"web.listen-address",
			"Address on which to expose metrics and web interface.",
		).Default(":9355").String()
		metricsPath = kingpin.Flag(
			"web.telemetry-path",
			"Path under which to expose metrics.",
		).Default("/metrics").String()
		deployment = kingpin.Flag(
			"deployment-name",
			"Name of the Rally deployment",
		).Required().String()
		exectime = kingpin.Flag(
			"execution-time",
			"Wait X minutes before next run. Default: 5",
		).Default("5").Int()
		taskcount = kingpin.Flag(
			"task-history",
			"Number of tasks to keep in history. Default: 10",
		).Default("10").Int()
	)

	kingpin.Version(version.Print("rally-exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	// Get rid of any additional metrics
    // we have to expose our metrics with a custom registry
	registry := prometheus.NewRegistry()

	runner := rally.NewPeriodicRunner(*deployment, *exectime, *taskcount)

    registry.MustRegister(runner)

	go runner.Run()

    handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	http.Handle("/metrics", handler)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(`<html>
			<head><title>OpenStack Rally Exporter</title></head>
			<body>
			<h1>OpenStack Rally Exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			</body>
			</html>`))
		if err != nil {
			log.Error(err)
		}
	})

	log.Infoln("Listening on", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
