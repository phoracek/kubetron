FROM golang:1.9 AS builder
WORKDIR /go/src/github.com/phoracek/kubetron/
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/kubetron-pod-admission github.com/phoracek/kubetron/cmd/pod-admission

FROM alpine
COPY --from=builder /bin/kubetron-pod-admission /bin/kubetron-pod-admission
ENTRYPOINT ["/bin/kubetron-pod-admission"]
