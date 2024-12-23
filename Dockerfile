# Build the manager binary
FROM golang:1.23.4 as builder

WORKDIR /go/src/github.com/gardener/kupid
# Copy the Go Modules manifests
COPY go.mod /go/src/github.com/gardener/kupid/go.mod
COPY go.sum /go/src/github.com/gardener/kupid/go.sum

# Copy the go source
COPY main.go /go/src/github.com/gardener/kupid/main.go
COPY api/ /go/src/github.com/gardener/kupid/api/
COPY pkg/ /go/src/github.com/gardener/kupid/pkg/
COPY vendor/ /go/src/github.com/gardener/kupid/vendor/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -v -mod=vendor -o kupid main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /go/src/github.com/gardener/kupid/kupid .
USER nonroot:nonroot

ENTRYPOINT ["/kupid"]
