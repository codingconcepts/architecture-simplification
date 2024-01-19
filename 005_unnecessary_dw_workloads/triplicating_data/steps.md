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

CREATE CHANGEFEED INTO 'webhook-https://host.docker.internal:3000/bigquery?insecure_tls_skip_verify=true&ca_cert=LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUY3ekNDQTllZ0F3SUJBZ0lVVVdMVElUTG5ETGRoSXpibTRaSnVNYUtOYWgwd0RRWUpLb1pJaHZjTkFRRUwKQlFBd2dZWXhDekFKQmdOVkJBWVRBbGhZTVJJd0VBWURWUVFJREFsVGRHRjBaVTVoYldVeEVUQVBCZ05WQkFjTQpDRU5wZEhsT1lXMWxNUlF3RWdZRFZRUUtEQXREYjIxd1lXNTVUbUZ0WlRFYk1Ca0dBMVVFQ3d3U1EyOXRjR0Z1CmVWTmxZM1JwYjI1T1lXMWxNUjB3R3dZRFZRUUREQlJEYjIxdGIyNU9ZVzFsVDNKSWIzTjBibUZ0WlRBZUZ3MHkKTkRBeE1UY3hNVEE1TWpoYUZ3MHpOREF4TVRReE1UQTVNamhhTUlHR01Rc3dDUVlEVlFRR0V3SllXREVTTUJBRwpBMVVFQ0F3SlUzUmhkR1ZPWVcxbE1SRXdEd1lEVlFRSERBaERhWFI1VG1GdFpURVVNQklHQTFVRUNnd0xRMjl0CmNHRnVlVTVoYldVeEd6QVpCZ05WQkFzTUVrTnZiWEJoYm5sVFpXTjBhVzl1VG1GdFpURWRNQnNHQTFVRUF3d1UKUTI5dGJXOXVUbUZ0WlU5eVNHOXpkRzVoYldVd2dnSWlNQTBHQ1NxR1NJYjNEUUVCQVFVQUE0SUNEd0F3Z2dJSwpBb0lDQVFESGZTdGQrR2Rwb0wvbUM3MGNmRGJYaXJndE85TXdBdmxudlFaNGhEUm5HOW01clVWWGtoblR6cVhKCi90djR4NFo2WXJQOUc5b0VEb2oyZlJZUWFJUmhtSG1ub2p3QVpXMlJpRkh6WTNDb1QrOVNyMkhRQ0R4dSt0VVUKZmtESHlZRld3empsZGVHZXdEM3d4WVlXZkYzbW4zaTY4Y3JCNVU5MEdDeE9KbDRNcXU3MkdzTzBNbVpyczlJcApLV3laeU9lOEg2RHhkU2NTWE83QktTdTNTODZiTmR0RVJVeW40dC9TUEE1WXZaS1VvSFVQUFUrNXNUTitTNkg2CnF6SlFHRCs0dGJqU082RVhXM1V5VjZrc2pjeXZ1UGVyUnJUQWJrTExqYml3L0RVb2FtR1ZlbkhueitERHF3NzcKS2Y0d1JSRWswWnBaWkZ2cUlzWElycDVwQlBhcWFZUWlMcUdHU2E5YjJSTVlORGJBU2xJV1J4c3ZtdDdHK2VRWApoN0hpTHhnNWpzQlNGQVRZbmw4d3pCZ3I4dERTbnFMSmVYaWt0ajlQSnJKUzNpM2hPbEhsd2Q0RzRnVnNUQXN1CmI1U0VCWFVEem4ySmJhRUpiTmFMNElvY1JkUGxqRDQrSlhVRWJMMFpzR1dpZ2NvVGhuT0JrdktOSVJmSFE2ZGoKbVh5NVcvNFlnOFpRQ1BraUlpTUI2Z2J4dE5sVjVxNWJCZndTVVRxRUQrRHRZbUhFVW5Da0ZrZTRXY3cxdElTRApMbEdmZVB1OHpkYlhDTkpFWUo5SDdUMHFmRGRMdVdxdUQ2OEdjYUp4cTFmQ1VFQ25mWjNUcUpXRmlGUG82STMrCkxxY2VVWFd6dmNhR2RkbW9rMXFJOUtRTFh1T0NobG1yODV5KzBQc0gyTW9vemFmMHl3SURBUUFCbzFNd1VUQWQKQmdOVkhRNEVGZ1FVVFdhNHhPSGx6MVpWU3FWc21RSEtIdnVPVWpZd0h3WURWUjBqQkJnd0ZvQVVUV2E0eE9IbAp6MVpWU3FWc21RSEtIdnVPVWpZd0R3WURWUjBUQVFIL0JBVXdBd0VCL3pBTkJna3Foa2lHOXcwQkFRc0ZBQU9DCkFnRUFnaWNlS1ZrVDBsMjVKUmljTngzYWc1WklTU0J2bTVnb3FQR3p5ejR0dmtCMU16U3VFeGZMVlN2dXBmdFAKVzlpempSYlNZRGhGSVZwTlVsVUoyU1VEOVNSazNBbmg2VlI2VnQrc3d1K0N4T2tFbkJkb3E3TTE1ejUva3p4VwpCUElwL21iUDlPWkMxSWdoMlNBZDk4TUk0L3RuU0xMWFQ0QjhIZjNLNkJDZk9GU2drRlZwclZCZDVEanRpWXVBCkFjcEx4MHBNK0NsQVcyZzF3NDdWa2RnMUQ0Nk90eEpkZStOdUlOdU1heGhqeFdnZjhhaFNESFZkdXpDcFpjUG4KTlVZUmU4eUJBUDNzZTlDSTFCK2ZLV0h2cjZLYXBtMW8weHA3V0lNVGUvNmhNZEp0a3NOTUtJOXZQQi95UkppUQpkUm1MTUtVcDNzaEJVMW1lQ1B4RnJ3eFBsd1ZuejRPNHMzc1FYSGk5aWR0d2xZditXM2cxZEZmeDdhU3kvcW5ZCkF5aDJBNzJjMXFzdStndGhJbVg3MU9GeVVMcXdFZUNvckt6NjROalJjWWFUemZLK2w1bk9jZ3ltaTd5L3FGSTAKNUdzWmlSWGpid0FaTDVLR2lIbVozTllNdjBFUFNCaFF0MXo5ak1sTnp2Y1NMZlIzdDU2T0tKMkMvVkZzZ080QwphdHhSOTI5WG9PaSt1N3E5TUVBQ0ZsZ1pZazB0dXNrS0cwZHcyVXRmN0l1RkdzbnpaZ1pUUHJuc3Q5UG1QWlJQCjZOSERrNG5ReW9tbFplM0xrLythTklwTWtQZ1Z0Q29STEtiS0RKYlJYWjBnSEtyTHRqSllGTGM3dHZYNVFXWHAKdTJBL3Z1RnRCMUlNR1RXQzQxdTA4ZXErVGlEaFl2VkJnUzlmNW8wci82WDdZVEk9Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K'
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