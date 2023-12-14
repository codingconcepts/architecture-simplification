# Before

### Dependencies

* cqlsh CLI
* gcloud CLI

### Introduction

* Lots of data duplication
* Lots of application responsibility
* Multiple writes (easy for databases to fall out-of-sync)

### Infra

Databases

``` sh
(
  cd 002_hyper_specialized_dbs/dual_write/before && \
  docker compose up --build --force-recreate -d
)
```

Tables

``` sh
# Postgres
psql "postgres://postgres:password@localhost:5432/postgres?sslmode=disable" \
  -c "CREATE TABLE orders (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), user_id UUID NOT NULL, total DECIMAL NOT NULL, ts TIMESTAMP NOT NULL DEFAULT now())"

# Cassandra
cqlsh -e "CREATE KEYSPACE example WITH REPLICATION = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 };"
cqlsh -e "CREATE TABLE example.orders (id UUID PRIMARY KEY, user_id UUID, total DOUBLE, ts TIMESTAMP)"

# BigQuery
bq mk \
  --api http://localhost:9050 \
  --project_id local \
  example

bq mk \
  --api http://localhost:9050 \
  --project_id local \
  --table example.orders id:STRING,user_id:STRING,total:FLOAT,ts:TIMESTAMP
```

### Run

``` sh
go run 002_hyper_specialized_dbs/dual_write/before/main.go
```

### Check data

``` sh
# Postgres
psql "postgres://postgres:password@localhost:5432/postgres?sslmode=disable" \
  -c "SELECT * FROM orders"

# Cassandra
cqlsh -e "SELECT * from example.orders"

# BigQuery
bq query \
  --api http://localhost:9050 \
  --project_id local \
  "SELECT * FROM example.orders WHERE id IS NOT NULL"
```

# After

### Dependencies

* gcloud CLI

### Infra

Databases

``` sh
(
  cd 002_hyper_specialized_dbs/dual_write/after && \
  docker compose up --build --force-recreate -d
)
```

Tables

``` sh
# CockroachDB
cockroach sql --insecure -e "CREATE TABLE orders (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), user_id UUID NOT NULL, total DECIMAL NOT NULL, ts TIMESTAMP NOT NULL DEFAULT now())"

# BigQuery
bq mk \
  --api http://localhost:9050 \
  --project_id local \
  example

bq mk \
  --api http://localhost:9050 \
  --project_id local \
  --table example.orders id:STRING,user_id:STRING,total:FLOAT,ts:TIMESTAMP
```

### Run

Start server (with certificates)

``` sh
(
  cd 002_hyper_specialized_dbs/dual_write/after && \
  openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -sha256 -days 3650 -nodes -subj "/C=XX/ST=StateName/L=CityName/O=CompanyName/OU=CompanySectionName/CN=CommonNameOrHostname" \
)

(cd 002_hyper_specialized_dbs/dual_write/after && go run main.go)
```

Convert to enterprise

``` sh
enterprise --url "postgres://root@localhost:26257/?sslmode=disable"
```

Get cert.pem base64

``` sh
base64 -i 002_hyper_specialized_dbs/dual_write/after/cert.pem | pbcopy
```

Connect to CockroachDB

``` sh
cockroach sql --insecure
```

Create product changefeed

``` sql
SET CLUSTER SETTING kv.rangefeed.enabled = true;
SET CLUSTER SETTING changefeed.new_webhook_sink_enabled = true;

CREATE CHANGEFEED INTO 'webhook-https://host.docker.internal:3000/bigquery?insecure_tls_skip_verify=true&ca_cert=LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUY3ekNDQTllZ0F3SUJBZ0lVTkQxVW5jdzBaNFRTSURYTnJ3M1piUUhVVDFrd0RRWUpLb1pJaHZjTkFRRUwKQlFBd2dZWXhDekFKQmdOVkJBWVRBbGhZTVJJd0VBWURWUVFJREFsVGRHRjBaVTVoYldVeEVUQVBCZ05WQkFjTQpDRU5wZEhsT1lXMWxNUlF3RWdZRFZRUUtEQXREYjIxd1lXNTVUbUZ0WlRFYk1Ca0dBMVVFQ3d3U1EyOXRjR0Z1CmVWTmxZM1JwYjI1T1lXMWxNUjB3R3dZRFZRUUREQlJEYjIxdGIyNU9ZVzFsVDNKSWIzTjBibUZ0WlRBZUZ3MHkKTXpFeU1UUXhOREU1TWpsYUZ3MHpNekV5TVRFeE5ERTVNamxhTUlHR01Rc3dDUVlEVlFRR0V3SllXREVTTUJBRwpBMVVFQ0F3SlUzUmhkR1ZPWVcxbE1SRXdEd1lEVlFRSERBaERhWFI1VG1GdFpURVVNQklHQTFVRUNnd0xRMjl0CmNHRnVlVTVoYldVeEd6QVpCZ05WQkFzTUVrTnZiWEJoYm5sVFpXTjBhVzl1VG1GdFpURWRNQnNHQTFVRUF3d1UKUTI5dGJXOXVUbUZ0WlU5eVNHOXpkRzVoYldVd2dnSWlNQTBHQ1NxR1NJYjNEUUVCQVFVQUE0SUNEd0F3Z2dJSwpBb0lDQVFDZGJVdkg2RzVPOGFua04wSHZLRHNGOUlDZDRqM0IxRG9oVjB5MkFuWjd0Z2FaV3dDSTY1dCtDaDlzCmlYRUZTV1dNOUZMbG1Sb0RFLzlZOHhkOTB1OHVZMHp6eEFGeFNoVzVpOTk2SzhCUitOWUdydDljbitEQXBDWSsKTldCQklkSUVoOUltOHFSeXNFeTlKc2JQTDNCSVh0cGVWV1VUNGNod0VJNU9YV0pnMFlWcXV1RzV1NjRFdlBKeApQQkJwdysvR3I2dzV6N2NKWVFTY1JwMmREazYrYVdZL3dFUVBXa0hpdXdFZVdPbHpCRWJUalc2OEJ3TFhta1Q4CmNGMzJRbkVqMlp4OEphTDBxMFJLcDJONVlFOExEVEVOMnpzandvR1lwbkJMS1RlM2JuS2hrc3dONDh6ellNSkcKRjliMkpnVnF6S2NQZDBJVUpvaFlaVUdORlVOa2dXblQ2Y1JoUlhTekhacVNMSzczK2I4cXVVaXN2Zjl5RERERApDT1Y0VHBHeEJMYUpKK3ZBT01zYmMyeStCaVQ0Z05OSFVQTlRNeHl0dzJxTjc0WW8xYWhhSkl1eUlXa2RjTkhmCnk5NnExczQ5WlVRdG9NeDR6MHVZUmVrWUhJbk9DOVZRdzI3anFBMjFEdVV0bVdYS1VpYWl1d0VwZWk2YXJqVXMKdVltV2VyNnFuaGdBUU1IdUk0bUU2TVE1QnZvdmMvOUV1SXcyQTdTcU5JbG5FZXNqQ3hBYWZ1NVZnMFVOdjhjSQoxdHhmdFcxMG4zS21pcnRsc2FTOCtEaEx3UXhKU2pXV3NsTmxiNGNna05FVWNUWDhJVUZoNmxBQmwrOW1WbGZMCnRPM013cFdGWFg2cEhnOGxpV2NEY2VmYWM1K1NjQjVpU3EzU082K2pXOTF2MXp4VHFRSURBUUFCbzFNd1VUQWQKQmdOVkhRNEVGZ1FVd3lKWXFocU95d3IwNVdzV21SVUM4eElDV0ZRd0h3WURWUjBqQkJnd0ZvQVV3eUpZcWhxTwp5d3IwNVdzV21SVUM4eElDV0ZRd0R3WURWUjBUQVFIL0JBVXdBd0VCL3pBTkJna3Foa2lHOXcwQkFRc0ZBQU9DCkFnRUFGRU9hZWNCL2JMOGFvSnY2bGJRQmxCODBzVGd0YjhaY2M4ZmIxclhlUU1Gd1pDS21jK1pSSVpMeloxamQKRERXczVmcVFpSzdJazRTbzVvampqM3FDRlN6R3Nyc0ZsbFN5ejNlZTNmV3MxWXVEV0FmQkMwNkJyanVBSnFyMQpVREhDeEU3dUNsTVFyVlZ3ZUxSWlRvOUlpSHkrb1FCZ2NNcEw1OWNkcEhyTG5mZUtrMEhkUkRQTE5YOFhpRWJ2CmdVdFdMaGswMDhkUStaaWVZckdQbU5PWDdRWUQ0VFRKejd1WjRnQXFVMG9RMG9RL0pva294TitzUXUwVmNaVm8KRW44bmg0SWVaL2Y4MEFydHJJcmpXbHhtYXRlaXJwSjQycmxXNDNmbEYyNEo5QXdNM0h1OTk3RFhxRktOSVIvOQpWWU9UWGxCdTBWMzdTenI3TVJuRk9hTUpGWUpMeWYrcWtsZ0pNVk44VjB4VzJ6UnE0YUFiM1l6MmtIcUJYV0o4ClJsMFltckg3ZFNNVXRFcWpNaVhLYXF1U2d4QVNKclhFUlRzZTU1Nzd5c1BJYjFMSnR0VWJFN2FVMG5zcUw2bjIKRW1IWERuc0dva1BwY1pKem14a0xLVzg1RGFVQTY2SDdHMlREdno1VDg2L0dXOUJGNnN1NXV6ckltVXdsOGFYVApGbkV4U2s4N0l4RlBWTHl3cmJ1MmxOMXBlZCtURGlpT3ozRHM1NjNaaVFBcmRIM3JMN2h0cUlRUTRoclJXMGZjCjB6UkV3WXNvMmhIbGU5SURYV21qU2toWmdqaEhIZHlZankxa3g0REhMYTVkWnFWT1A5MjZiZ2Y5cVZXYXhtOEUKY0NNM1ZWaXU3eFQ2M0dRQXRwMGlZcGpUdEJtL2lpT1dPWG9HZzRYTGhUeldHdEU9Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K'
AS SELECT
  "id",
  "user_id",
  "total",
  "ts"::TIMESTAMPTZ
FROM orders;
```

### Check data

``` sh
# CockroachDB
cockroach sql --insecure -e "SELECT count(*) FROM orders"

# BigQuery
bq query \
  --api http://localhost:9050 \
  --project_id local \
  "SELECT * FROM example.orders WHERE id IS NOT NULL"

bq query \
  --api http://localhost:9050 \
  --project_id local \
  "SELECT count(*) count FROM example.orders WHERE id IS NOT NULL"
```

### Down down BigQuery to show everything continues (and catches up)

**NOT WORKING YET** Try creating a volume.

``` sh
docker stop bigquery

docker start bigquery

bq query \
  --api http://localhost:9050 \
  --project_id local \
  "SELECT count(*) count FROM example.orders WHERE id IS NOT NULL"
```

### Teardown

``` sh
make teardown
```