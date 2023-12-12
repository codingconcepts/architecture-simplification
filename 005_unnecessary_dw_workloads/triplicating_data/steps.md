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
crl-lic -type "Evaluation" -org "Rob Test" -months 1 265b8f5b-74fc-4dcf-b4b7-c321872bb870
```

Apply license

``` sql
SET CLUSTER SETTING cluster.organization = 'Rob Test';
SET CLUSTER SETTING enterprise.license = 'crl-0-ChAmW49bdPxNz7S3wyGHK7hwEOWIhK0GGAIiCFJvYiBUZXN0';
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

INSERT INTO orders (user_id, total) VALUES ('7559c90e-4990-4676-bbad-b8a65e8bc01d', 10.99);
INSERT INTO orders (user_id, total) VALUES ('ca6c14bb-a49b-49e0-ad3c-d754eec17f22', 21.98);
INSERT INTO orders (user_id, total) VALUES ('bd3b5245-664a-463f-967e-140d4a229691', 7.37);
```

### Localstack

Terraform

``` sh
(cd 005_unnecessary_dw_workloads/triplicating_data/before && terraform init)

cp go.* 005_unnecessary_dw_workloads/triplicating_data/before/s3-to-bigquery
(cd 005_unnecessary_dw_workloads/triplicating_data/before && terraform apply --auto-approve)
```

### Test

``` sh
awslocal s3api list-buckets
awslocal s3api put-object --bucket s3-to-bigquery --key README.md --body README.md
awslocal s3api list-objects --bucket s3-to-bigquery
awslocal s3api get-bucket-notification-configuration --bucket s3-to-bigquery

awslocal sqs list-queues
awslocal sqs receive-message --queue-url http://sqs.us-east-1.localhost.localstack.cloud:4566/000000000000/s3-event-notification-queue

awslocal lambda get-function --function-name s3-to-bigquery

awslocal --endpoint-url=http://localhost:4566 logs tail /aws/lambda/s3-to-bigquery --follow
```

### Teardown

``` sh
# Just localstack.
awslocal s3 rm s3://s3-to-bigquery --recursive
(cd 005_unnecessary_dw_workloads/triplicating_data/before && terraform destroy --auto-approve)

# Everything.
make teardown
```