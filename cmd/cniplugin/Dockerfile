FROM golang:1.9 AS multus_builder
WORKDIR /go/src/github.com/Intel-Corp/multus-cni/
RUN git clone https://github.com/Intel-Corp/multus-cni .
RUN CGO_ENABLED=0 GOOS=linux ./build

FROM golang:1.9 AS cniplugin_builder
WORKDIR /go/src/github.com/phoracek/kubetron/
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
COPY . .
RUN dep ensure --vendor-only
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/kubetron-cniplugin github.com/phoracek/kubetron/cmd/cniplugin

FROM alpine
COPY --from=multus_builder /go/src/github.com/Intel-Corp/multus-cni/bin/multus /opt/multus
COPY --from=cniplugin_builder /bin/kubetron-cniplugin /opt/kubetron
# ENTRYPOINT ["/bin/kubetron-admission"]

# TODO: build cni plugin

# TODO: build multus

# TODO: copy cni plugin to image (not to bin, will be installed to host)

# TODO: copy multus to to image (not to to bin, will be installed to host)
