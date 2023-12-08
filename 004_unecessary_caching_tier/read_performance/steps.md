## Shared

Infra

``` sh
(cd 004_unecessary_caching_tier/read_performance && docker compose up -d)

docker exec -it node1 cockroach init --insecure
docker exec -it node1 cockroach sql --insecure
```

Create table and populate

``` sql
CREATE TABLE stock (
  product_id UUID PRIMARY KEY,
  quantity INT NOT NULL
);

INSERT INTO stock (product_id, quantity)
  SELECT
    gen_random_uuid(),
    1000
  FROM generate_series(1, 1000);
```

Copy ids to clipboard

``` sh
cockroach sql \
  --insecure \
  -e "SELECT json_build_object('ids', array_agg(product_id)) FROM stock" \
  | sed -n 's/.*\[\([^]]*\)\].*/\1/p' \
  | sed 's/""/"/g' \
  | sed 's/^/[ /; s/$/ ]/' \
  > 004_unecessary_caching_tier/read_performance/ids.json
```

## Before

App

``` sh
go run 004_unecessary_caching_tier/read_performance/before/main.go
```

Load

``` sh
k6 run 004_unecessary_caching_tier/read_performance/load.js \
  --summary-trend-stats="min,max,p(95)"

# min=716µs    avg=3.18ms  max=71.27ms p(95)=5.96ms
```

## After

DB performance tweaks

``` sh
docker exec -it node1 cockroach sql --insecure
```

Get cluster id

``` sql
SELECT crdb_internal.cluster_id();
```

Generate license

``` sh
crl-lic -type "Evaluation" -org "Rob Test" -months 1 2ecc21ea-0cb2-444a-bd86-ea23e2ae6749
```

Apply license

``` sql
SET CLUSTER SETTING cluster.organization = 'Rob Test';
SET CLUSTER SETTING enterprise.license = 'crl-0-ChAuzCHqDLJESr2G6iPirmdJELrn66wGGAIiCFJvYiBUZXN0';
```

``` sql
SET CLUSTER SETTING kv.range_split.load_qps_threshold = 1000;

-- OR

ALTER TABLE stock SPLIT AT
SELECT rpad(to_hex(prefix::INT), 32, '0')::UUID AS split_at
FROM generate_series(0, 255) AS prefix;
```

App

``` sh
go run 004_unecessary_caching_tier/read_performance/after/main.go
```

Load

``` sh
k6 run 004_unecessary_caching_tier/read_performance/load.js \
  --summary-trend-stats="min,max,p(95)"

# min=1.64ms avg=7.59ms  max=47.47ms p(95)=14.67ms
```

## Teardown

``` sh
docker ps -aq | xargs docker rm -f
```