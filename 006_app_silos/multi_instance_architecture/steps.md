# Before

### Introduction

* 3 regions, all separate
* Duplicated code to cater for differences between regions
* Separate translations and supported languages in each region
* Whenever you're running a global business, it's expensive.

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
curl -s "http://localhost:3003/products" | jq
```

### Summary

* No way of getting a holistic view of the business without a data warehousing solution
  * Unless all products, ids and SKUs remain consistent, there's no single consistent definition of a single product across all locations
* Adding/updating a product means performing as many operations as there are regions
* Adding/updating a translation means performing as many operations as there are regions
* Changes to the database, requires separate downtime for each region
* Data, code, infrastructure, and effort are duplicated everywhere
* High opex costs.
* Enforcing global constraints/rules (business or techincal) across regions, this is very hard.
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

Initialize the database

``` sh
cockroach init --host localhost:26001 --insecure
```

Convert to enterprise

``` sh
enterprise --url "postgres://root@localhost:26001/?sslmode=disable"
```

Create tables

``` sh
cockroach sql \
  --url "postgres://root@localhost:26001/defaultdb?sslmode=disable" \
  < 006_app_silos/multi_instance_architecture/after/services/global/create.sql
```

Observe data localities

``` sql
SET CLUSTER SETTING sql.show_ranges_deprecated_behavior.enabled = false;

SELECT DISTINCT
  split_part(unnest(replica_localities), ',', 1) replica_localities,
  unnest(replicas) replica,
  lease_holder,
  range_id
FROM [SHOW RANGE FROM TABLE product_markets FOR ROW ('eu-central-1', 'a50b1ae0-455d-4308-8d2f-ae17eeafd4b1', 'de')];

SELECT DISTINCT
  split_part(unnest(replica_localities), ',', 1) replica_localities,
  unnest(replicas) replica,
  lease_holder,
  range_id
FROM [SHOW RANGE FROM TABLE product_markets FOR ROW ('us-east-1', 'a50b1ae0-455d-4308-8d2f-ae17eeafd4b1', 'mx')];

SELECT DISTINCT
  split_part(unnest(replica_localities), ',', 1) replica_localities,
  unnest(replicas) replica,
  lease_holder,
  range_id
FROM [SHOW RANGE FROM TABLE product_markets FOR ROW ('ap-northeast-1', 'a50b1ae0-455d-4308-8d2f-ae17eeafd4b1', 'jp')];
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