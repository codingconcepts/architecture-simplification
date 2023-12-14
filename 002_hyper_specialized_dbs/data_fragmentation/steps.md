# Before

### Infra

``` sh
cp go.* 002_hyper_specialized_dbs/data_fragmentation/before/services/indexer

(
  cd 002_hyper_specialized_dbs/data_fragmentation/before && \
  docker compose up --build --force-recreate -d
)
```

### Run

Connect to database

``` sh
psql "postgres://postgres:password@localhost:5432/postgres"
```

``` sql
CREATE TABLE "product" (
  "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  "name" VARCHAR(255) NOT NULL,
  "description" VARCHAR(255)
);
```

Connect to redis

``` sh
docker exec -it redis redis-cli
```

Enable keyspace notifications

``` sh
config set notify-keyspace-events s
```

# After

### Infra

``` sh

```