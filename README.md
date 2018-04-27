# Kubetron

```
kubernetes + neutron = kubetron
```

Kubetron is a Kubernetes secondary networks plugin allowing users to connect
Pods to multiple networks created on Neutron (or other backend providing
Neutron-like API). Currently supports only OVN implementation.

PRs and issues are welcome.

## Development environment usage

Development environment provides multi-node setup of Kubernetes and
ovirt-ovn-provider (minimal implementation of Neutron).

```shell
# start master and nodes machines
vagrant up

# deploy kubernetes on master and two nodes
./hack/deploy-kubernetes

# check if nodes node1 and node2 were connected to master
./hack/kubectl get nodes

# deploy ovn on master (central) and nodes (host+controller)
./hack/deploy-ovn

# check if hosts are listed in southbound database
vagrant ssh master -c 'sudo ovn-sbctl show'

# deploy ovn provider on master
./hack/deploy-ovn-provider

# check that ovn provider works
vagrant ssh master -c 'curl 127.0.0.1:9696/v2.0/networks'

# verify ovirt-provider-ovn
./hack/verify-ovn-provider

# remove all machines
vagrant destroy
```

## Installation and usage

Check currenty implemented functionality of the plugin in development
environment.

```shell
# install plugin
./hack/install-addon

# or install with custom admission image
ADMISSION_IMAGE=user/image:version ./hack/install-addon

# in case admission image was changed, remove it and reinstall plugin
# (no need to call this if no changes were made since install)
./hack/kubectl delete ds admission --namespace kubetron
./hack/reinstall-addon

# check if admission is running and ready
./hack/kubectl get ds --namespace kubetron

# create two networks on neutron, red and blue
./hack/create-networks

# create a pod requesting networks red and blue
./hack/kubectl create -f deploy/example-kubetron-pod.yaml

# verify that networksSpec annotation was added
./hack/kubectl get pod example-kubetron-pod -o json | jq '.metadata.annotations'

# verify that sidecar requesting resource was added
./hack/kubectl get pod example-kubetron-pod -o json | jq '.spec.containers'

# verify that network ports were added
vagrant ssh master -c "curl http://localhost:9696/v2.0/ports"

# remove the pod
./hack/kubectl delete pod example-kubetron-pod

# verify that network ports were removed
vagrant ssh master -c "curl http://localhost:9696/v2.0/ports"
```

## Development

Some helpers for oblivious.

```shell
# refresh dependencies
dep ensure

# don't refresh, just download dependencies
dep ensure --vendor-only

# build admission locally
CGO_ENABLED=0 GOOS=linux go build cmd/admission/main.go

# build and push admission image
docker build -f cmd/admission/Dockerfile -t phoracek/kubetron-admission:latest .
docker push phoracek/kubetron-admission:latest
```

## TODO

- Device plugin was not implemented yet, that means it does not fully work yet.
- Design deck
- Currenty communicates with Neutron API in plaintext without any auth.
  Provide security configuration.
