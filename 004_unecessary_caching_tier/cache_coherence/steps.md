### Before

Read

``` mermaid
sequenceDiagram
    participant app
    participant cache
    participant database
    
    app->>cache: Get value
    cache-->>app: Does not exist
    app->>database: Get value
    database-->>app: Value
    app->>cache: Set value
```

Write

``` mermaid
sequenceDiagram
    participant app
    participant cache
    participant database
    
    app->>database: Set value
    app->>cache: Delete value
```

# Before simplification

### Create

Infrastructure

``` sh
(cd 004_unecessary_caching_tier/cache_coherence/before && docker compose up -d)
```

Connect to Postgres

``` sh
PGPASSWORD=password psql -h localhost -U postgres
```

Create table and populate

``` sql
CREATE TABLE stock (
  product_id VARCHAR(36) PRIMARY KEY,
  quantity INT NOT NULL
);

INSERT INTO stock (product_id, quantity) VALUES
  ('93410c29-1609-484d-8662-ae2d0aa93cc4', 1000),
  ('47b0472d-708c-4377-aab4-acf8752f0ecb', 1000),
  ('a1a879d8-58c0-4357-a570-a57c3b1fe059', 1000),
  ('5ded80d3-fb55-4a2f-b339-43fc9c89894a', 1000),
  ('b6afe0c5-9cab-4971-8c61-127fe5b4acd1', 1000),
  ('7098227b-4883-4992-bc32-e12335efbc8c', 1000);
```

### Run

``` sh
(cd 004_unecessary_caching_tier/cache_coherence/before && go run main.go -r 100ms -w 1s)
```

# After simplifcation

### Create

Infrastructure

``` sh
make teardown

(cd 004_unecessary_caching_tier/cache_coherence/after && docker compose -f compose.yaml up -d)
docker exec -it node1 cockroach init --insecure
docker exec -it node1 cockroach sql --insecure
```

Create table and populate

``` sql
CREATE TABLE stock (
  product_id UUID PRIMARY KEY,
  quantity INT NOT NULL
);

INSERT INTO stock (product_id, quantity) VALUES
  ('93410c29-1609-484d-8662-ae2d0aa93cc4', 1000),
  ('47b0472d-708c-4377-aab4-acf8752f0ecb', 1000),
  ('a1a879d8-58c0-4357-a570-a57c3b1fe059', 1000),
  ('5ded80d3-fb55-4a2f-b339-43fc9c89894a', 1000),
  ('b6afe0c5-9cab-4971-8c61-127fe5b4acd1', 1000),
  ('7098227b-4883-4992-bc32-e12335efbc8c', 1000);
```

### Run

``` sh
(cd 004_unecessary_caching_tier/cache_coherence/after && go run main.go)
```

### Todos

* Figure out why after scenario read and writes drift

### Summary

* Adding a cache ruins your ACID compliance
* Having a db and a cache introduces the dual write problem
* This demo showed the happy path; any comms issues to db or cache will also result in cache incoherence
* Having just a db means there's no dual write problem
* Having a value in the database that's higher than we're expecting, is OK, that just means there's more recent data than we're aware of

### Teardown

``` sh
make teardown
```