# Before

**2 terminal windows**

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
see -n 1 bq query \
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

**3 terminal windows**

### Remove unecessary infrastructure

Localstack

``` sh
awslocal s3 rm s3://s3-to-bigquery --recursive
(cd 005_unnecessary_dw_workloads/triplicating_data/before && terraform destroy --auto-approve)
```

Connect

``` sh
cockroach sql --insecure
```

Delete S3 changefeed

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
SET CLUSTER SETTING changefeed.new_webhook_sink_enabled = true;

CREATE CHANGEFEED INTO 'webhook-https://host.docker.internal:3000/bigquery?insecure_tls_skip_verify=true&ca_cert=LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUY3ekNDQTllZ0F3SUJBZ0lVVGtiMlUwUDNLRHBqVHNLWmRIK1dqVVI1K1lRd0RRWUpLb1pJaHZjTkFRRUwKQlFBd2dZWXhDekFKQmdOVkJBWVRBbGhZTVJJd0VBWURWUVFJREFsVGRHRjBaVTVoYldVeEVUQVBCZ05WQkFjTQpDRU5wZEhsT1lXMWxNUlF3RWdZRFZRUUtEQXREYjIxd1lXNTVUbUZ0WlRFYk1Ca0dBMVVFQ3d3U1EyOXRjR0Z1CmVWTmxZM1JwYjI1T1lXMWxNUjB3R3dZRFZRUUREQlJEYjIxdGIyNU9ZVzFsVDNKSWIzTjBibUZ0WlRBZUZ3MHkKTkRBeE1UQXhNVEV3TURSYUZ3MHpOREF4TURjeE1URXdNRFJhTUlHR01Rc3dDUVlEVlFRR0V3SllXREVTTUJBRwpBMVVFQ0F3SlUzUmhkR1ZPWVcxbE1SRXdEd1lEVlFRSERBaERhWFI1VG1GdFpURVVNQklHQTFVRUNnd0xRMjl0CmNHRnVlVTVoYldVeEd6QVpCZ05WQkFzTUVrTnZiWEJoYm5sVFpXTjBhVzl1VG1GdFpURWRNQnNHQTFVRUF3d1UKUTI5dGJXOXVUbUZ0WlU5eVNHOXpkRzVoYldVd2dnSWlNQTBHQ1NxR1NJYjNEUUVCQVFVQUE0SUNEd0F3Z2dJSwpBb0lDQVFEai9NUXhqR1lTUFpzVWxDb2NXdGpaN0pJeG80Z0l2QWZDaG1aYTlWa1BJT0VwTUIrRG9SU1ovN2pkCkk5cXNKeVIxb0tXT0VDbzV3NjJjLzYwc0VYWjdJQWVpQkNqNjhZQ3JPUk1uVVRUUUpDWUp0TkF4MGpDRWRMSWgKN2tPYnlrVHVmR0grMXkzc003U0g0SnpiOEx0bXlPUGFic0xqemFFQ3AyaUw1NUFuVDJJV21FSmdCZ0xqV1gvdgo3MTZzK2s2bmc5NkprajdZYkVYT1pmWHhyd1RkNy9rQ3dEM0RqMUdROEh6TWV3dGNleWo3SUV1eGV0ZGc4NzFFCjRYc0lyak5DRXFKdWdsdHhxeEZPQ0drVEpWU051U21kSExRcE82WktrNzBDWmtlMjlwTS90dUZWaGpzZnhHaEcKMUxmNURwR2lVTDlYZ1hJN1BwZ3dBY29vYjhueVRlbktzc3c3a3l4YVEyR2h4NjN5ZnNBdmNidWFBQldDYzhFKwpHTTZxMzhEdThSOGJYVFF3Z3lyZWp4Y25wUXNkUjdRdW13UlF5NjBaYTFWZWZ4Q0J6b3orWUVSdSs5VDhkRkV4ClFONkF6dTBuWFFISTBrbHJRSTg5UWxtTUlka3NUOEFsa1U4N1YrSk1FMUEyQWw1MFo3dG14ZmxQdmZ5K0Q4em8KRmhzNmkyZDNUc2xBT0dxSUU3d20wWXh3dHBTL1lKNW95NzZwV3Vad0ZDcnEwREVBays2VlBYcWhIS0owaURXRwpFN095dVdmZFZxUDdQSG9TQUV1ek5LYUs0QUhULzRDaVpCVjF5NkNMemVDS0t1RHpoRXpZdG5xcEJ2VFpxdEU3Cjdpa2k5MkZuaWRxWUI2V2NEcjdVQnVPRjBZeE1HQUdKWHNvNnd3Wk5KSFhWTFBOQ2l3SURBUUFCbzFNd1VUQWQKQmdOVkhRNEVGZ1FVZnRuWDZ5ZUxGVUo1UnRHSVB5S2lwRlhJQ0w0d0h3WURWUjBqQkJnd0ZvQVVmdG5YNnllTApGVUo1UnRHSVB5S2lwRlhJQ0w0d0R3WURWUjBUQVFIL0JBVXdBd0VCL3pBTkJna3Foa2lHOXcwQkFRc0ZBQU9DCkFnRUFMMUthWEhxZDk0VW5KWnpqcTc3QTkreU93ekpTQTdtNWk2OXhIczBUdVBXc1RIZWVDdFpOSmdrOG4ycEMKeGJlSDBVa2wzaEZpQjJMZTdqVGtRQlk4MGlMVFRka2o3U0xnMXRuUDZWY29jNU51U201dkNDLzZNK3hORWs4UgpqbERLYzJOYUxwczBrOVNuSXdrLzEzU3J3UExHZkcvbCtVN2VSTDZSNTBMRlVlQTlUNFdIcTk3UGdwZTBzNEpVCm8zTnlad2hGYUdSTW9TRWQydWxHa2FMdmZCc1JTSFgxdklnMGFKNFJ1ZzZ6NWpKV1lHWnptZENqRXNySWk3NloKUWZUQnoxQXFFUVNFQ2hBN3E5by9ZUUhwOUJnQ2YrdzZkbytYakt5b0ZNTnNRQ202bnZWR01WVjM2cThEVkx6KwpVNUdzakxNdkxRRmdCWjdOY0tmZkJoUmtiK3F5cjNYQTZCbmhXbXpGNkh5K3cxZXZkRm5vRGtVNExLQW4rbWFJCkQ4aXN2OUZmTFlaTXpySGVpM3phZ01POGU3OERKZzdMM2RCMEdTUnZsS0JWL1RNb0ZEdExId1FUYkJwNjE5eXYKOThXQU5KYnJXazJpdWMraWYxN2VkN3FXWGN6NmZKL1MyOG5VOHJJZWVMRGRKUjZQY0k2TXFwbURIaG9laU0wMQpUWVZESGhIOXJLdG1qbmIwZjI1WGxoOEp3T2ZJR3N1VXdscUxsUjBwQ3Z4Y2MrTW9RT1piM1puV3N3bXV6V2NxClpNb2JRTitVN1FZaXRiNTRMS05NWEk0OGQ5NGNqOGE1dGVveTl2cHVjQjFaMVY2VFEzeWEvMTNUMW1FdnQ0Z08KTVJZUlRHN3Awckl2NENkdVNocGwzMnZVZU1UNlRNdjZ2NDJMRTFmODFvcnVybXM9Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K'
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
```

Add orders

``` sql
INSERT INTO orders (user_id, total) VALUES
  ('cccccccc-cccc-cccc-cccc-cccccccccccc', ROUND(random() * 100, 2));

INSERT INTO orders (user_id, total) VALUES
  ('dddddddd-dddd-dddd-dddd-dddddddddddd', ROUND(random() * 100, 2));
```

### Summary

* Biggest win is a vastly simpler architecture with few moving parts
  * In before scenario:
    * Any of the components could have failed for any reason
    * More network hops means more changes for things to go wrong
    * All of our components would have needed maintenance
    * We're at the behest of the SLAs of each and every component
    * If one component doesn't offer at least once or exactly once delivery semantics, all bets are off
* In the after scenario:
  * Less data duplication, which, for a large dataset, means:
    * Less network transfer
    * Less storage
    * Less cost
  * ...and as we saw, data also arrived into the DW with lower latency

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