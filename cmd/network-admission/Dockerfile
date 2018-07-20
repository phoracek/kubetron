FROM golang:1.9 AS builder
WORKDIR /go/src/github.com/phoracek/kubetron/
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/kubetron-network-admission github.com/phoracek/kubetron/cmd/network-admission

FROM alpine
COPY --from=builder /bin/kubetron-network-admission /bin/kubetron-network-admission
ENTRYPOINT ["/bin/kubetron-network-admission"]
