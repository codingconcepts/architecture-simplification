### After

**3 terminal windows**

Create UK cluster

``` sh
cockroach start \
  --insecure \
  --store=path=node1 \
  --locality=region=eu-west-2 \
  --listen-addr=localhost:26257 \
  --http-addr=localhost:8080 \
  --join='localhost:26257,localhost:26258,localhost:26259' \
  --background

cockroach start \
  --insecure \
  --store=path=node2 \
  --locality=region=eu-west-2  \
  --listen-addr=localhost:26258 \
  --http-addr=localhost:8081 \
  --join='localhost:26257,localhost:26258,localhost:26259' \
  --background

cockroach start \
  --insecure \
  --store=path=node3 \
  --locality=region=eu-west-2  \
  --listen-addr=localhost:26259 \
  --http-addr=localhost:8082 \
  --join='localhost:26257,localhost:26258,localhost:26259' \
  --background

cockroach init --host localhost:26257 --insecure
cockroach sql --insecure
```

Create table

``` sql
CREATE TABLE customer (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email TEXT NOT NULL
);

INSERT INTO customer (id, email)
  SELECT
    gen_random_uuid(),
    CONCAT(gen_random_uuid()::STRING, '@gmail.com')
  FROM generate_series(1, 1000);
```

Start client

``` sh
go run 003_failover_region/horizontal_scaling/client.go \
  --url "postgres://root@localhost:26257/?sslmode=disable"
```

### Year 5 scale-up (multi-region)

``` sh
cockroach start \
  --insecure \
  --store=path=node4 \
  --locality=region=us-east-1 \
  --listen-addr=localhost:26260 \
  --http-addr=localhost:8083 \
  --join='localhost:26257,localhost:26258,localhost:26259' \
  --background

cockroach start \
  --insecure \
  --store=path=node5 \
  --locality=region=us-east-1  \
  --listen-addr=localhost:26261 \
  --http-addr=localhost:8084 \
  --join='localhost:26257,localhost:26258,localhost:26259' \
  --background

cockroach start \
  --insecure \
  --store=path=node6 \
  --locality=region=us-east-1  \
  --listen-addr=localhost:26262 \
  --http-addr=localhost:8085 \
  --join='localhost:26257,localhost:26258,localhost:26259' \
  --background
```

Enable enterprise (for geo-partitioning)

``` sh
enterprise --url "postgres://root@localhost:26257/?sslmode=disable"
```

All region column

> Notice how the customer queries aren't affected.

``` sql
ALTER TABLE customer
ADD REGION CRDB_INTERNAL_REGION NOT NULL DEFAULT 'eu-west-2';

ALTER DATABASE defaultdb
SET PRIMARY REGION 'eu-west-2';

ALTER DATABASE defaultdb
ADD REGION 'us-east-1';

ALTER TABLE customer
SET LOCALITY REGIONAL BY ROW;
```

Add data for both UK and US customers

``` sql
INSERT INTO customer (id, email, region)
  SELECT
    gen_random_uuid(),
    CONCAT(gen_random_uuid()::TEXT, '@gmail.com'),
    (ARRAY['eu-west-2', 'us-east-1'])[1 + floor((random() * 2))::int]
  FROM generate_series(1, 1000);
```