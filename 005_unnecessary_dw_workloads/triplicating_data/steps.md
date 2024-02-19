# Before

**3 terminal windows**

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

CREATE CHANGEFEED INTO 'webhook-https://host.docker.internal:3000/bigquery?insecure_tls_skip_verify=true&ca_cert=LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUY3ekNDQTllZ0F3SUJBZ0lVR2dEaHphaWdacXJuSXdFdjNtcHBIM050ZElvd0RRWUpLb1pJaHZjTkFRRUwKQlFBd2dZWXhDekFKQmdOVkJBWVRBbGhZTVJJd0VBWURWUVFJREFsVGRHRjBaVTVoYldVeEVUQVBCZ05WQkFjTQpDRU5wZEhsT1lXMWxNUlF3RWdZRFZRUUtEQXREYjIxd1lXNTVUbUZ0WlRFYk1Ca0dBMVVFQ3d3U1EyOXRjR0Z1CmVWTmxZM1JwYjI1T1lXMWxNUjB3R3dZRFZRUUREQlJEYjIxdGIyNU9ZVzFsVDNKSWIzTjBibUZ0WlRBZUZ3MHkKTkRBeU1Ua3hNRFF5TXpoYUZ3MHpOREF5TVRZeE1EUXlNemhhTUlHR01Rc3dDUVlEVlFRR0V3SllXREVTTUJBRwpBMVVFQ0F3SlUzUmhkR1ZPWVcxbE1SRXdEd1lEVlFRSERBaERhWFI1VG1GdFpURVVNQklHQTFVRUNnd0xRMjl0CmNHRnVlVTVoYldVeEd6QVpCZ05WQkFzTUVrTnZiWEJoYm5sVFpXTjBhVzl1VG1GdFpURWRNQnNHQTFVRUF3d1UKUTI5dGJXOXVUbUZ0WlU5eVNHOXpkRzVoYldVd2dnSWlNQTBHQ1NxR1NJYjNEUUVCQVFVQUE0SUNEd0F3Z2dJSwpBb0lDQVFDcG9kUld3UmxkbzZ4NUxpRERKZUx4Z3lJZHZra2Z1cUFoU09qdnJ6WFgwNFRwYlExYnFaM0hPdFBXCnZPSzMzZUd1eDNoR2lxWEllaFppQTl0RmdDMG83dlJqNzZPdDllWmJFNVJLNEhJeVhzQ0ZtT2pWbCtjTVZacVEKUnhXcGQxWnNLWTBMYU0vWUZCRVJsZUIwWWVCdk5XVXpjZ3B5ek1HeVRkTU80cTdKYUFXcC93bFZmenBza05EMQpaS0R5V3UzSXlaaUxZNDJyOWcrb1FzZnRWcC9zVHFDcm14VzVjSXZzNEhBZGdDcXE0UVFEUFhQa3lLQW5Oa3JVCkt1RVgydERUSVJqSUttMC9rTEZ4bFdEWktaZFp5RXhaeTh0U2RseGVXdGdZdm9KcWRyRlVKZlFHdlRGRk5OaXMKUU9id3JRa09lY2YyQld0c0JSS2FhTDM4UHEvK3pEbzU5QzFJWUxjR3NTOTF2U3RBRFZyeEwveE1jWHhRVEJtSwoycHNYNDFmRjFRWFQ1QTNzV3JBZmxTS1RmNDl3T0FqS1c1ak54bHYrMWQzYVZrcnB6RWhrVmk2VHpCbVB1LytTCjB6RXNZalBwNDNiZHFwRmo4cnJqdy9YVm13azRJdHd1OG1vbEhkemVRNThXNHVleG9uUWVsd3VzaHk5QnRrUmwKRkcxY2NMNWd5OEplU0phdmxlTEFZdUMrc0kzVS92a21GNXBUUVpYVUV2ZHF3Mkc0QWxLcUlEWHE5ZTd0bDE5YQpBbnZXWk9UenU4K1N4NTN5Ulhib0lQVm9LS2U2c0J4c1RKbVpXMzFEcHBWREZqb1puaWFPYlhyb3VmejN6L0pFCnc4VmNPTTJjUEFyU1lZbm9vK3JFRjJ4bExPZ2pzRTBhdFZ2NmgydUtucU9STDIvNkpRSURBUUFCbzFNd1VUQWQKQmdOVkhRNEVGZ1FVT2hJdVZCVGcyY21RU24waERMVHZaNUgwSngwd0h3WURWUjBqQkJnd0ZvQVVPaEl1VkJUZwoyY21RU24waERMVHZaNUgwSngwd0R3WURWUjBUQVFIL0JBVXdBd0VCL3pBTkJna3Foa2lHOXcwQkFRc0ZBQU9DCkFnRUFiUkRWelMvZTQ5ZzJnR2VFL0NWU0I0dmNoS1VjQjhGajV5YTVadnVteEtqaVUvb0pKVENCc2hkR0dJRHAKSUQ3Nm9ycGlGZVl0RmRZbXNFQVJYY0N6ZnJBRllWNldaZ0VMaXJmZjY5RWxsbisrRitQQlBtdW05c3ZaWWFsZgp2MlM1dDFWamVyeS9FQ0s4cG02bzZsbzZndnBjeUxNUXdPVm9QNklKRjdmbzhuMGNCUDBaeWxZZlp3Q3VNa014CkFzeTZOcE16WklOR2ppRmxQVitJTW5NRk1UVXo3bkltamFIV3loRTY0OEM5UXl5MEZXSER4WDFDVmhGSlltT1kKMEJSUXhSZmpyWUJ2OXBuNFpTbHlwbXhubWVYaWtkOWNsc1l0RHZ3aDZ6Lzl0SUc2THhvNVpVUW1BUFhGcXR2dwpQbGJjcjFSa2w4MVg0OFJmRlpFVldvQkVRYkpVbXVSTmU1cWtZeEY0bkZXRmNidkxmbmw1UTBDWmI0cWdwOTVDClhYZWFUamVkMGhrTzc5TmNwNWJxUXBOZEd0TnJzVUozSElwQjFVRHYvQWRMcklaUmZpTldxWWY3NzBYenA4VUcKc0hMNTRrOTVSQ2UrQzY0Q1JyVlJ5SXVmOFBmcFZIblZwaitoN2lLZjFVaWx0TkNWNytFOGdYR0JONCtoZkRzUgpvTVRoYWgzUWMzUS91ZmplOGlPRGFxajd3M3QwYkVDV2pHeFZRK0k1eDdGL2FhTDdJa3lVUEhnUWhjMnJLQkNRCklVZkQ2U2w5d0IvMjNXRW5jUjR3NGI5L1FIeENvUVBsS29HODZEd0E0dFhJT204bzNyWldCTjhMOURMOEtOaVkKT1Fuc1luVlNWRUFiTGZrNW12WDRtSm51UzhOcEhWSW5LZldHUG1GL1ZCSUNmZ1E9Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K'
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