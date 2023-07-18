FROM golang:1.13-stretch AS builder
WORKDIR /go/src/app
COPY . .
RUN go build

FROM xrally/xrally-openstack:1.7.0
COPY --from=builder /go/src/app/rally-exporter /rally-exporter
ENTRYPOINT ["/rally-exporter"]
