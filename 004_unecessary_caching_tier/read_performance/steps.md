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
k6 run \
  --vus 50 \
  --duration 10s \
  004_unecessary_caching_tier/read_performance/load.js

# ~13,300/s
```

## After

DB performance tweaks

``` sh
docker exec -it node1 cockroach sql --insecure
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
k6 run \
  --vus 50 \
  --duration 10s \
  004_unecessary_caching_tier/read_performance/load.js

# ~8,500/s
```

## Teardown

``` sh
docker ps -aq | xargs docker rm -f
```