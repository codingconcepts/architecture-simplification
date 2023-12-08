# Before

### Introduction

* 3 regions, all separate
* Duplicated code to cater for differences between regions
* Separate translations and supported languages in each region

### Infra

Services

``` sh
cp go.* 006_app_silos/multi_instance_architecture/before/services/eu
cp go.* 006_app_silos/multi_instance_architecture/before/services/jp
cp go.* 006_app_silos/multi_instance_architecture/before/services/us

(
  cd 006_app_silos/multi_instance_architecture/before && \
  docker compose up --build --force-recreate -d
)
```

### Run

Populate the databases

``` sh
psql "postgres://postgres:password@localhost:5432/postgres?sslmode=disable" \
  -f 006_app_silos/multi_instance_architecture/before/services/us/create.sql

psql "postgres://postgres:password@localhost:5433/postgres?sslmode=disable" \
  -f 006_app_silos/multi_instance_architecture/before/services/eu/create.sql

psql "postgres://postgres:password@localhost:5434/postgres?sslmode=disable" \
  -f 006_app_silos/multi_instance_architecture/before/services/jp/create.sql
```

Test the services

``` sh
curl -s "http://localhost:3001/products?lang=en" | jq
curl -s "http://localhost:3002/products?lang=es" | jq
curl -s "http://localhost:3003/products?lang=ja" | jq
```

### Summary

* No way of getting a holistic view of the business without a data warehousing solution
* Adding/updating a product means performing as many operations as there are regions
* Adding/updating a translation means performing as many operations as there are regions
* Data is duplicated everywhere
* Muliple databases and applications to maintain
* This wouldn't be acceptable in code (DRY), so why would it be acceptable in architecture?

# After

### Introduction

* 3 regions, 1 database
* Same code to cater for differences between regions
* Translations and supported languages shared by all regions

### Infra

Services

``` sh
cp go.* 006_app_silos/multi_instance_architecture/after/services/global

(
  cd 006_app_silos/multi_instance_architecture/after && \
  docker compose up --build --force-recreate -d
)
```

### Run

Populate the database

``` sh
cockroach init \
  --host localhost:26001 \
  --insecure
```

Generate license

``` sh
cockroach sql \
  --url "postgresql://root@localhost:26001/?sslmode=disable" \
  -e "SELECT crdb_internal.cluster_id()"

crl-lic -type "Evaluation" -org "Rob Test" -months 1 029c665d-e668-4f2a-8ff9-165a56c8b2cf
```

Apply license

``` sh
cockroach sql \
  --url "postgresql://root@localhost:26001/?sslmode=disable" \
  -e "SET CLUSTER SETTING cluster.organization = 'Rob Test'"

cockroach sql \
  --url "postgresql://root@localhost:26001/?sslmode=disable" \
  -e "SET CLUSTER SETTING enterprise.license = 'crl-0-ChACnGZd5mhPKo/5FlpWyLLPEKCT8KwGGAIiCFJvYiBUZXN0'"
```

Create tables

``` sh
cockroach sql \
  --url "postgres://root@localhost:26001/defaultdb?sslmode=disable" \
  < 006_app_silos/multi_instance_architecture/after/services/global/create.sql
```

Test the services

``` sh
curl -s "http://localhost:3001/products/uk?lang=en" | jq
curl -s "http://localhost:3002/products/us?lang=es" | jq
curl -s "http://localhost:3003/products/jp?lang=ja" | jq
```


# Teardown

``` sh
make teardown
```