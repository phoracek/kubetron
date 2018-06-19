# Kubetron

Kubetron is a Kubernetes secondary networks plugin allowing users to connect
Pods to multiple networks created on Neutron (or other backend providing
Neutron-like API). Currently supports only OVN implementation. Both overlay
and physical networks are supported.

PRs and issues are welcome.

Slide deck with desired kubetron model and demo: [Google Docs](https://docs.google.com/presentation/d/1KiHQyZngdaL8gtreL9Tmy7S1XiY5Sbnn0YuNCqhggF8/edit?usp=sharing)

1. [Development environment usage](#development-environment-usage)
2. [Installation and usage](#installation-and-usage)
3. [Demo](#demo)
4. [Development](#development)
5. [TODO](#todo)

## Development environment usage

Development environment provides multi-node setup of Kubernetes and
ovirt-ovn-provider (minimal implementation of Neutron).

```shell
# clone repository with its submodules
git clone https://github.com/phoracek/kubetron
cd kubetron

# install development environment dependencies
dnf install vagrant ansible kubernetes-client python-pip
pip install -r hack/kubespray/requirements.txt

# deploy cluster with kubernetes, ovn and ovirt-provider-ovn
./hack/deploy-cluster

# remove all machines (don't if you want to go through the next step)
./hack/destroy-cluster
```

## Installation and usage

Check currenty implemented functionality of the plugin in development
environment.

```shell
# install plugin
./hack/install-addon

# check if admission and deviceplugins are running and ready
./hack/kubectl get ds --namespace kubetron

# create two networks on neutron, red and blue, both of them have a subnet assigned
./hack/create-networks

# create two pods requesting networks red and blue
./hack/kubectl create -f deploy/example-kubetron-pods.yaml

# verify that networksSpec annotation was added
./hack/kubectl get pod example-kubetron-pod1 -o json | jq '.metadata.annotations'
./hack/kubectl get pod example-kubetron-pod2 -o json | jq '.metadata.annotations'

# verify that sidecar requesting resource was added
./hack/kubectl get pod example-kubetron-pod1 -o json | jq '.spec.containers[] | select(.name=="kubetron-request-sidecart")'
./hack/kubectl get pod example-kubetron-pod2 -o json | jq '.spec.containers[] | select(.name=="kubetron-request-sidecart")'

# verify that ports to all networks with a subnet are passed as arguments to sidecar
./hack/kubectl get pod example-kubetron-pod1 -o json | jq '.spec.containers[] | select(.name=="kubetron-request-sidecart") | .args'
./hack/kubectl get pod example-kubetron-pod2 -o json | jq '.spec.containers[] | select(.name=="kubetron-request-sidecart") | .args'

# verify that network ports were added
vagrant ssh master -c "curl http://localhost:9696/v2.0/ports" | jq

# check if pods are running and ready
./hack/kubectl get pod example-kubetron-pod1
./hack/kubectl get pod example-kubetron-pod2

# verify that pods obtained IP addresses from OVN DHCP server
./hack/kubectl exec -ti example-kubetron-pod1 -c example-container ip address
./hack/kubectl exec -ti example-kubetron-pod2 -c example-container ip address

# try to ping from one pod to another
./hack/kubectl exec -ti example-kubetron-pod1 -c example-container ping $BLUE_OR_RED_POD2_ADDRESS

# remove pods
./hack/kubectl delete -f deploy/example-kubetron-pods.yaml

# verify that network ports were removed
vagrant ssh master -c "curl http://localhost:9696/v2.0/ports" | jq
```

## Demo

[![asciicast](https://asciinema.org/a/7nB3vgIJcz05TxRNiaD2vLLdE.png)](https://asciinema.org/a/7nB3vgIJcz05TxRNiaD2vLLdE)

## Development

Some helpers for oblivious.

```shell
# refresh dependencies
dep ensure

# don't refresh, just download dependencies
dep ensure --vendor-only

# build admission binary locally
CGO_ENABLED=0 GOOS=linux go build cmd/admission/main.go

# build deviceplugin binary locally
CGO_ENABLED=0 GOOS=linux go build cmd/deviceplugin/main.go

# build and push admission image
docker build -f cmd/admission/Dockerfile -t phoracek/kubetron-admission:latest .
docker push phoracek/kubetron-admission:latest

# build and push deviceplugin image
docker build -f cmd/deviceplugin/Dockerfile -t phoracek/kubetron-deviceplugin:latest .
docker push phoracek/kubetron-deviceplugin:latest

# build and push sidecar image
docker build -f cmd/sidecar/Dockerfile -t phoracek/kubetron-sidecar:latest .
docker push phoracek/kubetron-sidecar:latest
```

## TODO

- Currenty communicates with Neutron API in plaintext without any auth.
  Provide security configuration.
- If possible, communicate with OVN NB, not Neutron (or support both).
- Make images smaller.
- Limit security only to needed.
- Add readiness check to sidecar.
