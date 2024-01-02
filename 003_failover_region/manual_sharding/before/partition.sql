SELECT start_metadata_sync_to_node('eu_db', 5432);
UPDATE pg_dist_node SET noderack = 'eu' WHERE nodename = 'eu_db';
UPDATE pg_dist_node SET noderack = 'jp' WHERE nodename = 'jp_db';
UPDATE pg_dist_node SET noderack = 'us' WHERE nodename = 'us_db';


CREATE SCHEMA IF NOT EXISTS us;
CREATE SCHEMA IF NOT EXISTS eu;
CREATE SCHEMA IF NOT EXISTS jp;
CREATE SCHEMA IF NOT EXISTS global;


CREATE OR REPLACE FUNCTION match_schema_with_noderack(shard_id bigint, node_id int)
RETURNS BOOL LANGUAGE plpgsql AS $$
DECLARE
  table_id regclass;
  schema_name TEXT;
  node_location TEXT;
BEGIN
  -- find the schema name
  SELECT nspname INTO schema_name
  FROM pg_dist_shard s 
  JOIN pg_class c ON (s.logicalrelid = c.oid) 
  JOIN pg_namespace n ON (c.relnamespace = n.oid) 
  WHERE s.shardid = shard_id;
  
  -- schemas that do not have a corresponding "noderack" are allowed anywhere 
  IF NOT EXISTS (SELECT 1 FROM pg_dist_node WHERE noderack = schema_name) THEN
    RETURN true;
  END IF;

  -- noderack field is used to store node location
  SELECT noderack INTO node_location
  FROM pg_dist_node
  WHERE nodeid = node_id;
  
  -- allow schemas that match a noderack only on that noderack
  RETURN node_location = schema_name;
END; $$;

SELECT citus_add_rebalance_strategy('geo', 'citus_shard_cost_1', 'citus_node_capacity_1', 'match_schema_with_noderack', 0);

CREATE TABLE us.customer (
  id UUID NOT NULL DEFAULT gen_random_uuid(),
  email TEXT NOT NULL,
  country TEXT NOT NULL,

  PRIMARY KEY (id, country)
);
SELECT create_distributed_table('us.customer', 'country');

CREATE TABLE jp.customer (
  id UUID NOT NULL DEFAULT gen_random_uuid(),
  email TEXT NOT NULL,
  country TEXT NOT NULL,

  PRIMARY KEY (id, country)
);
SELECT create_distributed_table('jp.customer', 'country');

CREATE TABLE eu.customer (
  id UUID NOT NULL DEFAULT gen_random_uuid(),
  email TEXT NOT NULL,
  country TEXT NOT NULL,

  PRIMARY KEY (id, country)
);
SELECT create_distributed_table('eu.customer', 'country');

SELECT rebalance_table_shards('us.customer', rebalance_strategy := 'geo');
SELECT rebalance_table_shards('eu.customer', rebalance_strategy := 'geo');
SELECT rebalance_table_shards('jp.customer', rebalance_strategy := 'geo');


CREATE EXTENSION IF NOT EXISTS postgres_fdw;


CREATE SERVER us_server FOREIGN DATA WRAPPER postgres_fdw
  OPTIONS (host 'us_db', port '5432');

CREATE SERVER eu_server FOREIGN DATA WRAPPER postgres_fdw
  OPTIONS (host 'eu_db', port '5432');

CREATE SERVER jp_server FOREIGN DATA WRAPPER postgres_fdw
  OPTIONS (host 'jp_db', port '5432');


CREATE USER MAPPING FOR postgres
  SERVER us_server
  OPTIONS (user 'postgres');

CREATE USER MAPPING FOR postgres
  SERVER eu_server
  OPTIONS (user 'postgres');

CREATE USER MAPPING FOR postgres
  SERVER jp_server
  OPTIONS (user 'postgres');


CREATE TABLE global.customer (
  id UUID NOT NULL DEFAULT gen_random_uuid(),
  email TEXT NOT NULL,
  country TEXT NOT NULL
) PARTITION BY LIST (country);

-- Create foreign tables as partitions of the global table.
CREATE FOREIGN TABLE global.customer_jp
  PARTITION OF global.customer FOR VALUES IN ('jp', 'sg', 'zh', 'in')
  SERVER eu_server
  OPTIONS (schema_name 'jp', table_name 'customer');

CREATE FOREIGN TABLE global.customer_us
  PARTITION OF global.customer FOR VALUES IN ('us', 'mx', 'br', 'ca')
  SERVER eu_server
  OPTIONS (schema_name 'us', table_name 'customer');

CREATE FOREIGN TABLE global.customer_eu
  PARTITION OF global.customer DEFAULT
  SERVER eu_server
  OPTIONS (schema_name 'eu', table_name 'customer');