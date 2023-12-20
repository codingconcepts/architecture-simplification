SET CLUSTER SETTING kv.rangefeed.enabled = true;

-- Sagas
CREATE TYPE saga_status AS ENUM ('in_progress', 'finished', 'cancelling', 'cancelled');
CREATE TYPE saga_step AS ENUM ('order', 'payment', 'reservation', 'shipment', 'finished');

CREATE TABLE "sagas" (
  "order_id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  "payment" DECIMAL NOT NULL,
  "products" JSON NOT NULL,
  "step" saga_step NOT NULL DEFAULT 'order',
  "status" saga_status NOT NULL DEFAULT 'in_progress',
  "failures" JSON NOT NULL
);

CREATE CHANGEFEED INTO 'kafka://redpanda:29092?topic_name=sagas'
WITH kafka_sink_config='{"Flush": {"MaxMessages": 1, "Frequency": "100ms"}, "RequiredAcks": "ONE" }'
AS
  SELECT
    "order_id",
    "payment",
    "products",
    "status",
    "step",
    "failures"
  FROM sagas
  WHERE NOT event_op() = 'delete';

-- Orders
CREATE TYPE order_status AS ENUM ('in_progress', 'cancelled');

CREATE TABLE "orders" (
  "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  "status" order_status NOT NULL DEFAULT 'in_progress',
  "ts" TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Payments
CREATE TYPE payment_status AS ENUM ('preauth', 'auth', 'cancelled');

CREATE TABLE payments (
  "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  "order_id" UUID NOT NULL,
  "amount" DECIMAL NOT NULL,
  "status" payment_status NOT NULL DEFAULT 'preauth',
  "ts" TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Inventory
CREATE TABLE stock (
  "product_id" UUID PRIMARY KEY NOT NULL,
  "quantity" INT NOT NULL
);

CREATE TABLE reservations (
  "order_id" UUID NOT NULL REFERENCES "orders"("id"),
  "product_id" UUID NOT NULL,
  "quantity" INT NOT NULL,
  "ts" TIMESTAMPTZ NOT NULL DEFAULT now(),

  PRIMARY KEY ("order_id", "product_id")
);

-- Fulfillment
CREATE TYPE shipment_status AS ENUM ('in_progress', 'cancelled', 'dispatched');

CREATE TABLE shipments (
  "order_id" UUID PRIMARY KEY NOT NULL,
  "status" shipment_status NOT NULL DEFAULT 'in_progress'
);

-- Testing
CREATE OR REPLACE FUNCTION check_order(o_id IN UUID) RETURNS RECORD LANGUAGE SQL AS $$
  SELECT
    s.step step,
    s.status saga_status,
    o.status order_status,
    p.amount payment_amount,
    p.status payment_status,
    array_agg(r.product_id) product,
    array_agg(r.quantity) quantity,
    sh.status shipment_status
  FROM sagas s
  LEFT JOIN orders o ON s.order_id = o.id
  LEFT JOIN payments p ON o.id = p.order_id
  LEFT JOIN reservations r ON o.id = r.order_id
  LEFT JOIN shipments sh ON o.id = sh.order_id
  WHERE s.order_id = o_id
  GROUP BY s.step, s.status, o.status, p.amount, p.status, sh.status;
$$;