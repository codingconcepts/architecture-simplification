# Before

**WARNING:** Might need to change this, as Bens thoughts are more geared towards the duplication of data into other services to be the anti-pattern.

### Infra

``` sh
cp go.* 001_fragile_data_integrations/tightly_coupled_services/before/aggregators/shopper
cp go.* 001_fragile_data_integrations/tightly_coupled_services/before/services/product
cp go.* 001_fragile_data_integrations/tightly_coupled_services/before/services/stock

(
  cd 001_fragile_data_integrations/tightly_coupled_services/before && \
  docker compose up --build --force-recreate -d
)
```

### Test infra

**Wait for 20s**

``` sh
curl -s "http://localhost:3001/products/ac9384f7-12f7-4431-8a78-c9ccc6d321af" | jq
curl -s "http://localhost:3002/stock/ac9384f7-12f7-4431-8a78-c9ccc6d321af" | jq

curl -s "http://localhost:3000/products/ac9384f7-12f7-4431-8a78-c9ccc6d321af" | jq
```

# After



# Teardown

``` sh
make teardown
```