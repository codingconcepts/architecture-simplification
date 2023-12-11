# Before

### Dependencies

* cqlsh CLI
* gcloud CLI

### Introduction

* Lots of data duplication
* Lots of application responsibility
* Multiple writes (easy for databases to fall out-of-sync)

### Infra

Databases

``` sh
(
  cd 002_hyper_specialized_dbs/dual_write/before && \
  docker compose up --build --force-recreate -d
)
```

Tables

``` sh
# Postgres
psql "postgres://postgres:password@localhost:5432/postgres?sslmode=disable" \
  -c "CREATE TABLE orders (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), user_id UUID NOT NULL, total DECIMAL NOT NULL, ts TIMESTAMP NOT NULL DEFAULT now())"

# Cassandra
cqlsh -e "CREATE KEYSPACE example WITH REPLICATION = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 };"
cqlsh -e "CREATE TABLE example.orders (id UUID PRIMARY KEY, user_id UUID, total DOUBLE, ts TIMESTAMP)"

# BigQuery
bq mk \
  --api http://localhost:9050 \
  --project_id local \
  example

bq mk \
  --api http://localhost:9050 \
  --project_id local \
  --table example.orders id:STRING,user_id:STRING,total:FLOAT,ts:TIMESTAMP
```

### Run

``` sh
go run 002_hyper_specialized_dbs/dual_write/before/main.go
```

### Check data

``` sh
# Postgres
psql "postgres://postgres:password@localhost:5432/postgres?sslmode=disable" \
  -c "SELECT * FROM orders"

# Cassandra
cqlsh -e "SELECT * from example.orders"

# BigQuery
bq query \
  --api http://localhost:9050 \
  --project_id local \
  "SELECT * FROM example.orders WHERE id IS NOT NULL"

```

### Teardown

``` sh
make teardown
```