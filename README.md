```shell
# start master and nodes machines
vagrant up

# deploy kubernetes on master and two nodes
./deploy-kubernetes

# check if nodes node1 and node2 were connected to master
./kubectl get nodes

# deploy ovn on master (central) and nodes (host+controller)
./deploy-ovn

# check if hosts are listed in southbound database
vagrant ssh master -c 'sudo ovn-sbctl show'

# deploy ovn provider on master
./deploy-ovn-provider

# check that ovn provider works
vagrant ssh master -c 'curl 127.0.0.1:9696/v2.0/networks'

# remove all machines
vagrant destroy
```
