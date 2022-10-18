FROM golang:1.19 AS builder
WORKDIR /usr/src/polaris-scheduler

# Copy the Go Modules manifests
COPY go.work go.work.sum ./
COPY ./cluster-agent/go.mod ./cluster-agent/go.sum ./cluster-agent/
COPY ./framework/go.mod ./framework/go.sum ./framework/
COPY ./k8s-connector/go.mod ./k8s-connector/go.sum ./k8s-connector/
COPY ./scheduler/go.mod ./scheduler/go.sum ./scheduler/

# Cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the source code
COPY ./ ./

# Build the cluster-agent
RUN make build-cluster-agent


FROM alpine:3.16
WORKDIR /usr/apps/polaris-cluster-agent

# Copy the default config
COPY --from=builder /usr/src/polaris-scheduler/cluster-agent/manifests/default-polaris-cluster-agent-config.yaml /etc/polaris-cluster-agent/polaris-cluster-agent-config.yaml

# Copy the build output
COPY --from=builder /usr/src/polaris-scheduler/bin/polaris-cluster-agent ./

CMD [ "./polaris-cluster-agent", "--config", "/etc/polaris-cluster-agent/polaris-cluster-agent-config.yaml" ]
