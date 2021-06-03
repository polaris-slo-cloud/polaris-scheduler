# Builds a Docker image with the k8s-test-cluster which can be used with Fogify.

FROM python:3.8 AS builder

# Compile the cpu-limit tool for a fully functional Fogify
WORKDIR /src
RUN git clone https://github.com/opsengine/cpulimit/
WORKDIR /src/cpulimit
RUN make


FROM kindest/node:v1.20.2

# Copy the RAINBOW k8s-test-cluster setup files
COPY ./rainbow /rainbow

# Copy the cpu-limit tool for Fogify
COPY --from=builder /src/cpulimit/src/cpulimit /usr/bin

# Install the stress tool for Fogify and for debugging and make sure that the scripts are executable
RUN apt-get update && \
    apt-get install -y stress iputils-ping && \
    apt-get clean && \
    chmod +x /rainbow/rainbow-entrypoint.sh && \
    chmod +x /rainbow/control-plane-node/rainbow-control-node-setup.sh && \
    chmod +x /rainbow/worker-node/rainbow-worker-node-setup.sh

ENTRYPOINT ["/rainbow/rainbow-entrypoint.sh"]
