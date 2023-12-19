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

CREATE CHANGEFEED INTO 'webhook-https://host.docker.internal:3000/bigquery?insecure_tls_skip_verify=true&ca_cert=LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUY3ekNDQTllZ0F3SUJBZ0lVTk16TXEyaGtHL0VpaEpqNXJWMW5LVkJ5SU5rd0RRWUpLb1pJaHZjTkFRRUwKQlFBd2dZWXhDekFKQmdOVkJBWVRBbGhZTVJJd0VBWURWUVFJREFsVGRHRjBaVTVoYldVeEVUQVBCZ05WQkFjTQpDRU5wZEhsT1lXMWxNUlF3RWdZRFZRUUtEQXREYjIxd1lXNTVUbUZ0WlRFYk1Ca0dBMVVFQ3d3U1EyOXRjR0Z1CmVWTmxZM1JwYjI1T1lXMWxNUjB3R3dZRFZRUUREQlJEYjIxdGIyNU9ZVzFsVDNKSWIzTjBibUZ0WlRBZUZ3MHkKTXpFeU1Ua3hNRE0xTURSYUZ3MHpNekV5TVRZeE1ETTFNRFJhTUlHR01Rc3dDUVlEVlFRR0V3SllXREVTTUJBRwpBMVVFQ0F3SlUzUmhkR1ZPWVcxbE1SRXdEd1lEVlFRSERBaERhWFI1VG1GdFpURVVNQklHQTFVRUNnd0xRMjl0CmNHRnVlVTVoYldVeEd6QVpCZ05WQkFzTUVrTnZiWEJoYm5sVFpXTjBhVzl1VG1GdFpURWRNQnNHQTFVRUF3d1UKUTI5dGJXOXVUbUZ0WlU5eVNHOXpkRzVoYldVd2dnSWlNQTBHQ1NxR1NJYjNEUUVCQVFVQUE0SUNEd0F3Z2dJSwpBb0lDQVFEVFc5VTJRRzZQeWdmcXJZMy9teHBFbGdwS0wxc1I4dGVRRVJvZXdLeFo2UXYzQitQc1l4UVIvN0ZTCjNtSy9reUJtS2VCOHVibXU2VDIxclhNbG4rSVdRZC95V1pHdy9qNDRQQVh6aGY2c2Rxck8zRU8yalpzcmpUOEIKK2ZpM3NpcjFlU2RheUlTQ1VQblJuczNqdUR0Nnp5TUFDVTZCbU9uaXNiRzlUOTg0bnlva1RoVi81M3ppcHBNMgo3bWhPRm1nTHI0ZVhJeGtPc2RvVVhCaVNTcHVzdXBTdWt3S0dLSmRMVCtSUlduOWVmc0R2SG04SzlrQXpIbk1BCmszZXBEMVhubHJoZVpYdiszemo4YXdCTWl0NzUveUUwUFRENFVhU1o5Y25zR05hbUdRZTZmOEl3Yy9ETjNwRGMKV1FyeHM3RkViSjgwSys3dGxCRlBFZW1TMXBhZ0loVXlUSDYzWU5tQlNFWlJQSHE3eGNZZVlCL0RyTWRvUG5qbAo0ZzdHdHVMUXA1ZGpkQk9iM3NjWXZzSEdTQ3d6ZEUvR0hLZlpWNEtxdlE3WUhVWStLNlBpdGlVZjVlbktxaXN5ClBpWnR5SW1kdGlXOS91TnptbDVwVWExMGEwWCtpTTFjV0Q5c2dGaUpoNkJrbnJRZy9ZQ3BkTEptS0o3d1c0ZUEKZW5GbkNGNXVCZ3ByamYvSWhyWVFFbGNCdHlEeUFNeXpIWHRsaXN5MVdnMllQOGRjeXl3eVZVeFRuUEhlSE4yLwowZk9BbXQ3eEJNcVluRW9aZkFqaXd5SERxN3V3dXY0cVdOOEQ4d0lFQ0dVclA2Y3ptQmVEeFJDYjJRVTluM2pmCldmTmxnQTFDOHRpekNEZTRjcUdIUWtxZEl2OVJZdHhlNXIya3R0YjlVc0laL1hSVVRRSURBUUFCbzFNd1VUQWQKQmdOVkhRNEVGZ1FVWEYvUGJEY1NZb0lnNmdOSWNtREJTbnRiMHUwd0h3WURWUjBqQkJnd0ZvQVVYRi9QYkRjUwpZb0lnNmdOSWNtREJTbnRiMHUwd0R3WURWUjBUQVFIL0JBVXdBd0VCL3pBTkJna3Foa2lHOXcwQkFRc0ZBQU9DCkFnRUFablN2cjNYYkZCdWluOTQxeGVTNEpjVHVBanhZZjgwUE5ZMktFYmRScXFnSzFSb1ZIYndIT05LQm5QekUKRHBObVZVZzBwUmEzczNQR0krODVDT0sxZG5HMktIRDdsbXdpYlkxQkJxRCtqcjJvUzFuK0FFY3RKc2hMT044egplaGN1a29Ca2VBRkVXQnNkS1EzekFMUmt0RzllbXB6VDI1MDlaQlNoT3RhbkNHZDdnckNsZzVkTDFtTDFaT2NPCnh6RWJpL3VsbnA0S0dqMEo4SnRGTzhocjJhVURmYWNmQTcwNStqZ09qWm0xbU5qRGVqTVlqdng2OUFiSWZHMWkKTHUzRWJPMmRoMm9kLy9iNjRBRlhMOWEvQnpJQ2lmYThSeENaRVdGUStDU3BzbFFSaW1TU0liZVZuZUxZd1FoZQp4KzZJQ045REhqZGJmSjRHU3U1Zi9rZzJQZ3k5U0RWWjNOVDBwQnVkamsvNFFKNitheXp5emluTXk3c0NVZEh1CmZtQTlwNGJtZ01FaXIxWnMxS1VrVUVjS2x0d2x2eGVwR2xOWEVoREdycC80NEpIOGp3cnl5L2xFaEcrT1hlOVcKOVhZenpnSzdaNHdhTG5MNHNuMkc2REJJbGgrU2NJMXQrVVBwUGFHV05MRGxjMjh2MHc5ZUYxUFEzM3J0QnlUawpybndWNmxsSzJjeGtRMEliTlRYUFVxc3JoMFBCYzZTR3prYUltZmZ4dmQzY2NNZGVCdlZmVlZyMzlpYkhCdkxSCmdGZnN5VXFjUUxYZWpjNzREWTRkSHEwbWdjelNEY1NwM2lVVlYwQXNNWEQxQmdwbFZPZVJSUDBTREE0cHF1WksKSWFEWjBNWTZJVnVPSytYT0ZuVFdseEJZUXNTS0RIeHoyU3hxSFVGcDhvelhGaVk9Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K'
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