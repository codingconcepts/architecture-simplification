# Before

### Resources

* https://levelup.gitconnected.com/aws-run-an-s3-triggered-lambda-locally-using-localstack-ac05f03dc896

### Infra

Compose

``` sh
(
  cd 005_unnecessary_dw_workloads/triplicating_data/before && \
  docker compose up --build --force-recreate -d
)
```

### BigQuery

``` sh
bq mk \
  --api http://localhost:9050 \
  --project_id local \
  example

bq mk \
  --api http://localhost:9050 \
  --project_id local \
  --table example.orders id:STRING,user_id:STRING,total:FLOAT,ts:TIMESTAMP
```

### Localstack

Terraform

``` sh
(cd 005_unnecessary_dw_workloads/triplicating_data/before && terraform init)

cp go.* 005_unnecessary_dw_workloads/triplicating_data/before/s3-to-bigquery
(cd 005_unnecessary_dw_workloads/triplicating_data/before && terraform apply --auto-approve)
rm 005_unnecessary_dw_workloads/triplicating_data/before/s3-to-bigquery/go.*
rm 005_unnecessary_dw_workloads/triplicating_data/before/s3-to-bigquery/app
```

### CockroachDB

Convert to enterprise

``` sh
enterprise --url "postgres://root@localhost:26257/?sslmode=disable"
```

Connect

``` sh
cockroach sql --insecure
```

Create table and changefeed

``` sql
CREATE TABLE orders (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL,
  total DECIMAL NOT NULL,
  ts TIMESTAMP NOT NULL DEFAULT now()
);

SET CLUSTER SETTING kv.rangefeed.enabled = true;

CREATE CHANGEFEED FOR TABLE orders
  INTO 's3://s3-to-bigquery?AWS_ENDPOINT=http%3A%2F%2Fhost.docker.internal%3A4566&AWS_ACCESS_KEY_ID=fake&AWS_SECRET_ACCESS_KEY=fake&AWS_REGION=us-east-1';
```

### Test

Monitor

``` sh
see bq query \
  --api http://localhost:9050 \
  --project_id local \
  "SELECT * FROM example.orders WHERE id IS NOT NULL"
```

Add orders

``` sql
INSERT INTO orders (user_id, total) VALUES
  ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', ROUND(random() * 100, 2));

INSERT INTO orders (user_id, total) VALUES
  ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', ROUND(random() * 100, 2));
```

# After

### Remove unecessary infrastructure

Localstack

``` sh
awslocal s3 rm s3://s3-to-bigquery --recursive
(cd 005_unnecessary_dw_workloads/triplicating_data/before && terraform destroy --auto-approve)
```

S3 changefeed

``` sql
SELECT
  job_id
FROM [SHOW CHANGEFEED JOBS] WHERE
status = 'running'
AND description LIKE '%AWS%';

CANCEL JOB 925652922578042881;
```

### Infra

Run BigQuery webhook worker

``` sh
(
  cd 005_unnecessary_dw_workloads/triplicating_data/after && \
  openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -sha256 -days 3650 -nodes -subj "/C=XX/ST=StateName/L=CityName/O=CompanyName/OU=CompanySectionName/CN=CommonNameOrHostname" \
)

(cd 005_unnecessary_dw_workloads/triplicating_data/after && go run main.go)
```

Get cert.pem base64

``` sh
base64 -i 005_unnecessary_dw_workloads/triplicating_data/after/cert.pem | pbcopy
```

Create webhook changefeed

``` sql
SET CLUSTER SETTING kv.rangefeed.enabled = true;
SET CLUSTER SETTING changefeed.new_webhook_sink_enabled = true;

CREATE CHANGEFEED INTO 'webhook-https://host.docker.internal:3000/bigquery?insecure_tls_skip_verify=true&ca_cert=LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUY3ekNDQTllZ0F3SUJBZ0lVTW5JV2xQWGJzbzZ0em1MNnNRK1JYOUdwd2g0d0RRWUpLb1pJaHZjTkFRRUwKQlFBd2dZWXhDekFKQmdOVkJBWVRBbGhZTVJJd0VBWURWUVFJREFsVGRHRjBaVTVoYldVeEVUQVBCZ05WQkFjTQpDRU5wZEhsT1lXMWxNUlF3RWdZRFZRUUtEQXREYjIxd1lXNTVUbUZ0WlRFYk1Ca0dBMVVFQ3d3U1EyOXRjR0Z1CmVWTmxZM1JwYjI1T1lXMWxNUjB3R3dZRFZRUUREQlJEYjIxdGIyNU9ZVzFsVDNKSWIzTjBibUZ0WlRBZUZ3MHkKTXpFeU1qQXhOVEUwTXpSYUZ3MHpNekV5TVRjeE5URTBNelJhTUlHR01Rc3dDUVlEVlFRR0V3SllXREVTTUJBRwpBMVVFQ0F3SlUzUmhkR1ZPWVcxbE1SRXdEd1lEVlFRSERBaERhWFI1VG1GdFpURVVNQklHQTFVRUNnd0xRMjl0CmNHRnVlVTVoYldVeEd6QVpCZ05WQkFzTUVrTnZiWEJoYm5sVFpXTjBhVzl1VG1GdFpURWRNQnNHQTFVRUF3d1UKUTI5dGJXOXVUbUZ0WlU5eVNHOXpkRzVoYldVd2dnSWlNQTBHQ1NxR1NJYjNEUUVCQVFVQUE0SUNEd0F3Z2dJSwpBb0lDQVFERTJyaitJTnlYbEQ1OHRJcnQrblJ2b2dIOTNIMGZJMk91a2FQUklNaVpVS0hkbjR5MElHcGFPWHZlCjZ4WEZJeXVneEZuc2IzSm05dDBoZ2h2Qk0ycWc1RVdNUDFKTEV5UE85ZVpTVXNkallYSmxZOVFUVG9EenFVcHYKeEtjdnlzcUJ2L3dVQ0NUd2NQT05Bc0tNd3RaMFdKZkJ1T2lVeTI0RzM5NG9zRU9pQnNHbUxsd1ZCUWdDeGtLZwp6b0V5SE5pcHdRMEU4bnVZejVkSldVam4weE5TK1FhWDhaK3I2N0FSL3kyWGwvUldFSHBRQ3A0cmVzdDdwdmtDCitZdVVYdVM5ZlZZRjk4SjR6dVJuUDUrR0VZdk1rbVNoMFNMUVJXUEtuTy9xWGI5a3BxVGVLUkRBVkNUbkNWTUQKa2FvdDRpcHR2VlhaOWE1YktlK2Z2bCtVTEZTQUJHNFdDZnJMMThDUXdlV2tobUFTczJUZjB0QjFvUVJDdS9WeQpneUVhNjFTSDd4QzFyS2pNUVU5REJTUk9SNzNYUlhtdlMrR3JQeW9GVnM1Q1gvVXNWdFk3R2dNYmh0dW9XL2xTCnBhVmRPcHFFNGppNEE0c0NPdVJzTW5nU2lTTkNWZnFHQ2lJMDBqeEx5SlVpUDZGOGFyUk5QTlRBOWJzdE5aQzgKQ1pvVVlKd1FXdVpmVWEybDllZFFsZGNpZDZvOTkzWHNJY0pNRlFtVUJvNlZIeTBWNFdSaFBJYVJIMVo5MitPWgpWNUl4RVE5QmRmV0QveTl3Q1lIWVA4Uks3dnYrenNPWVZkbEtYcmJHOTZTNTVZOHB1VXJXYTUrMTRaYVR1dC9iCkVoN25hZTBWS0hpNnMwdEYzck9jQ25RZ1V3WmkxRzZzdWFPa09HSkwyRTl5NzQvc2N3SURBUUFCbzFNd1VUQWQKQmdOVkhRNEVGZ1FVK3FxNjZtVUFpNlc5YWxzTHFSOGhCNFUyV3VFd0h3WURWUjBqQkJnd0ZvQVUrcXE2Nm1VQQppNlc5YWxzTHFSOGhCNFUyV3VFd0R3WURWUjBUQVFIL0JBVXdBd0VCL3pBTkJna3Foa2lHOXcwQkFRc0ZBQU9DCkFnRUFzc2xXNWxpREJITzY4Q1RyOGI2SFlSTjhVNHdJcy9LY3Q2aUUrRXgzcnh5aTJaSkVKaDdwZ3FCb3A0RCsKMmd6ZERKNlUxNlpVSUtVM2V1VFJISU1yd3VXMEs0YjE0Zmk5NHlmcjRXL0RUcXNIaUpISHUxRkJUVCtlRFNTbwp2OUxNa3M0bXI4L2NrcW5uOWFLUlhBbi9pVGlnR0xXMko5ZGdXMmtFVFl2ZnE0ZkFOd0RVaUdBK29aVHNuVWQzCnVlMXcxQWhqSzAycnZaTnRwSk9TZmU5Q3Y5bnJ6dFJrUHRGTkV2ODFEOHBFY0Y3TkNGcmxjeFNzSzdBWWYyaW8KdmxCWTdHM3JDY0xaOUZrNTBGUGNlbjNqbnlEejgxMGkyOFQraXJWa0IzbXQzMUlDMUFjblNLdW9RME1VZStCMQpVQWRzTVo1Vlo3TUJ3UWlacGNwMjJCZ0lLaFFwV3JmS2IrQ2RDeFdPUjY1WGdFZUI2bTcvRGppZzNSUFZ5dFVpCjA1NVJhcFZ1a0x3a2VFc2JyYjREci9peGpNWVNQZTVXZFF1cnhkTFhKNlYvTXBaamVpNDFEZ0hIbkNzYlBQancKVndYcVBoWENNYUNQdTFybFczd2VoU3NiMWtYTDkyNk03SkVSLyt2MVFTZzI2dXdhMGpobmxrWnJJSmhoMXpZcgpZWURjSlNOOEMzY3JCWTFnbGFubEUvQXNOWFRmQzM4dllRSStJMmQyOEw4dnFENFBWQXB4d0Nhd3lmbHJGZlNaClkyNHZBM1hhMlRHWUNLelFyUk05Y0pkYnpTWjZIUUpCaTVKbFVuY0IwcTdaSWp0dnRmeENHV3lwdDZqZEhlMzkKTEpyeitPcXZDV0NwK1JXUDFyVVl3TVBBNThPd2ZIZy8yYTczd2pYa1dwNURMMEE9Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K'
AS SELECT
  "id",
  "user_id",
  "total",
  "ts"::TIMESTAMPTZ
FROM orders;
```

Monitor

``` sh
bq query \
  --api http://localhost:9050 \
  --project_id local \
  "DELETE FROM example.orders WHERE id IS NOT NULL"

see -n 1 bq query \
  --api http://localhost:9050 \
  --project_id local \
  "SELECT * FROM example.orders WHERE id IS NOT NULL"
```

Add orders

``` sql
INSERT INTO orders (user_id, total) VALUES
  ('cccccccc-cccc-cccc-cccc-cccccccccccc', ROUND(random() * 100, 2));

INSERT INTO orders (user_id, total) VALUES
  ('dddddddd-dddd-dddd-dddd-dddddddddddd', ROUND(random() * 100, 2));
```

### Summary

* Biggest win is a vastly simpler architecture with few moving parts.
* Faster to get data from database into DW.

### Teardown

``` sh
make teardown
```

### Debugging

``` sh
awslocal s3api list-buckets
awslocal s3api get-bucket-location --bucket s3-to-bigquery
awslocal s3api put-object --bucket s3-to-bigquery --key README.md --body README.md
awslocal s3api list-objects --bucket s3-to-bigquery
awslocal s3api get-bucket-notification-configuration --bucket s3-to-bigquery

awslocal sqs list-queues
awslocal sqs receive-message --queue-url http://sqs.us-east-1.localhost.localstack.cloud:4566/000000000000/s3-event-notification-queue

awslocal lambda get-function --function-name s3-to-bigquery

awslocal --endpoint-url=http://localhost:4566 logs tail /aws/lambda/s3-to-bigquery --follow
```