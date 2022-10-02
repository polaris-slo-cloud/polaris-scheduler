FROM golang:1.19 AS builder
WORKDIR /usr/src/polaris-scheduler

# Copy the Go Modules manifests
COPY go.work go.work.sum ./
COPY ./cluster-broker/go.mod ./cluster-broker/go.sum ./cluster-broker/
COPY ./framework/go.mod ./framework/go.sum ./framework/
COPY ./k8s-connector/go.mod ./k8s-connector/go.sum ./k8s-connector/
COPY ./scheduler/go.mod ./scheduler/go.sum ./scheduler/

# Cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the source code
COPY ./ ./

# Build the scheduler
RUN make build-scheduler


FROM alpine:3.16
WORKDIR /usr/apps/polaris-scheduler

# Copy the default config
COPY --from=builder /usr/src/polaris-scheduler/scheduler/manifests/polaris-scheduler/default-polaris-scheduler-config.yaml ./polaris-scheduler-config.yaml

# Copy the build output
COPY --from=builder /usr/src/polaris-scheduler/bin/polaris-scheduler ./

CMD [ "./polaris-scheduler", "--config", "./polaris-scheduler-config.yaml" ]
