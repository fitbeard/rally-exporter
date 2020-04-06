package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"

	"opendev.org/vexxhost/rally-exporter/rally"
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
		cloud = kingpin.Arg(
			"cloud",
			"Name of the cloud from clouds.yaml",
		).Required().String()
		task = kingpin.Arg(
			"file",
			"Name of the task file",
		).Required().String()
	)

	kingpin.Version(version.Print("rally-exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	runner := rally.NewPeriodicRunner(*cloud, *task)
	prometheus.MustRegister(runner)

	go runner.Run()

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(`<html>
			<head><title>Rally Exporter</title></head>
			<body>
			<h1>Rally Exporter</h1>
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
