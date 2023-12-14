# Before

### Infra

Services

``` sh
cp go.* 001_fragile_data_integrations/tightly_coupled_services/before/services/product
cp go.* 001_fragile_data_integrations/tightly_coupled_services/before/services/stock

(
  cd 001_fragile_data_integrations/tightly_coupled_services/before && \
  docker compose up --build --force-recreate -d
)
```

Convert to enterprise

``` sh
enterprise --url "postgres://root@localhost:26001/?sslmode=disable"
```

Connect to cluster

``` sh
cockroach sql --host localhost:26001 --insecure
```

Create product changefeed

``` sql
SET CLUSTER SETTING kv.rangefeed.enabled = true;

CREATE CHANGEFEED INTO 'kafka://redpanda:29092?topic_name=products'
AS SELECT
  "id",
  "name"
FROM product.products;
```

### Test infra

**Wait for 20s**

``` sh
curl -s "http://localhost:3001/products/ac9384f7-12f7-4431-8a78-c9ccc6d321af" | jq
curl -s "http://localhost:3002/stock/ac9384f7-12f7-4431-8a78-c9ccc6d321af" | jq
```

Update source products and see stock products updated.

``` sql
UPDATE product.products
SET "name" = 'new a'
WHERE id = 'ac9384f7-12f7-4431-8a78-c9ccc6d321af';
```

# After

### Infra

``` sh
cp go.* 001_fragile_data_integrations/tightly_coupled_services/after/services/product
cp go.* 001_fragile_data_integrations/tightly_coupled_services/after/services/stock

(
  cd 001_fragile_data_integrations/tightly_coupled_services/after && \
  docker compose up --build --force-recreate -d
)
```

### Test infra

**Wait for 20s**

``` sh
curl -s "http://localhost:3001/products/ac9384f7-12f7-4431-8a78-c9ccc6d321af" | jq
curl -s "http://localhost:3002/stock/ac9384f7-12f7-4431-8a78-c9ccc6d321af" | jq
```

### Summary

* There is tight coupling between microservices in the after scenario.

# Teardown

``` sh
make teardown
```