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
go run 003_failover_region/predictable_failover_latency/before/main.go
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

### Teardown

``` sh
make teardown
```