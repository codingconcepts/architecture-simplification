# Before

**2 terminal windows (vertical)**

### Introduction

* With Cassandra (and by virtue, Scylla), data is modelled around queries:
  * If you have varying query requirements, you might need to duplicate data to achieve that
  * This is why we're using an indexer database (in this case, Postgres)

### Infra

``` sh
cp go.* 002_hyper_specialized_dbs/data_fragmentation/before/services/indexer

(
  export DEBEZIUM_VERSION=2.5.0.CR1 && \
  cd 002_hyper_specialized_dbs/data_fragmentation/before && \
  docker-compose -f compose.yaml up --build -d
)
```

Kafka topic and consumer

``` sh
kafkactl create topic products.store.product
kafkactl consume products.store.product
```

### Run

Create index table

``` sh
psql "postgres://postgres:password@localhost/?sslmode=disable" \
  -c 'CREATE TABLE product (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        name TEXT NOT NULL,
        description TEXT NOT NULL,
        ts TIMESTAMPTZ NOT NULL DEFAULT now()
      );'
```

Run local indexer for testing

``` sh
KAFKA_URL="localhost:9092" \
INDEX_URL="postgres://postgres:password@localhost/?sslmode=disable" \
  go run 002_hyper_specialized_dbs/data_fragmentation/before/services/indexer/main.go
```

Create keyspace and table (wait for a short while before attempting to connect)

``` sh
clear && cqlsh
```

Create keyspace and table

``` sql
CREATE KEYSPACE IF NOT EXISTS store
  WITH REPLICATION = {
    'class' : 'SimpleStrategy', 'replication_factor': 1
  }
  AND durable_writes = true;

USE store;

CREATE TABLE product (
  id uuid,
  , name text
  , description text
  , ts timeuuid
  , PRIMARY KEY (id, ts)
) WITH cdc=true;
```

Generate load

``` sh
go run 002_hyper_specialized_dbs/data_fragmentation/before/services/load/main.go
```

Watch index for updates

``` sh
see psql "postgres://postgres:password@localhost/?sslmode=disable" \
  -c 'SELECT COUNT(*) FROM product;'
```

Insert data

``` sql
INSERT INTO product (id, name, description, ts)
VALUES (a975c293-ca78-437e-b098-ef13f93f3e88, 'Latte', 'A mikly coffee', 1097ff70-abd3-11ee-911f-3d6bc11da4eb);

INSERT INTO product (id, name, description, ts)
VALUES (b1be93bd-95fd-4f78-859d-520434793fd9, 'Cortado', 'A less mikly coffee', 109a4960-abd3-11ee-911f-3d6bc11da4eb);
```

Update data

``` sql
INSERT INTO product (id, name, description, ts)
VALUES (cba441ec-a841-41ca-a684-69aba7aa34f6, 'Flat White', 'A less mikly coffee', 10c466a0-abd3-11ee-911f-3d6bc11da4eb);

UPDATE product
SET description = 'A much less mikly coffee'
WHERE id = b1be93bd-95fd-4f78-859d-520434793fd9
AND ts = 10c466a0-abd3-11ee-911f-3d6bc11da4eb;
```

Check for updates

``` sql
SELECT * FROM product;
```

### Summary

* Some of our customers have performed this exact migration:
  * Their query requirements outgrew the database they were writing to
  * ...which necessitated a read-specialized database

* With CockroachDB, you can scale for both reads and writes:
  * Meaning one database
  * Less infrastructure
  * Less to manage
  * ...and less to go wrong
