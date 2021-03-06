# Keep in sync with Dockerfile_debug!

# Build binary in golang container #

FROM golang:1.14 as builder

RUN mkdir /app
ADD . /app/
WORKDIR /app

# Build service
# note: -ldflags '-s -w' strips debugger info
RUN CGO_ENABLED=0 go build -a -ldflags '-s -w' -o apigee-remote-service-envoy .

# Get health probe
RUN GRPC_HEALTH_PROBE_VERSION=v0.3.1 && \
    wget -qO/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64 && \
    chmod +x /grpc_health_probe

# Build runtime container #

FROM scratch

# Add health probe
COPY --from=builder /grpc_health_probe /grpc_health_probe

# Add certs
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

# Add service
COPY --from=builder /app/apigee-remote-service-envoy .

# Run
ENTRYPOINT ["/apigee-remote-service-envoy"]
EXPOSE 50051
CMD ["--address=:50051", "--log_level=info"]
