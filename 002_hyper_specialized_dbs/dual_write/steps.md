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

Generate license

``` sh
cockroach sql \
  --insecure \
  -e "SELECT crdb_internal.cluster_id()"

crl-lic -type "Evaluation" -org "Rob Test" -months 1 6756cfad-e805-4bef-b3d5-51fd0e8b41d2
```

Connect to CockroachDB

``` sh
cockroach sql --insecure
```

Apply license

``` sql
SET CLUSTER SETTING cluster.organization = 'Rob Test';
SET CLUSTER SETTING enterprise.license = 'crl-0-ChBnVs+t6AVL77PVUf0Oi0HSEL2X/6wGGAIiCFJvYiBUZXN0';
```

Get cert.pem base64

``` sh
base64 -i 002_hyper_specialized_dbs/dual_write/after/cert.pem | pbcopy
```

Create product changefeed

``` sql
SET CLUSTER SETTING kv.rangefeed.enabled = true;
SET CLUSTER SETTING changefeed.new_webhook_sink_enabled = true;

CREATE CHANGEFEED INTO 'webhook-https://host.docker.internal:3000/bigquery?insecure_tls_skip_verify=true&ca_cert=LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUY3ekNDQTllZ0F3SUJBZ0lVYko3RlZuTks5eXlBbzB0aWorT2lTWnZ2QU1jd0RRWUpLb1pJaHZjTkFRRUwKQlFBd2dZWXhDekFKQmdOVkJBWVRBbGhZTVJJd0VBWURWUVFJREFsVGRHRjBaVTVoYldVeEVUQVBCZ05WQkFjTQpDRU5wZEhsT1lXMWxNUlF3RWdZRFZRUUtEQXREYjIxd1lXNTVUbUZ0WlRFYk1Ca0dBMVVFQ3d3U1EyOXRjR0Z1CmVWTmxZM1JwYjI1T1lXMWxNUjB3R3dZRFZRUUREQlJEYjIxdGIyNU9ZVzFsVDNKSWIzTjBibUZ0WlRBZUZ3MHkKTXpFeU1URXhNVEl6TVRKYUZ3MHpNekV5TURneE1USXpNVEphTUlHR01Rc3dDUVlEVlFRR0V3SllXREVTTUJBRwpBMVVFQ0F3SlUzUmhkR1ZPWVcxbE1SRXdEd1lEVlFRSERBaERhWFI1VG1GdFpURVVNQklHQTFVRUNnd0xRMjl0CmNHRnVlVTVoYldVeEd6QVpCZ05WQkFzTUVrTnZiWEJoYm5sVFpXTjBhVzl1VG1GdFpURWRNQnNHQTFVRUF3d1UKUTI5dGJXOXVUbUZ0WlU5eVNHOXpkRzVoYldVd2dnSWlNQTBHQ1NxR1NJYjNEUUVCQVFVQUE0SUNEd0F3Z2dJSwpBb0lDQVFDdE4zdlBuR3hESFpjZk9rTHVOUGtiVWhQRmRhOVNTVFJBN1pmbHo4NHV1YW1sT3liaXhPZzk2SDRKCktnWm8yaTBRLzhHQkRuMnVCeU9JT2hMV1NaQmFCSDZDT2NVYUYzN05YdkxjV3FKUncwcng0QW1LVFhPc3ZPQ0UKNktFVGxrdC8rYm9KR0VIeG81Z0M2bmVuNStpbytlUC9RKyswMGxyRm1vTFFsanNKcWtpR0R0b3NRMFd2S1o0cAo5YzhVNGRlRVhSY2xzdk14eWc0L3QyZDZKRzFoMlhqd1BiOUdYL0J3MGxWRmRCeGQySk03aHRkRUx1QjFld3JpCndYTTZYNjlLTmZOVHM4bzFjRDlZQTVOZ1V3anNjNmwvSzR4NEp4QWtjUkx6d2krSFpxbmtHTEJFNnVmbGlGVU8KcUtEYzBXS1BsbFYzZFFaaG52MlRDZ2M1VXF1R0pXM09KaEhpWkJuQlZSdVZYTTdLVlVnWE5ZYW0xNi9PbzhNWgpZeHNiUVdEV1hsdU9KVG9qK2lkWjRKYi9Ga0Y1bitHZDZwUW1nNjZzbHY1VnptREk2d0t1TWdTOVQwb1Z0aVVuClN5dTl6YXdyRXZ3VG4wWnZCSFlqblI4ZGZPWThST1dHNW1HeVZHR0Q1QWpxekRiUFZiL3RzMjZ2ZE8yZ2F3dEsKaUxJWThoamt3djl4dHFMNUpZVUZieXV1MGRBT3l0NGI0QkYxRjNFTTdoLzFaOUE3N0NjOWtqZFQwYmRBNU1Hbgp5V3M3eGNvUW9sZmFGdElENzNmWHpCUWpQaDgxS2EyS0VvMU1VckFCTHdneWhjZnBKU0lJMTZPSGpab21IRitmCndBanRselBmZUp3bFhCeWQrVFRkWHJWcTNnampJck5XVlU1bDhQcGpMbWVka3p3V3ZRSURBUUFCbzFNd1VUQWQKQmdOVkhRNEVGZ1FVamZ5OEYxTDQ2c3J6bDE2YzhLUlRRcWxWODA0d0h3WURWUjBqQkJnd0ZvQVVqZnk4RjFMNAo2c3J6bDE2YzhLUlRRcWxWODA0d0R3WURWUjBUQVFIL0JBVXdBd0VCL3pBTkJna3Foa2lHOXcwQkFRc0ZBQU9DCkFnRUFlYzhlNHVlT2tqRWNDS2FYVFo1MGN1T0pkMFEzMzFTUEl0S0Jsa2FkRFBQOGlwNVN5SVNCdW9VbEZQRS8KNWE0UElOTTdKVDVlWnUzbFlsakcxZ3hTMm5DWUNRTWhwWlpHQVN6ek8zTm11cEVYTDdTTU5VZmpTeVNUTDVqQgpBTjVoNlFSeERpN0Z1eDdNbjVBTEVzWVcxZG5ZT1RwWWR0U0haYm53cEZsd0hWSU5FaTg1V3lTTktIcHUvSnJuCm9LTmNHUnlIWEsvd0JXcENQQkllQ0JzVm5LVGlyWXdMNTIxUkZpZUZZaDJGQXkxcXFkZ2pnVXFmd2VVVjhKNTAKczM0aFJJWEZvTTE3K3l5LzhBN21hR1pSclpaSksrc2JoR0VVc2U1b2g3S3RzRzJlaFZLdkN3K3g4dVVRZysxVwpxMGU1bjJ0Q1QxQlRNemQxVTZkZTJaQlFUR1FXa0FNZGk5SE51N2d1S1BtNXdMbHA2eGNTc3c1NGtBV096Q3poCk9RRjRHQVljZGRBT1JKZDFMeG5tZGtDb1FGcmRnWXlMQUlTWGNRcEtkM1ZrcThPdDArUzh6aElMWnFqaWlOOGIKQjYvTys5bngzNC9QSHZyam5wcEg5bkRQbGtnUzNUUStWeXJmRGU5ck5TS3IrSU9pWWRPeFJoOEtlY0IyazVlcApsMitCc0U0VG9MLzgwOW5CU0VacmVtMnRiK1NqQ1dCMTNHbzRuVGtYc2FDYUFOa2FMenM1NzdIak1BS3krMmFjCkhSbm1uVndJSkpTUDh5VTUyeHUyay9QRjdMd0djR2w5K3ZuakFpWmJsb1BDc1RUMnBkRUtVVE9RM0phaklEZ24KdklEcnBVNTNoWVFNUDRvMFZKWFN4b2lYUTQ1VnBHWnl0blR6ejN0WExkb2dsMHc9Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K'
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

### Teardown

``` sh
make teardown
```