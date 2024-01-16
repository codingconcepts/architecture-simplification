### Create

**2 terminal windows**

Infra

``` sh
(cd 005_unnecessary_dw_workloads/analytics_in_cockroachdb && docker compose up -d)
docker exec -it node1 cockroach init --insecure
docker exec -it node1 cockroach sql --insecure
```

Create table and populate

``` sql
CREATE TABLE customers (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email STRING UNIQUE NOT NULL
);

CREATE TABLE products (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name STRING NOT NULL,
  price DECIMAL NOT NULL
);

CREATE TABLE orders (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  customer_id UUID NOT NULL REFERENCES customers(id),
  ts TIMESTAMPTZ NOT NULL DEFAULT now(),
  total DECIMAL NOT NULL
);

CREATE TABLE order_items (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  order_id UUID NOT NULL REFERENCES orders(id),
  product_id UUID NOT NULL REFERENCES products(id),
  quantity INTEGER NOT NULL
);

CREATE TABLE payments (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  order_id UUID REFERENCES orders(id),
  ts TIMESTAMPTZ DEFAULT now(),
  amount DECIMAL NOT NULL
);
```

Generate data

``` sh
dg \
  -c 005_unnecessary_dw_workloads/analytics_in_cockroachdb/dg.yaml \
  -o 005_unnecessary_dw_workloads/analytics_in_cockroachdb/csvs \
  -i imports.sql

python3 \
  -m http.server 9090 \
  -d 005_unnecessary_dw_workloads/analytics_in_cockroachdb/csvs
```

Import data

``` sql
IMPORT INTO customers (
	id, email
)
CSV DATA (
    'http://host.docker.internal:9090/customers.csv'
)
WITH skip='1', nullif = '', allow_quoted_null;

IMPORT INTO products (
	id, name, price
)
CSV DATA (
    'http://host.docker.internal:9090/products.csv'
)
WITH skip='1', nullif = '', allow_quoted_null;

IMPORT INTO orders (
	id, customer_id, ts, total
)
CSV DATA (
    'http://host.docker.internal:9090/orders.csv'
)
WITH skip='1', nullif = '', allow_quoted_null;

IMPORT INTO order_items (
	id, order_id, product_id, quantity
)
CSV DATA (
    'http://host.docker.internal:9090/order_items.csv'
)
WITH skip='1', nullif = '', allow_quoted_null;

IMPORT INTO payments (
	order_id, id, ts, amount
)
CSV DATA (
    'http://host.docker.internal:9090/payments.csv'
)
WITH skip='1', nullif = '', allow_quoted_null;
```

Simulate transactional workload

``` sh
go run 005_unnecessary_dw_workloads/analytics_in_cockroachdb/main.go

k6 run 005_unnecessary_dw_workloads/analytics_in_cockroachdb/load.js

```

**DEBUG** Test transactional workload

``` sh
# Insert customer
curl "http://localhost:3000/customers" \
  -H 'Content-Type: application/json' \
  -d '{
    "id": "68b790f4-9527-4a51-b0fd-b530613f34a9",
    "email": "abc@gmail.com"
  }'

# Get products
curl "http://localhost:3000/products" | jq

# Insert order
curl "http://localhost:3000/orders" \
  -H 'Content-Type: application/json' \
  -d '{
    "id": "318c7a41-aacb-4166-9179-706d4e60de83",
    "customer_id": "68b790f4-9527-4a51-b0fd-b530613f34a9",
    "items": [
      {
        "id": "c4c12e8f-6dc3-48d3-86ea-93cb01ee63c0",
        "quantity": 48
      },
      {
        "id": "9dcfbcd2-1848-4b50-b3f5-73b09d70d5be",
        "quantity": 1
      }
    ],
    "total": 100
  }'
```

### Analytics

Setup

``` sql
CREATE ROLE analytics WITH login;
GRANT SELECT ON * TO analytics;

CREATE USER analytics_user;
GRANT analytics TO analytics_user;

ALTER ROLE analytics SET default_transaction_use_follower_reads = 'on';
ALTER ROLE analytics SET default_transaction_priority = 'low';
ALTER ROLE analytics SET default_transaction_read_only = 'on';
ALTER ROLE analytics SET statement_timeout = '10m';

-- Remove some payments for the analytics queries.
DELETE FROM payments p
WHERE true
ORDER BY random()
LIMIT 5;
```

``` sh
cockroach sql --url "postgres://analytics@localhost:26257/defaultdb?sslmode=disable" --insecure
```

Queries

``` sql
-- Fetch a customer and their orders.
SELECT
  c.email,
  o.id,
  o.total,
  oi.quantity,
  p.price
FROM customers c
LEFT JOIN orders o ON c.id = o.customer_id
LEFT JOIN order_items oi ON o.id = oi.order_id
LEFT JOIN products p ON oi.product_id = p.id
WHERE c.id = '0a3546b5-6ad3-49b2-b960-dc6958faca30'
ORDER BY c.id, o.id, oi.id;

-- Show user-specific variables.
SHOW TRANSACTION PRIORITY;

-- Orders by month.
SELECT
  date_trunc('month', ts)::DATE mth,
  count(*)
FROM orders 
GROUP BY date_trunc('month', ts) 
ORDER BY mth DESC
LIMIT 10;

-- Busiest months in history.
SELECT
  date_trunc('month', ts)::DATE mth,
  count(*)
FROM orders 
GROUP BY date_trunc('month', ts) 
ORDER BY count DESC
LIMIT 10;

-- Most profitable months in history.
SELECT
  date_trunc('month', o.ts) AS month,
  SUM(p.amount) AS monthly_revenue
FROM orders o
JOIN payments p ON o.id = p.order_id
GROUP BY month
ORDER BY monthly_revenue DESC
LIMIT 10;

-- Biggest spenders.
SELECT
  c.email,
  SUM(o.total) AS total_spend,
  COUNT(o.id) AS order_count,
  ROUND(SUM(o.total) / COUNT(o.id)) AS order_average
FROM customers c
JOIN orders o ON c.id = o.customer_id
GROUP BY c.email
ORDER BY total_spend DESC
LIMIT 10;

-- Biggest average spenders.
SELECT
  c.email,
  ROUND(AVG(o.total)) AS average_spend
FROM customers c
JOIN orders o ON c.id = o.customer_id
GROUP BY c.email
ORDER BY average_spend DESC
LIMIT 10;

-- Most popular products.
SELECT
  p.name AS product,
  SUM(oi.quantity) AS total_quantity_sold
FROM products p
JOIN order_items oi ON p.id = oi.product_id
GROUP BY p.name
ORDER BY total_quantity_sold DESC
LIMIT 10;

-- Least popular products.
SELECT
  p.name AS product,
  SUM(oi.quantity) AS total_quantity_sold
FROM products p
JOIN order_items oi ON p.id = oi.product_id
GROUP BY p.name
ORDER BY total_quantity_sold
LIMIT 10;

-- Idle customers.
SELECT
  c.email,
  MAX(o.ts) AS latest_order_date
FROM customers c
JOIN orders o ON c.id = o.customer_id
GROUP BY c.email
ORDER BY latest_order_date
LIMIT 10;

-- Orders pending payment.
SELECT
  o.id AS order_id,
  o.customer_id,
  o.total
FROM orders o
LEFT JOIN payments p ON o.id = p.order_id
WHERE p.id IS NULL
ORDER BY o.total DESC;
```

### Summary

* Follower reads won't interfere with other transactions or cause retries.
* Follower read transaction will always run without interruption, as they won't get pushed because of writes occurring mid query.

### Teardown

``` sh
make teardown
```