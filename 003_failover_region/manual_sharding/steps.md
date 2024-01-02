# Before

### Scenario

* Running a business in the EU
* Our only customers are EU customers, so we not considering any form of sharding

### Create

Infrastructure

```sh
(
  cd 003_failover_region/manual_sharding/before && \
  docker compose up --build --force-recreate -d
)
```

Create tables

> No mention of country yet, as we've assumed everyone's in the UK.

```sh
psql "postgres://postgres:password@localhost:5432/?sslmode=disable" \
  -c 'CREATE TABLE customer (
        id UUID NOT NULL DEFAULT gen_random_uuid(),
        email TEXT NOT NULL
      );'
```

### Run

Server

```sh
go run 003_failover_region/manual_sharding/before/main.go \
  --url "postgres://postgres:password@localhost:5432/?sslmode=disable"
```

Simulate customers

``` sh
k6 run 003_failover_region/manual_sharding/before/load.js \
  --vus 10 \
  --duration 1h \
  --summary-trend-stats="min,max,p(95)"
```

> At this point we gain US customers. Explain their experience.
> At this point we gain JP customers. Explain their experience.
> We decide to shard.
> Bring up two additional nodes (one in US, one in JP) - already done.

Install citus

``` sh
(
  cd 003_failover_region/manual_sharding/before && \
  docker exec -i eu_db bash < install_citus.sh &\
  docker exec -i jp_db bash < install_citus.sh &\
  docker exec -i us_db bash < install_citus.sh
)

docker restart eu_db us_db jp_db
```

Install the citus extension

``` sh
psql "postgres://postgres:password@localhost:5432/?sslmode=disable" \
  -c "CREATE EXTENSION IF NOT EXISTS citus;"

psql "postgres://postgres:password@localhost:5433/?sslmode=disable" \
  -c "CREATE EXTENSION IF NOT EXISTS citus;"

psql "postgres://postgres:password@localhost:5434/?sslmode=disable" \
  -c "CREATE EXTENSION IF NOT EXISTS citus;"
```

Prepare the coordinator

``` sh
psql "postgres://postgres:password@localhost:5432/?sslmode=disable" \
  -c "SELECT citus_set_coordinator_host('eu_db', 5432);" \
  -c "SELECT * FROM citus_add_node('jp_db', 5432);" \
  -c "SELECT * FROM citus_add_node('us_db', 5432);"
```

Create tables and partition

``` sh
psql "postgres://postgres:password@localhost:5432/?sslmode=disable" \
  < 003_failover_region/manual_sharding/before/partition.sql
```

Populate the partitioned customer table

``` sh
psql "postgres://postgres:password@localhost:5432/?sslmode=disable" \
  < 003_failover_region/manual_sharding/before/populate.sql
```

Test partitioning

``` sql
INSERT INTO global.customer (country, email) VALUES
  ('de', 'uk_a@gmail.com'),
  ('fr', 'uk_a@gmail.com'),
  ('ei', 'uk_a@gmail.com'),
  ('uk', 'uk_a@gmail.com'),
  ('us', 'us_a@gmail.com'),
  ('mx', 'us_a@gmail.com'),
  ('ca', 'us_a@gmail.com'),
  ('br', 'us_a@gmail.com'),
  ('jp', 'jp_a@gmail.com'),
  ('zh', 'jp_a@gmail.com'),
  ('in', 'jp_a@gmail.com'),
  ('sg', 'jp_a@gmail.com');

SELECT * FROM eu.customer;
SELECT * FROM us.customer;
SELECT * FROM jp.customer;
```

# ISSUE - All data is on all nodes

Test direct inserts into workers

``` sh
psql "postgres://postgres:password@localhost:5433/?sslmode=disable" \
  -c "INSERT INTO jp.test (country, email) VALUES ('jp', 'jp_direct@gmail.com')"

psql "postgres://postgres:password@localhost:5434/?sslmode=disable" \
  -c "INSERT INTO us.test (country, email) VALUES ('us', 'us_direct@gmail.com')"

psql "postgres://postgres:password@localhost:5432/?sslmode=disable" \
  -c "INSERT INTO eu.test (country, email) VALUES ('de', 'de_direct@gmail.com')" \
  -c "INSERT INTO eu.test (country, email) VALUES ('fr', 'fr_direct@gmail.com')"
```

# After

# Teardown

```sh
make teardown
```
