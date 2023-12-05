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

Generate license

``` sh
cockroach sql \
  --url "postgresql://root@localhost:26001/?sslmode=disable" \
  -e "SELECT crdb_internal.cluster_id()"

crl-lic -type "Evaluation" -org "Rob Test" -months 1 cdd0461c-219b-4642-b90f-9a8bb5dffc06
```

Connect to cluster

``` sh
cockroach sql --host localhost:26001 --insecure
```

Apply license

``` sql
SET CLUSTER SETTING cluster.organization = 'Rob Test';
SET CLUSTER SETTING enterprise.license = 'crl-0-ChDN0EYcIZtGQrkPmou13/wGEI+x4KwGGAIiCFJvYiBUZXN0';
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

# Teardown

``` sh
make teardown
```