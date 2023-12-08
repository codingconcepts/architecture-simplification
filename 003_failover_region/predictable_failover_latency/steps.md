# Before

**5 console windows**

### Infra

``` sh
(
  cd 003_failover_region/predictable_failover_latency/before && \
  docker compose up --build --force-recreate -d
)
```

### Run

Connect to the primary node

``` sh
psql postgres://user:password@localhost:5432/postgres 
```

Create table and insert data

``` sql
CREATE TABLE product (
  "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  "name" VARCHAR(255) NOT NULL,
  "price" DECIMAL NOT NULL
);

INSERT INTO product ("name", "price") VALUES
  ('a', 0.99),
  ('b', 1.99),
  ('c', 2.99),
  ('d', 3.99),
  ('e', 4.99);
```

Connect to the secondary node

``` sh
psql postgres://user:password@localhost:5433/postgres 
```

Query table

``` sql
SELECT count(*) FROM product;
```

Spin up load balancer

``` sh
dp \
  --server "localhost:5432" \
  --server "localhost:5433" \
  --port 5430
```

Run application

``` sh
CONNECTION_STRING=postgres://user:password@localhost:5430/postgres?sslmode=disable \
  go run 003_failover_region/predictable_failover_latency/main.go
```

Take down primary

``` sh
docker stop primary
```

Promote replica

``` sh
docker exec -it replica bash

pg_ctl promote
```

Switch load balancer to point to replica (new primary)

### Summary

* The failover to the replica was successfully. Now what?
  * How do you get back to the primary?
  * Does the primary now become the replica?
  * How much data was lost during the outage and how to we backfill?

# After

### Infra

``` sh
(
  cd 003_failover_region/predictable_failover_latency/after && \
  docker compose up --build --force-recreate -d
)
```

### Run

Initialise the cluster

``` sh
docker exec -it secondary cockroach init --insecure
docker exec -it secondary cockroach sql --insecure 
```

Get cluster id

``` sql
SELECT crdb_internal.cluster_id();
```

Generate license

``` sh
crl-lic -type "Evaluation" -org "Rob Test" -months 1 b85654e0-5c89-47f3-8219-a5eb5b62c8dd
```

Apply license

``` sql
SET CLUSTER SETTING cluster.organization = 'Rob Test';
SET CLUSTER SETTING enterprise.license = 'crl-0-ChC4VlTgXIlH84IZpetbYsjdEMuE66wGGAIiCFJvYiBUZXN0';
```

Create table and insert data

``` sql
CREATE DATABASE store
  PRIMARY REGION "us-east-1"
  REGIONS "us-central-1", "us-west-2";

ALTER DATABASE store SET SECONDARY REGION = "us-central-1";

SHOW REGIONS FROM DATABASE store;

USE store;

CREATE TABLE product (
  "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  "name" STRING NOT NULL,
  "price" DECIMAL NOT NULL
);

INSERT INTO product ("id", "name", "price") VALUES
  ('a4aebc20-0355-40fa-86f7-b2ba25907cf2', 'a', 0.99),
  ('ba7a5891-8d82-46f3-8232-00aa7813392b', 'b', 1.99),
  ('cd5069b7-d399-4d7f-a733-e96ff31671c9', 'c', 2.99),
  ('dd9e1e42-81a8-454f-afae-c5fb9fac27f3', 'd', 3.99),
  ('ec7c7142-4bbc-418a-99c5-1fe621d0aca4', 'e', 4.99);
```

Run application

``` sh
CONNECTION_STRING=postgres://root@localhost:26257/store?sslmode=disable \
  go run 003_failover_region/predictable_failover_latency/main.go
```

View row locality

``` sh
SELECT DISTINCT
  split_part(unnest(replica_localities), ',', 1) replica_localities,
  unnest(replicas) replica,
  range_id
FROM [SHOW RANGE FROM TABLE product FOR ROW ('eu-west1', 'b7eaba08-2d20-4109-9a39-aeba20c486a4')];
```

Take down primary

``` sh
docker stop primary
```

### Teardown

``` sh
make teardown
```