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

CREATE CHANGEFEED INTO 'webhook-https://host.docker.internal:3000/bigquery?insecure_tls_skip_verify=true&ca_cert=LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUY3ekNDQTllZ0F3SUJBZ0lVUjI0Sm5iNVZBMitpd2ZDbWxUamhYeTN5Q2lzd0RRWUpLb1pJaHZjTkFRRUwKQlFBd2dZWXhDekFKQmdOVkJBWVRBbGhZTVJJd0VBWURWUVFJREFsVGRHRjBaVTVoYldVeEVUQVBCZ05WQkFjTQpDRU5wZEhsT1lXMWxNUlF3RWdZRFZRUUtEQXREYjIxd1lXNTVUbUZ0WlRFYk1Ca0dBMVVFQ3d3U1EyOXRjR0Z1CmVWTmxZM1JwYjI1T1lXMWxNUjB3R3dZRFZRUUREQlJEYjIxdGIyNU9ZVzFsVDNKSWIzTjBibUZ0WlRBZUZ3MHkKTXpFeU1UUXhNak0yTVRKYUZ3MHpNekV5TVRFeE1qTTJNVEphTUlHR01Rc3dDUVlEVlFRR0V3SllXREVTTUJBRwpBMVVFQ0F3SlUzUmhkR1ZPWVcxbE1SRXdEd1lEVlFRSERBaERhWFI1VG1GdFpURVVNQklHQTFVRUNnd0xRMjl0CmNHRnVlVTVoYldVeEd6QVpCZ05WQkFzTUVrTnZiWEJoYm5sVFpXTjBhVzl1VG1GdFpURWRNQnNHQTFVRUF3d1UKUTI5dGJXOXVUbUZ0WlU5eVNHOXpkRzVoYldVd2dnSWlNQTBHQ1NxR1NJYjNEUUVCQVFVQUE0SUNEd0F3Z2dJSwpBb0lDQVFETENYZGVaS2xnWU91dVljNVArZXZsUHJNaXNydTZiVHc4REF0NTFkakYySXhqeUE2QVN1M2hLMTFhCjVMeFdxWEJJWnFlOUFZM1JESCs4VzRadUJBd2ZZbE1UK1RhWVl6ZGcxT2xOZ2ZTRGJHc3BVTm5LM2J4NDk3ZXUKRGJIZWc0NGJreEtKL3lYaWR1ckE0cFZsRFV5a2k2VGJiWWRRQmtZQmhFSEluMkExRlJoMEhpRGNiK3dPUEpxZwphR0pwRWN6MjljUUo0ank5bG83UGtpejFkeUFPY24xVERwTUtkZ3dneVBDZUtUbkwyeHpiVldDbUYzS1NCa004CjhaMUd4VVRiMkVzOGNRdkpYMnEzMnUvQ01rZWQ5TWNMd2dTSXRSeEluOUlNdm45UER5RkR1dEhlTlQrclR4Vy8KeE5lUG55NGhKSkZnRGkySzZLTG1lWDhWeHFibGtmQzBCSU9vNEJpQngrb3BuVjVOSCt2ZGFVaTJvdUl6UjZVUgpTVnpENG1sdXBPS0N4d2ttV0I3TTVjUXAwVU0xL1ByQ2VGY2ZWUGdscjRnUmQ5dmdnQ2xUSkhtR01xMWNKT0Q1Ck5WZy8waTFvMVVmdVZmcHNOQmFxckNqcjVwa2JqT1hTVVdlVWdkOWV4QklFVHdNUjBmOFRQdlhnRnRyejIyaTUKNlVBc0NmenE2d1pyWUIrNVFPSzVWYnJ1T3NkWEpCQXFKRE9vYlNkZzVXemdrMFN2YmxaUXE0RkFmYVpyYndhZQp4VXhmczBWQS9kcjBIVm01TDU5R3R5Vmg5VUJvVS8wUlZCaFhvdXpwVW4yanNIMjhCbVBpelB2UlJna3ZHOWNxCklxZ1djS2NhTHl5a1FHSU5pd0QwazdOMmJmWkd6ZG0yOEt2S2Z0UTB6MW9XclVhMm93SURBUUFCbzFNd1VUQWQKQmdOVkhRNEVGZ1FVUkpUUDNoZ0xyeVpWK3VrU1djcWdnWmRMa2owd0h3WURWUjBqQkJnd0ZvQVVSSlRQM2hnTApyeVpWK3VrU1djcWdnWmRMa2owd0R3WURWUjBUQVFIL0JBVXdBd0VCL3pBTkJna3Foa2lHOXcwQkFRc0ZBQU9DCkFnRUFRWTcyZkRCTk1zN1FEcjdaNHppdmNuT0w1eDlJK1d3SEFGQms4YnZIY0EzUFFjQUk3OXZzY0tQT0p6R3UKaWxQSmJWdTRDaldyNlpkdEFJWnZST1NBUTh0czllVHZ5NDdudEtjeVluTWhieTRncGhCZVFGUXlXSy9sNGpqbQpsRkZxUjhvQ3h1ZGRpaGFha2hCQzBmU3J4S0N1Yit4aFBzblZ2MFozOHlpc25UNjFKUEVUYVhaa25ocUdTT1ZqCm13RHlMYlZPQjJud2c2MjZ5eUFtNVlDRVlnaUs1c25yOVBVNHJUTXJtRWtpRnJPM3lOZ2VuQ0IzQ3pOZGNTS0sKS1c1WHlKU0ZzVFlSVGovZVp3VlZFNEE1S0grcnkzQTQ4Q0RJUUNRTS9uN2xiS2dLNjRVNkdCY0g1UmlKdmxicQovQTBTOXRYMUVYS2xzNHcwL2laZnFKelpBTks1SXlIdnNMS01RSjFjNHczV3o0ZHAxSjhiZXk2WUVnbGZuNzIxCmxEUHBZclRLblZIZmhmdFlZKzBtcWZlVmlhWldvbkUzMUpVaTdoZGZWRXNvOTBPRnJHNGcvc1BqelY4cWdWUlUKbjJlR0I0YjBaWUtMTjNFam1JeDY1anFwaFlHUW5Md3FPakRrK3B6SVhoRnZXbUhxZ0J4UE50OGt6Qm4yL0JBeQpWRy9oRi9Hd09ndE5WYldPajhnSndwNnJKMU04ODdzamxqaFJVeEg0ZWdJSzFpeHFOd0dxYmFUV0o1SGV1LzBVCkRIcGVUNm1zREdzbXVXYlFVOEkxTHMraWlJNkgzQmZFekhyTHBKSTJPeDU0dzE1S3JXdTFxY3BuNFpoRVFoWVoKRm9XU2s3SHEwSkdoMW92RUJOV1FxVE5Xa056ZjZZdHhCWElZaEVxWHZ2REY0anM9Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K'
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