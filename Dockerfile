FROM golang:1.20-bullseye AS builder

WORKDIR /go/src/app
COPY . .
RUN go build

# FROM xrally/xrally-openstack:2.2.0
FROM xrally/xrally-openstack:latest

COPY delete-tasks.sh /delete-tasks.sh
RUN chmod +x /delete-tasks.sh
COPY --from=builder /go/src/app/rally-exporter /rally-exporter
ENTRYPOINT ["/rally-exporter"]
