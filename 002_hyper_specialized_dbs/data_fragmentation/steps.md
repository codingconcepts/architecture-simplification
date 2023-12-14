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

Connect to redis

``` sh
docker exec -it redis redis-cli
```

Enable keyspace notifications

``` sh
config set notify-keyspace-events KEA
```

Write data

``` sh
SET e4619022-7f6f-4292-a158-673c25d7ed37 '{"id": "e4619022-7f6f-4292-a158-673c25d7ed37", "name": "Latte", "description": "A milky coffee"}'

SET e8368c71-4786-484e-9101-da4396c5a411 '{"id": "e8368c71-4786-484e-9101-da4396c5a411", "name": "Cortado", "description": "A less milky coffee"}'
```

Connect to database

``` sh
cockroach sql --insecure
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