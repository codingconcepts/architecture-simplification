# Before

**2 terminal windows (vertical)**

### Infra

``` sh
cp go.* 002_hyper_specialized_dbs/data_fragmentation/before/services/indexer

(
  cd 002_hyper_specialized_dbs/data_fragmentation/before && \
  docker compose up --build --force-recreate -d
)
```

### Run

Create index table

``` sh
psql "postgres://postgres:password@localhost/?sslmode=disable" \
  -c 'CREATE TABLE product (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        name TEXT NOT NULL,
        description TEXT NOT NULL
      );'
```

Create keyspace and table (wait for a short while before attempting to connect)

``` sh
cqlsh
```

Create keyspace and table

``` sql
CREATE KEYSPACE IF NOT EXISTS store
  WITH REPLICATION = {
    'class' : 'SimpleStrategy', 'replication_factor': 1
  }
  AND DURABLE_WRITES = true;

USE store;

CREATE TABLE product (
  id uuid PRIMARY KEY,
  , name text
  , description text
)
WITH cdc = {
  'enabled': 'true',
  'postimage': 'true'
};
```

Insert data

``` sql
INSERT INTO product (id, name, description)
VALUES (a975c293-ca78-437e-b098-ef13f93f3e88, 'Latte', 'A mikly coffee');

INSERT INTO product (id, name, description)
VALUES (b1be93bd-95fd-4f78-859d-520434793fd9, 'Cortado', 'A less mikly coffee');

INSERT INTO product (id, name, description)
VALUES (cba441ec-a841-41ca-a684-69aba7aa34f6, 'Flat White', 'A less mikly coffee');
```

Watch index for updates

``` sh
see psql "postgres://postgres:password@localhost/?sslmode=disable" \
  -c 'SELECT * FROM product;'
```

Update data

``` sql
UPDATE product
SET description = 'A much less mikly coffee'
WHERE id = b1be93bd-95fd-4f78-859d-520434793fd9;
```

Check for updates

``` sql
SELECT * FROM product;
```

### Summary

* Redis is a very capable data store, I just needed to demonstrate _some_ ingestion database.

# After

* Just the database

### Infra

``` sh

```