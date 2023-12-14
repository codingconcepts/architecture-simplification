# Before

### Infra

Start Postgres (no replication, because we don't know we need it yet)

``` sh
docker run \
  -d \
  --name postgres_primary \
  -p 5432:5432 \
  -e POSTGRES_USER=user \
  -e POSTGRES_DB=postgres \
  -e POSTGRES_PASSWORD=password \
    postgres:15.2-alpine
```

Connect

``` sh
psql postgres://user:password@localhost:5432/postgres 
```

Create table and insert data

``` sql
CREATE TABLE purchase (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  basket_id UUID NOT NULL,
  member_id UUID NOT NULL,
  amount DECIMAL NOT NULL,
  timestamp TIMESTAMP NOT NULL DEFAULT now()
);
```

Run application

``` sh
CONNECTION_STRING="postgres://user:password@localhost:5432/postgres?sslmode=disable" \
  go run 003_failover_region/database_migration/before/main.go
```

### Migration

**NOTE**: At this point, we've realised we need to migration, so need to update Postgres to enable replication.

Connect to the shell

``` sh
psql postgres://user:password@localhost:5432/postgres 
```

Commands

``` sql
CREATE ROLE replica_user WITH REPLICATION LOGIN PASSWORD 'password';
GRANT ALL PRIVILEGES ON DATABASE postgres TO replica_user;

ALTER SYSTEM SET wal_level = 'logical';
```

Restart postgres

``` sh
docker restart postgres_primary
```

Connect to the shell

``` sh
psql postgres://user:password@localhost:5432/postgres 
```

Commands

``` sql
CREATE PUBLICATION replication_pub FOR TABLE purchase;
-- OR
SELECT pg_create_physical_replication_slot('replication_slot');
```

Grant replica access

``` sh
docker exec -it postgres_primary bash

cd /var/lib/postgresql/data
vi pg_hba.conf

echo "host    replication     replica_user      0.0.0.0/               trust" >> pg_hba.conf
```

Bring up replica

``` sh
docker run \
  -d \
  --name postgres_replica \
  -p 5433:5432 \
  -e POSTGRES_USER=user \
  -e POSTGRES_DB=postgres \
  -e POSTGRES_PASSWORD=password \
    postgres:15.2-alpine
```

<!-- Connect to the shell

``` sh
psql postgres://user:password@localhost:5433/postgres 
```

Commands

``` sql
CREATE TABLE purchase (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  basket_id UUID NOT NULL,
  member_id UUID NOT NULL,
  amount DECIMAL NOT NULL,
  timestamp TIMESTAMP NOT NULL DEFAULT now()
);

CREATE SUBSCRIPTION replication_sub
CONNECTION 'host=host.docker.internal dbname=postgres user=replica_user password=password application_name=replication'
PUBLICATION replication_pub;
``` -->

Connect to the container

``` sh
docker exec -it postgres_replica bash

pg_basebackup \
  --pgdata=/var/lib/postgresql/data \
  -R \
  --slot=replication_slot \
  --host=host.docker.internal \
  --port=5432 \
  -U replica_user \
  -W
```




pg_basebackup --pgdata=/var/lib/postgresql/data -R --slot=replication_slot --host=host.docker.internal --port=5432