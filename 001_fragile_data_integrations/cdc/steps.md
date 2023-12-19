# Before

### Infra

Containers

``` sh
(cd 001_fragile_data_integrations/cdc/before && docker compose up -d)
```

Kafka topic and consumer

``` sh
kafkactl consume events.public.payment
```

Table

``` sh
psql "postgres://postgres:password@localhost/?sslmode=disable" \
  -c 'CREATE TABLE payment (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        amount DECIMAL NOT NULL,
        ts TIMESTAMPTZ NOT NULL DEFAULT now()
      );'
```

Debezium connector

``` sh
curl "localhost:8083/connectors" \
  -H 'Content-Type: application/json' \
  -d '{
        "name": "db-connector",
        "config": {
          "connector.class": "io.debezium.connector.postgresql.PostgresConnector",
          "database.hostname": "postgres",
          "database.port": "5432",
          "database.user": "postgres",
          "database.password": "password",
          "database.dbname" : "postgres",
          "topic.prefix": "events",
          "tasks.max": 1,
          "decimal.handling.mode": "string",
          "include.schema.changes": "false"
        }
      }'
```

### Run

Listen for changes

``` sh
kafkactl consume events.public.payment
```

Run application

``` sh
go run 001_fragile_data_integrations/cdc/main.go \
  --url "postgres://postgres:password@localhost/?sslmode=disable"
```

# After

### Infra

``` sh

```

# Cleanup

Delete Debezium connector

``` sh
curl -X DELETE "localhost:8083/connectors/db-connector"

psql "postgres://postgres:password@localhost/?sslmode=disable" \
  -c "select pg_drop_replication_slot('debezium');"
```

Delete replication slot

