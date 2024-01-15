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

CREATE CHANGEFEED INTO 'webhook-https://host.docker.internal:3000/bigquery?insecure_tls_skip_verify=true&ca_cert=LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUY3ekNDQTllZ0F3SUJBZ0lVUVFBbU9mYnB2T29kN2F2WWVSWHc4ZjlRZGZzd0RRWUpLb1pJaHZjTkFRRUwKQlFBd2dZWXhDekFKQmdOVkJBWVRBbGhZTVJJd0VBWURWUVFJREFsVGRHRjBaVTVoYldVeEVUQVBCZ05WQkFjTQpDRU5wZEhsT1lXMWxNUlF3RWdZRFZRUUtEQXREYjIxd1lXNTVUbUZ0WlRFYk1Ca0dBMVVFQ3d3U1EyOXRjR0Z1CmVWTmxZM1JwYjI1T1lXMWxNUjB3R3dZRFZRUUREQlJEYjIxdGIyNU9ZVzFsVDNKSWIzTjBibUZ0WlRBZUZ3MHkKTkRBeE1URXhOVEk1TWpKYUZ3MHpOREF4TURneE5USTVNakphTUlHR01Rc3dDUVlEVlFRR0V3SllXREVTTUJBRwpBMVVFQ0F3SlUzUmhkR1ZPWVcxbE1SRXdEd1lEVlFRSERBaERhWFI1VG1GdFpURVVNQklHQTFVRUNnd0xRMjl0CmNHRnVlVTVoYldVeEd6QVpCZ05WQkFzTUVrTnZiWEJoYm5sVFpXTjBhVzl1VG1GdFpURWRNQnNHQTFVRUF3d1UKUTI5dGJXOXVUbUZ0WlU5eVNHOXpkRzVoYldVd2dnSWlNQTBHQ1NxR1NJYjNEUUVCQVFVQUE0SUNEd0F3Z2dJSwpBb0lDQVFDWEh4b05SazJzcGdMakFDZHRYbjhGa0Uzd0t2c1hacE50RTJFQ2pNZWJ3REQ4WmtUd3hYLzVVMFlYCmN2RkRYckhwWEt6SHR4ajRTMHp4SzRLZGUrT1hsTDRMd09sSk41Rjl3cExKSGZubGJWV3pFUE5WN2pCcXJhcDgKb2FXb2E5cDdZSzIzSWZRVU1MU0FSdUlsSWxyRXNmSThOZ3BnZDB2MFRFdGlocE1sa0V2ZXhZWnlFViszQUxobApVV0xBak1vUzY2L1ZtMlJ6ck13RWU5emFhbGpSd0FwTm41dzhMWVVYWldyQlBLVWtHbE90ZWptYjVVUHhZemljClpZb0k0cm1tUzFpbzlMblNYUlVYeFI2VEkrbFoyR05HaTdpMzF6bjFkUUQwdmYwLzJkTVFvZTk2UzA5aU5Kd1kKNy9Hd1M5M3JLUWpZaEVZV0lGME1DcTZSdnBHUEIrYTBKQkpzTFNla0VpWFBrTC9kSGxzU2pGVCtBcTIyTTlESwpnQmZNS1F6UnJZYXFVNU9QS0R0SWY3ME5tdWhlMEtEWi9lazhtMFRSQVJwK3VxQkVCUkdjL08vYUVpNkdGbzVMCkpzb1V6K1UwZmNXV3k4UTgrMXpLanFoRTE1Y2EvRUgybmREZkV0WklQOGRFWUdLN0hOVEJKUk5BcWN0bVg3WGoKeDJDd3o4bXBQOUFiNUJIdkk4MlRCZ2lhMll0a1VZYjdmUlRpdFNiTzdkQktxWXAzd2UwbFcxdm01dXk1SFdpVQplR0JETjltbGRkaVN0QnFwa0h0SmRxVzM3aUswYVhEdnZydDQzN003TzlBMFRyM1NiZWk1STRtYSs3WXdpajBOCnpOMWYrMG9sZExmZEpQUWpDS2RpSFhDaTFWcmNTUWR0WU55elNZSHQxUjJ0TmZRVjRRSURBUUFCbzFNd1VUQWQKQmdOVkhRNEVGZ1FVSzVqNjB1aTdTMHlBZEhDUTEvMVM1R3ZpbHNrd0h3WURWUjBqQkJnd0ZvQVVLNWo2MHVpNwpTMHlBZEhDUTEvMVM1R3ZpbHNrd0R3WURWUjBUQVFIL0JBVXdBd0VCL3pBTkJna3Foa2lHOXcwQkFRc0ZBQU9DCkFnRUFWd1FQenlmWHhxczB4R2J5Y0NsMEM5QUswVWJqT1Fwd3lOSHUzQmIzUlV6RTZ1dGdBM25sUWhHcDcvdHgKU0ZibGNCMUplRUlQL1c2VzQrMVdPeW5uUkErRnF1U1ZSMzFSaVhMMzdQYmFLVHFTYVBJcVJjY2xDSVdNVHArYQpPZUZGZGJtZHEyWHQ5bUJsMEwxMW0xUmtLVndRd1RqcGtpcXI2MDlaNDhaTHBZMjFLZ1ZxM012OC8rWWFHMmFlCk5SYjJQTG9wSnFhUTBOQi9YOU8xVjlQRmFybloraEp5Zkh0azVGRlp2SnlTWUtRa3hGbGU2YUtibzdBeXhvYlAKemNyZ091c1lMajg4R3ZoSmZuNUFtd0xNZkQzL2lZc05zaXJPQldDOXdRVThxZUl5NDZ1THVySTZCTFdzeWU5cwpwQWdveGtTTHdUcnNvSU8vblFDKzNGK3lzdXZVb2ZCYWNHek1xUDJjSkF2aFpqM1lpeCtrNXgra0RRaWhmU1JDCmlDNnJDNFNzWFJRYk13SHFHSnBrM01CcklGdXFrMkVTRy9zNk44NmNIdFBZTWc3UklWR2pKQWMyM045amRVMDIKOUxoK3VmTmJCS2ZYSlEzN2w3Z0tMZmYzL1hxdUJFcnNsY0prQnZ4TWtId3QyZlgrdm5Zc2E0Y1JSZ1NidjlNWgp5UXJ2eCtpMkJjdFJNZnU0VUh3VFVMZm95THYvanYxNXkrSjZmM00xbGNLREpEOWl0SVVicSs3a0FKSGpYSkN3Cm5hYlpZL3JXRnVaejNVUXB4RUNaOWN1TVNnbW5OaHlROUdPY3JmRk9YYzdRTFlQdW85YndYK0ZyTTRGS3AwOXcKdkFqb0RmLzl5bisweDMvSmRUcWdjbmlNcllIb3V6YlhFNGRYVHVRN1BKWVlDZ1U9Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K'
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