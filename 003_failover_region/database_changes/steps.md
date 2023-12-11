# Before

### Introduction

* In order to do this without downtime, you'd have to run some kind of blue/green, failover setup.
* I cover that in another demo, so will just perform the

### Infra

Kubernetes cluster

``` sh
k3d registry create local-registry --port 9090

k3d cluster create local \
  --registry-use k3d-local-registry:9090 \
  --registry-config 003_failover_region/database_changes/registries.yaml \
  --k3s-arg "--disable=traefik,metrics-server@server:*;agents:*" \
  --k3s-arg "--disable=servicelb@server:*" \
  --wait
```