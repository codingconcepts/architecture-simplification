#!/bin/bash
echo Wait for servers to be up
sleep 10

HOSTPARAMS="--host cockroachdb --insecure"
SQL="/cockroach/cockroach.sh sql $HOSTPARAMS"

# order_svc
$SQL -e "CREATE DATABASE orders;"
$SQL -e "CREATE TABLE orders.purchase(
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL, total DECIMAL NOT NULL,
  ts TIMESTAMPTZ NOT NULL DEFAULT now(),
  failures JSONB NOT NULL
);"
$SQL -e "CREATE USER orders_user;"
$SQL -e "GRANT ALL ON orders.* to orders_user;"

# payment_svc
$SQL -e "CREATE DATABASE payments;"
$SQL -e "CREATE TABLE payments.purchase(
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  order_id UUID NOT NULL,
  user_id UUID NOT NULL,
  balance DECIMAL NOT NULL,
  failures JSONB NOT NULL
);"
$SQL -e "CREATE USER payments_user;"
$SQL -e "GRANT ALL ON payments.* to payments_user;"

# inventory_svc
$SQL -e "CREATE DATABASE inventory;"
$SQL -e "CREATE TABLE inventory.stock(
  product_id UUID PRIMARY KEY NOT NULL,
  quantity INT NOT NULL,
  failures JSONB NOT NULL
);"
$SQL -e "CREATE USER inventory_user;"
$SQL -e "GRANT ALL ON inventory.* to inventory_user;"

# fulfillment_svc
$SQL -e "CREATE DATABASE fulfillment;"
$SQL -e "CREATE TABLE fulfillment.shipment(
  order_id UUID PRIMARY KEY NOT NULL,
  failures JSONB NOT NULL
);"
$SQL -e "CREATE USER fulfillment_user;"
$SQL -e "GRANT ALL ON fulfillment.* to fulfillment_user;"