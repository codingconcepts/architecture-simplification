# Before

### Resources

* https://levelup.gitconnected.com/aws-run-an-s3-triggered-lambda-locally-using-localstack-ac05f03dc896

### Infra

Compose

``` sh
cp go.* 005_unnecessary_dw_workloads/triplicating_data/before/s3-to-bigquery

(
  cd 005_unnecessary_dw_workloads/triplicating_data/before && \
  docker compose up --build --force-recreate -d
)
```

Terraform

``` sh
(cd 005_unnecessary_dw_workloads/triplicating_data/before && terraform init)
(cd 005_unnecessary_dw_workloads/triplicating_data/before && terraform apply --auto-approve)
(cd 005_unnecessary_dw_workloads/triplicating_data/before && terraform destroy --auto-approve)
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
awslocal s3 rm s3://s3-to-bigquery --recursive
```