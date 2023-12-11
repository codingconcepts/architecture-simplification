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

Deploy MySQL

``` sh
docker pull mysql:8.1.0
docker tag mysql:8.1.0 localhost:9090/mysql:8.1.0
docker push localhost:9090/mysql:8.1.0

kubectl apply -f 003_failover_region/database_changes/before/manifests/mysql/pv.yaml
kubectl apply -f 003_failover_region/database_changes/before/manifests/mysql/v8.1.0.yaml
```

Connect to MySQL

``` sh
kubectl run --rm -it mysqlshell --image=k3d-local-registry:9090/mysql:8.1.0 -- mysqlsh root:password@mysql --sql
```

Create tables

``` sql
CREATE DATABASE defaultdb;
USE defaultdb;

CREATE TABLE purchase (
  id VARCHAR(36) DEFAULT (uuid()) PRIMARY KEY,
  basket_id VARCHAR(36) NOT NULL,
  member_id VARCHAR(36) NOT NULL,
  amount DECIMAL NOT NULL,
  timestamp TIMESTAMP NOT NULL DEFAULT now()
);
```

Deploy application

``` sh
cp go.* 003_failover_region/database_changes/before
(cd 003_failover_region/database_changes/before && docker build -t app .)
docker tag app:latest localhost:9090/app:latest
docker push localhost:9090/app:latest
kubectl apply -f 003_failover_region/database_changes/before/manifests/app/deployment.yaml
```

Monitor application

``` sh
kubetail app
```

Update MySQL

``` sh
docker pull mysql:8.2.0
docker tag mysql:8.2.0 localhost:9090/mysql:8.2.0
docker push localhost:9090/mysql:8.2.0

kubectl apply -f 003_failover_region/database_changes/before/manifests/mysql/v8.2.0.yaml
```