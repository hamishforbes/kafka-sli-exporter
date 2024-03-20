# Build the manager binary
FROM golang:1.22-alpine as builder

ARG TARGETARCH
ARG CGO_ENABLED
ARG GO_LDFLAGS

ENV CGO_ENABLED=$CGO_ENABLED
ENV GOOS=linux
ENV GOARCH=$TARGETARCH

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY main.go main.go
COPY config.yaml config.yaml
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer

# Copy the go source
COPY cmd/ cmd/
COPY config/ config/
COPY pkg/ pkg/
#COPY vendor/ vendor/

# Build
RUN go build -ldflags $GO_LDFLAGS -a -o kafka-sli-exporter
RUN echo 'nonroot:x:1000:2000::/home/nonroot:/dev/null' > /tmp/passwd

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM scratch
WORKDIR /
COPY --from=builder /tmp/passwd /etc/passwd
COPY --from=builder /workspace/kafka-sli-exporter /kafka-sli-exporter
COPY --from=builder /workspace/config.yaml /config.yaml

USER 1000

ENTRYPOINT ["/kafka-sli-exporter","app"]
