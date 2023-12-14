### Introduction

* Anectode on my caching experience.

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

Copy ids into a file

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

### Summary

* Possible to run run workloads directly against a database without a cache.
* Try historical reads before caching; it offers a big performance boost at the cost of slightly stale data (which you'll see from caching anyway).
* Additional application complexity.
* Additional network latency.
* Could use local in-memory caches but with multiple service instances, you need to orchestrate cache consistency across services.
* Consider caching only after careful consideration and load testing. Your environment and workload will dictate your requirements on caching.