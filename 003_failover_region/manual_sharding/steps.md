# Before

### Scenario

* Running a business in the EU
* Our only customers are EU customers, so we not considering any form of sharding

### Infrastructure

Start single UK node (simulating business start-up)

``` sh
docker run -d \
  --name eu_db \
  --platform linux/amd64 \
  -p 5432:5432 \
  -v eu_db:/var/lib/postgresql/data \
  -e POSTGRES_PASSWORD=password \
    postgres:16
```

Create tables

> No mention of country yet, as we've assumed everyone's in the UK.

```sh
psql "postgres://postgres:password@localhost:5432/?sslmode=disable" \
  -c 'CREATE TABLE customer (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        email TEXT NOT NULL
      );'

psql "postgres://postgres:password@localhost:5432/?sslmode=disable" \
  -c 'CREATE TABLE purchase (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        customer_id UUID NOT NULL REFERENCES customer(id),
        amount DECIMAL NOT NULL
      );'
```

### Run

Simulate customers

```sh
go run 003_failover_region/manual_sharding/before/main.go \
  --url "postgres://postgres:password@localhost:5432/?sslmode=disable"
```

### Expand UK

Bring up additional UK node

``` sh
docker run -d \
  --name eu_db_2 \
  --platform linux/amd64 \
  -p 5433:5432 \
  -v eu_db:/var/lib/postgresql/data \
  -e POSTGRES_PASSWORD=password \
    postgres:16
```

### Expand to US and JP

> At this point we gain US customers. Explain their experience.
> At this point we gain JP customers. Explain their experience.
> We decide to shard.
> Bring up two additional nodes (one in US, one in JP) - already done.



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
