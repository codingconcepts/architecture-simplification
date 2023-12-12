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

Connect

``` sh
cockroach sql --insecure
```

Get cluster id

``` sql
SELECT crdb_internal.cluster_id();
```

Generate license

``` sh
crl-lic -type "Evaluation" -org "Rob Test" -months 1 453a0c52-8545-45b9-92bd-863be9e8d54a
```

Apply license

``` sql
SET CLUSTER SETTING cluster.organization = 'Rob Test';
SET CLUSTER SETTING enterprise.license = 'crl-0-ChBFOgxShUVFuZK9hjvp6NVKEKDIhK0GGAIiCFJvYiBUZXN0';
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

CREATE CHANGEFEED INTO 'webhook-https://host.docker.internal:3000/bigquery?insecure_tls_skip_verify=true&ca_cert=LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUY3ekNDQTllZ0F3SUJBZ0lVUW4vVm5sUC9hcnBaYzZjOGNhMUx3UVJHK3NRd0RRWUpLb1pJaHZjTkFRRUwKQlFBd2dZWXhDekFKQmdOVkJBWVRBbGhZTVJJd0VBWURWUVFJREFsVGRHRjBaVTVoYldVeEVUQVBCZ05WQkFjTQpDRU5wZEhsT1lXMWxNUlF3RWdZRFZRUUtEQXREYjIxd1lXNTVUbUZ0WlRFYk1Ca0dBMVVFQ3d3U1EyOXRjR0Z1CmVWTmxZM1JwYjI1T1lXMWxNUjB3R3dZRFZRUUREQlJEYjIxdGIyNU9ZVzFsVDNKSWIzTjBibUZ0WlRBZUZ3MHkKTXpFeU1USXhNVFF3TXpWYUZ3MHpNekV5TURreE1UUXdNelZhTUlHR01Rc3dDUVlEVlFRR0V3SllXREVTTUJBRwpBMVVFQ0F3SlUzUmhkR1ZPWVcxbE1SRXdEd1lEVlFRSERBaERhWFI1VG1GdFpURVVNQklHQTFVRUNnd0xRMjl0CmNHRnVlVTVoYldVeEd6QVpCZ05WQkFzTUVrTnZiWEJoYm5sVFpXTjBhVzl1VG1GdFpURWRNQnNHQTFVRUF3d1UKUTI5dGJXOXVUbUZ0WlU5eVNHOXpkRzVoYldVd2dnSWlNQTBHQ1NxR1NJYjNEUUVCQVFVQUE0SUNEd0F3Z2dJSwpBb0lDQVFDcWZkZlhoSEJ0aVVxUXFZSWNQUHJHbmQrVk9HVDdrc2ZTYmV1YVVDZ1IxK1RvVkFYaDlIYUUvQndBCjlKSVpUUytpOWoyM0F0ZWNkckRjNVd2aTZrRTlDTUU2UzlmZWFFaDRVb3hETkx1bmpzdGFoVkM1THR2dmlnZUMKOTBTWkx3RjFORmtna0VjUVJOWTNmM1BDNFhHTStWQndRaXBEVk5veG8yQ0lKdjF1ZFR4Z004cWJsTWg2dndEQQozNllFMkRNL0tGL2JxaWxXelduU1I5V2VSdU5paXpTT0ZuUURLTDhBbjJWS3huY3BKMElrTmRITHJIVE9MbDZzCnVrU0VtRU1QT1N5RDNVdmIrZFdzaGtRM2RiRStBenlzMnFnYzk1VVQwMmxPSEJqWXJ3ZFVZRXlZdWlnQWlmckwKOWZpZFhEamp0ejRkZUlCbk1tekwzZjVlYktMdEhOWTZadnNUY1BQcDhTc1U0TjlOTVRGaEcweFdVWDcrdG5KMApydGJtb3ByZURhU0p0YVFPL3V2VVNwRElOVkkwYzVVcnpqUUlDb1NUVkgyOVdmdWZlTkh5VUdoYmtxUVFkeDJ6CmI2NU8xL3RJa3ZlS2xYTW5EYlZ3N3phcDJ3c1NxZGFPKzh5aWxOdU1oWGRkZnJjOVY1d3Zhcm9hcHdDYlRWd1UKd2l3bGJ0L21LNWplRjI1UWl1dnpycFZmZXBYOGhRWVlSQytMUlRPVkRGOE81SVNFa3QwWHRWNDBSL2JQMU5rNwpTOW1nMVhlZk1QN1pkdTJja2o2OE40dmEvYitLUFhpMWNsSVR6V1YzaklUMzcvemNaWkpaaWk5WWVBL25FV2s3CllYV1RFRXY0V3lMRnRkS01rNzFzeU5TVzg4R0FEY082T2g2dTFhVmhBU3ZXRmQxSGR3SURBUUFCbzFNd1VUQWQKQmdOVkhRNEVGZ1FVbUJXRG1ZelRDOENxVks5Z2p6RGprcmJhblpJd0h3WURWUjBqQkJnd0ZvQVVtQldEbVl6VApDOENxVks5Z2p6RGprcmJhblpJd0R3WURWUjBUQVFIL0JBVXdBd0VCL3pBTkJna3Foa2lHOXcwQkFRc0ZBQU9DCkFnRUFhVU1PWEdjNzBoeXpkOVh2TEJ4M0ZTanpJUXJzdlRkODQzaFM2djB3WGJManZoZzZnMHhGSzIwa1U5V0YKQ1JQeVQ4MDduKzJzelFLQ3dEQzQ2WnQ0WUR6eGpDWjZVb1FJQWU3S0dWYUtWY0dpMWlIZzJDSGRGekFSWUNhQwpMdjJRMEY2UFdyQ3V1RlpTSERpMkNkNndCVUdhbVViYTl1MmkyWWN2UVlZdFU0UUhOMERta1VVbi9uTFNSWlZUCmFSWi9YUUZTZHZlaDVJcWtYcjVLL21TeWp1UnpHcHNTbmxxTXAybVBPSEpGVnBRNUVmTXpmZ2dMbklxVEQ4NFgKbEYvWm9GSDNQZEFpaEQxM2xMY1pZaUo5UDQ4QjVnak9Ob0lYZEgvRFg0dXd1UVVlT1JuMUxFSVREbTZkVFlpcwpNNFpCa3RkL2JUUnU3RldrOGRSZWluNlAyNWxzU1JzVWs0VTREc2NQcEVLMmFRaE93Z0NxKzRnaEM5bDMxTFZqCmZxMG1UcmFFd04zbEhBTDVZMXVSa2JnTG05M0NlOW00elpoem1sVys0ZCtHa25lRk5MZml3WitwSUFtbmRzeXAKdHdLOWI5NC9scnk2WWZ2cDRZZSs4YjA4d2txTHB4UXhGTkZjdTNKN0lqMFhhQ3IzYTdlZTU3cHFhS2RPK3RDbQpkQU9WNHlaUTJIY21VOHpTeHZqMmtFb29oNG1oaTM2OGhYa3NEbWs4QzAyUDFQZm9jMmlYTktITU9GdGtzSEluClJaMkVaQ3FsNGdkT1hOZmdXRVVMZjJ3a2NiNDhGUzJlRDhYMVF2Vi9SZWtGeXlkRk1rc0E0Si9pb2cySHJOTEIKNEcwMTNBRDRmMFZxUUN3NnZWcGZBNHFRZVZiMXRvd2JLRWE1T0Jwam1ySkE1VWs9Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K'
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

see bq query \
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