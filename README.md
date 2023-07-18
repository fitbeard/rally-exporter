# OpenStack Rally Exporter for Prometheus

Originally forked from: <https://opendev.org/vexxhost/rally_exporter>

## Latest Docker image

```sh
docker pull quay.io/tadas/rally-exporter:latest
```

## Release Docker image

```sh
docker pull quay.io/tadas/rally-exporter:0.1
```

### Description

The OpenStack Rally exporter, exports Prometheus metrics from a running OpenStack Rally test suite
for consumption by Prometheus. This version is compatible with Rally deployments [which has preconfigured users](https://docs.openstack.org/rally/latest/quick_start/tutorial/step_3_benchmarking_with_existing_users.html#registering-deployment-with-existing-users-in-rally).

This exporter can be run:

```sh
docker run \
-v "$PWD/examples/deployment.yml":/conf/deployment.yml \
-v "$PWD/examples/tasks.yml":/conf/tasks.yml \
-v "$PWD/examples/args.yml":/conf/arguments.yml \
-it -p 9355:9355 quay.io/tadas/rally-exporter:latest --deployment-name cloud1
```

Paths to configuration files are hardcoded:

* Deployment - /conf/deployment.yml
* Tasks - /conf/tasks.yml
* Task arguments - /conf/arguments.yml

### Command line options

The current list of command line options (by running --help)

```sh
usage: rally-exporter --deployment-name=DEPLOYMENT-NAME [<flags>]

Flags:
  -h, --help              Show context-sensitive help (also try --help-long and --help-man).
      --web.listen-address=":9355"
                          Address on which to expose metrics and web interface.
      --web.telemetry-path="/metrics"
                          Path under which to expose metrics.
      --deployment-name=DEPLOYMENT-NAME
                          Name of the Rally deployment
      --execution-time=5  Wait X minutes before next run. Default: 5
      --task-history=10   Number of tasks to keep in history. Default: 10
      --version           Show application version.
```

## Example metrics

```sh
# HELP rally_task_duration Rally task duration
# TYPE rally_task_duration gauge
rally_task_duration{title="glance"} 1.014307975769043
rally_task_duration{title="keystone"} 1.9640920162200928
# HELP rally_task_passed Rally task passed
# TYPE rally_task_passed gauge
rally_task_passed{title="glance"} 1
rally_task_passed{title="keystone"} 1
# HELP rally_task_time Rally last run time
# TYPE rally_task_time gauge
rally_task_time 1.68969289e+09
```
