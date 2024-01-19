### Pre-scale

Create first node

``` sh
docker run -d \
  --name eu_db_1 \
  --platform linux/amd64 \
  -p 5432:5432 \
  -v eu_db_1:/var/lib/postgresql/data \
  -e POSTGRES_PASSWORD=password \
    postgres:16

dw "postgres://postgres:password@localhost:5432/?sslmode=disable"
psql "postgres://postgres:password@localhost:5432/?sslmode=disable"
```

Create table

``` sql
CREATE TABLE customer (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email VARCHAR(255) NOT NULL
);

INSERT INTO customer (id, email)
  SELECT
    gen_random_uuid(),
    CONCAT(gen_random_uuid(), '@gmail.com')
  FROM generate_series(1, 1000);
```

### Scale-up 1

Enable the FDW extension.

``` sql
CREATE EXTENSION IF NOT EXISTS postgres_fdw;
```

Allow access to the foreign servers.

``` sh
docker exec -it eu_db_1 bash

# Mention that this is super unsafe.
echo "host all all 0.0.0.0/0 md5" >> /var/lib/postgresql/data/pg_hba.conf

exit
```

Restart eu_db_1 for changes to take effect.

``` sh
docker restart eu_db_1
dw "postgres://postgres:password@localhost:5432/?sslmode=disable"
psql "postgres://postgres:password@localhost:5432/?sslmode=disable"
```

Create second node

``` sh
docker run -d \
  --name eu_db_2 \
  --platform linux/amd64 \
  -p 5433:5432 \
  -v eu_db_2:/var/lib/postgresql/data \
  -e POSTGRES_PASSWORD=password \
    postgres:16

dw "postgres://postgres:password@localhost:5433/?sslmode=disable"
psql "postgres://postgres:password@localhost:5433/?sslmode=disable"
```

Add the FDW extension on eu_db_2.

``` sql
CREATE EXTENSION IF NOT EXISTS postgres_fdw;
```

Create the customer table on eu_db_2.

``` sql
CREATE TABLE customer (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email VARCHAR(255) NOT NULL
);
```

Back on eu_db_1, create a foreign server that points to the new server.

``` sql
CREATE SERVER eu_db_2 FOREIGN DATA WRAPPER postgres_fdw
  OPTIONS (
    host 'host.docker.internal',
    port '5433',
    dbname 'postgres'
  );
```

Map a local user to the foreign user and grant access to the FDW.

``` sql
CREATE USER MAPPING FOR postgres
  SERVER eu_db_2
  OPTIONS (
    user 'postgres',
    password 'password'
  );

GRANT USAGE ON FOREIGN SERVER eu_db_2 TO postgres;
```

Partition table.

``` sql
-- Create partitioned version of customer table.
CREATE TABLE customer_partitioned
  (LIKE customer)
  PARTITION BY HASH (id);

CREATE TABLE customer_0 PARTITION OF customer_partitioned
  FOR VALUES WITH (MODULUS 2, REMAINDER 0);

CREATE FOREIGN TABLE customer_1
  PARTITION OF customer_partitioned
  FOR VALUES WITH (MODULUS 2, REMAINDER 1)
  SERVER eu_db_2
  OPTIONS (
    table_name 'customer'
  );

-- Insert data into the partitioned table.
INSERT INTO customer_partitioned
SELECT * FROM customer;

-- Drop original table and replace with partitioned.
DROP TABLE customer;
ALTER TABLE customer_partitioned RENAME TO customer;

-- Test.
INSERT INTO customer (id, email)
  SELECT
    gen_random_uuid(),
    CONCAT(gen_random_uuid(), '@gmail.com')
  FROM generate_series(1, 1000);

-- Check.
SELECT COUNT(*) FROM customer;
SELECT COUNT(*) FROM customer_0;
SELECT COUNT(*) FROM customer_1;

SELECT
  table_name,
  pg_size_pretty(pg_total_relation_size(quote_ident(table_name))),
  pg_relation_size(quote_ident(table_name))
FROM information_schema.tables
WHERE table_schema = 'public'
ORDER BY table_name;
```

Check row count on eu_db_2.

``` sh
psql "postgres://postgres:password@localhost:5433/?sslmode=disable" \
  -c "SELECT COUNT(*) FROM customer";
```

### Scale-up 2

Create third node

``` sh
docker run -d \
  --name eu_db_3 \
  --platform linux/amd64 \
  -p 5434:5432 \
  -v eu_db_3:/var/lib/postgresql/data \
  -e POSTGRES_PASSWORD=password \
    postgres:16

dw "postgres://postgres:password@localhost:5434/?sslmode=disable"
psql "postgres://postgres:password@localhost:5434/?sslmode=disable"
```

Add the FDW extension on eu_db_3.

``` sql
CREATE EXTENSION IF NOT EXISTS postgres_fdw;
```

Create the customer table on eu_db_3.

``` sql
CREATE TABLE customer (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email VARCHAR(255) NOT NULL
);
```

Back on eu_db_1, create a foreign server that points to the new server.

``` sql
CREATE SERVER eu_db_3 FOREIGN DATA WRAPPER postgres_fdw
  OPTIONS (
    host 'host.docker.internal',
    port '5434',
    dbname 'postgres'
  );
```

Map a local user to the foreign user and grant access to the FDW.

``` sql
CREATE USER MAPPING FOR postgres
  SERVER eu_db_3
  OPTIONS (
    user 'postgres',
    password 'password'
  );

GRANT USAGE ON FOREIGN SERVER eu_db_3 TO postgres;
```

Create new table on eu_db_2.

``` sh
psql "postgres://postgres:password@localhost:5433/?sslmode=disable" \
  -c "CREATE TABLE customer_new (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        email VARCHAR(255) NOT NULL
      );"
```

Partition table.

``` sql
CREATE TABLE customer_partitioned
  (LIKE customer)
  PARTITION BY HASH (id);

CREATE TABLE customer_0_new PARTITION OF customer_partitioned
  FOR VALUES WITH (MODULUS 3, REMAINDER 0);

CREATE FOREIGN TABLE customer_1_new
  PARTITION OF customer_partitioned
  FOR VALUES WITH (MODULUS 3, REMAINDER 1)
  SERVER eu_db_2
  OPTIONS (
    table_name 'customer_new'
  );

CREATE FOREIGN TABLE customer_2_new
  PARTITION OF customer_partitioned
  FOR VALUES WITH (MODULUS 3, REMAINDER 2)
  SERVER eu_db_3
  OPTIONS (
    table_name 'customer'
  );

-- Insert data into the partitioned table.
INSERT INTO customer_partitioned
  SELECT * FROM customer;

DROP TABLE customer;
ALTER TABLE customer_partitioned RENAME TO customer;
ALTER TABLE customer_0_new RENAME TO customer_0;
ALTER TABLE customer_1_new RENAME TO customer_1;
ALTER TABLE customer_2_new RENAME TO customer_2;

-- Test.
INSERT INTO customer (id, email)
  SELECT
    gen_random_uuid(),
    CONCAT(gen_random_uuid(), '@gmail.com')
  FROM generate_series(1, 1000);

-- Check.
SELECT COUNT(*) FROM customer;
SELECT COUNT(*) FROM customer_0;
SELECT COUNT(*) FROM customer_1;
SELECT COUNT(*) FROM customer_2;

SELECT
  table_name,
  pg_size_pretty(pg_total_relation_size(quote_ident(table_name))),
  pg_relation_size(quote_ident(table_name))
FROM information_schema.tables
WHERE table_schema = 'public'
ORDER BY table_name;
```

Check row count on eu_db_2 and eu_db_3.

``` sh
psql "postgres://postgres:password@localhost:5433/?sslmode=disable" \
  -c "SELECT COUNT(*) FROM customer_new";

psql "postgres://postgres:password@localhost:5434/?sslmode=disable" \
  -c "SELECT COUNT(*) FROM customer";
```

Drop original customer table on eu_db_2 and rename customer_new.

``` sh
psql "postgres://postgres:password@localhost:5433/?sslmode=disable" \
  -c "DROP TABLE customer";

psql "postgres://postgres:password@localhost:5433/?sslmode=disable" \
  -c "ALTER TABLE customer_new RENAME TO customer;";

psql "postgres://postgres:password@localhost:5433/?sslmode=disable" \
  -c "ALTER FOREIGN TABLE customer_1
        OPTIONS (
          SET table_name 'customer'
        );"
```

### Teardown

``` sh
make teardown
```