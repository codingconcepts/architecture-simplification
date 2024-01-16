# Before

**2 terminal windows**

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
  -c "CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    total DECIMAL NOT NULL,
    ts TIMESTAMP NOT NULL DEFAULT now()
  )"

# BigQuery
bq mk \
  --api http://localhost:9050 \
  --project_id local \
  example

bq mk \
  --api http://localhost:9050 \
  --project_id local \
  --table example.orders id:STRING,user_id:STRING,total:FLOAT,ts:TIMESTAMP

# Cassandra
cqlsh -e "CREATE KEYSPACE example
  WITH REPLICATION = {
    'class' : 'SimpleStrategy',
    'replication_factor' : 1
  };"

cqlsh -e "CREATE TABLE example.orders (
  id UUID PRIMARY KEY,
  user_id UUID,
  total DOUBLE,
  ts TIMESTAMP
)"
```

### Check data

``` sh
go run 002_hyper_specialized_dbs/dual_write/eod/main.go \
  --postgres "postgres://postgres:password@localhost:5432/postgres?sslmode=disable" \
  --cassandra "localhost:9042" \
  --bigquery "http://localhost:9050"
```

### Run

``` sh
go run 002_hyper_specialized_dbs/dual_write/before/main.go
```

# After

**3 terminal windows**

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
cockroach sql --insecure -e "CREATE TABLE orders (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL,
  total DECIMAL NOT NULL,
  ts TIMESTAMP NOT NULL DEFAULT now()
)"

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

### Check data

``` sh
go run 002_hyper_specialized_dbs/dual_write/eod/main.go \
  --postgres "postgres://root@localhost:26257/defaultdb?sslmode=disable" \
  --bigquery "http://localhost:9050"
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

CREATE CHANGEFEED INTO 'webhook-https://host.docker.internal:3000/bigquery?insecure_tls_skip_verify=true&ca_cert=LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUY3ekNDQTllZ0F3SUJBZ0lVTmVEVityNmM1ank4c29nS0FTQ3RtZnlpMCtzd0RRWUpLb1pJaHZjTkFRRUwKQlFBd2dZWXhDekFKQmdOVkJBWVRBbGhZTVJJd0VBWURWUVFJREFsVGRHRjBaVTVoYldVeEVUQVBCZ05WQkFjTQpDRU5wZEhsT1lXMWxNUlF3RWdZRFZRUUtEQXREYjIxd1lXNTVUbUZ0WlRFYk1Ca0dBMVVFQ3d3U1EyOXRjR0Z1CmVWTmxZM1JwYjI1T1lXMWxNUjB3R3dZRFZRUUREQlJEYjIxdGIyNU9ZVzFsVDNKSWIzTjBibUZ0WlRBZUZ3MHkKTkRBeE1UVXhNRE0zTVRCYUZ3MHpOREF4TVRJeE1ETTNNVEJhTUlHR01Rc3dDUVlEVlFRR0V3SllXREVTTUJBRwpBMVVFQ0F3SlUzUmhkR1ZPWVcxbE1SRXdEd1lEVlFRSERBaERhWFI1VG1GdFpURVVNQklHQTFVRUNnd0xRMjl0CmNHRnVlVTVoYldVeEd6QVpCZ05WQkFzTUVrTnZiWEJoYm5sVFpXTjBhVzl1VG1GdFpURWRNQnNHQTFVRUF3d1UKUTI5dGJXOXVUbUZ0WlU5eVNHOXpkRzVoYldVd2dnSWlNQTBHQ1NxR1NJYjNEUUVCQVFVQUE0SUNEd0F3Z2dJSwpBb0lDQVFDempoU25JUStTQnFPNVRBWkliaGtNV2lTVHlWUkF6Z3lCMmducGtVVDN3bktLWW0xakl4ekUwU2hQCi81WUl4eUo5OW5MY1JsZ3k3aGRJa0hIZVVRYXdwMHVsYVl5MDY0MytlN21adHJFWis2M0JWL1lqenR0TkJ1WG4KaklWT1RHc1Vsb0wzYTVvN3pObWlkbkVZNjVVZU04R29qaldzSzRwRC9QQnRyUXd5bml3S251MDlyZU9tNVNnWApnRWlBbDd3a2FKeFZVNnpSL1JvcWhPRHJQVE9pcEd6TFlCS01DeXpkSVV4QmdLaUk2czl3WUJZNDRIbHV2Q3ZICjhMaDRDaHRpelVydkhWWXUzdXBHVEsrMVk2VnFmOGFXVnlzcGl5ajBIQlNaVmJVSVFDUnJKNkU5MjUvcDNOT2YKblh3SDF0a3BSNEFNUk1VSG5RSnJ1czBFZjZ1ZUY3ZENuTHNEb0dKMVQxT0haU3lFZFVVMXNIVU53aUVsSm95NQpTdXhrTFlGa0hXVkRqRnV2cCtGM1JuMWlXZTZudG9wdUlCNmg2aUhBMHVwVnVzT050aFFkVk43Z29EaXhkQzBYClF1ZWx0T0dESCtPVFJrMHFpYmtCaWRFSlNMaTJadGdNcmVNUVpsa2VyQTIxT1BaM09QNVluUW9wZVZnZENScTQKcTkveDNaK3dveVhUNUlzM3o0aWY0Ulp4YnF0a2l6dXVKM2Zkcld4a3J1OXRCVEhwYjRvOExMdlBTbE5QT3NzdQp5WFVJVFJqSCtjc2JqV0ZHV1g0ckIrRVl0V3JFZTR4VGdTQS9VV2QyVVI4ak1DblVVNG9lREFvUWtiaWkzL3EyCkwvUGNtNXdaNUZmREdyYXNURkFFZmNxZjRrWDJQUUpMOVBOcllibXZSU0t1UVAzeHdRSURBUUFCbzFNd1VUQWQKQmdOVkhRNEVGZ1FVSlpLZnJMVmVRQ09VN1htZFNVQ0lCNWV4RUhVd0h3WURWUjBqQkJnd0ZvQVVKWktmckxWZQpRQ09VN1htZFNVQ0lCNWV4RUhVd0R3WURWUjBUQVFIL0JBVXdBd0VCL3pBTkJna3Foa2lHOXcwQkFRc0ZBQU9DCkFnRUFKckZhWk45blNPZ0ppdEpmTkpCSXF5MUk5OVZMeHFOcTVhTHNzVWFqeG1ZeUJTZEdYZVBqcW5Wcm91NlcKUzZ5TC9KUW5WSS8vQjFXcXZCZHRjODh4bDFDejZkK1ZtMHJ4ZStlM0xOK1I4YlFLUVJadE1Qa1ZzNXhMUFJpNApZS0tydGwwTWtTSElKb3VUZGlhVC9GMUJ6TTdlMjNzbkp2U3dycGlCaUw3aGVxdEVlUUpzVi9qNVpwWkhibmFsCndqbDdtcHFTV0ZqV3ZteGs3SHd0MitJbmQ4OFgxSEpoV3RWY1Q1QUxFU1VTdjJTa3MyRnllaWh2b1JXeUYzbkkKYWlMTS9oWDZYRndmaUhHM1BzYTJlODNSR0lCbGs3QStJaFczc1FTYW5MRzYrUEhvTjgzZUIyYmErSVJjdWkvMwovMGhpd2RnTnBWRDJZUytZUXQ5OXhiRHA2N1BJakY3c3BZZmRWbmFzcTQ1T3VvcWVJMkg4RlVrZW83TnltSnVlCkpEMzkwbTMzeDBtK2gvY0hRVlRWTjJaNzFucENUeE42VU04NWo5SDdWRUkxMUhUYmx2TXpFQ3VTRytHWVpFaWkKTWlpZkQ4SHlZeTIzWTFBOHhaWkJSczdqcU1sdnZTNUlsOFMzSjZFSmoxd2R6eVpMcCtwTWc0TGFlVlkvY3BYUApqL0RZVlpUcmI1RUFIM0VDbEo5SDBHYWFXcllVZzl5a01jSlFsbWpFbG44SDFUTnpRQ256T0pqVkQ4em1WZVAwCmNZL1ZsMUU1L0FGejlhSkNGWCtDSDlzZ09xVlFENWdvZUJMcnRLQ1J6WVhuRVVlZE9YTHhQSU5TNXg3bVhib04KNWxoS3JUYU43NEoxOUpyTG0rbEJseXNTTVYwd3VFU2M4amhDbnNTdWpYUVF6SkU9Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K#'
AS SELECT
  "id",
  "user_id",
  "total",
  "ts"::TIMESTAMPTZ
FROM orders;
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