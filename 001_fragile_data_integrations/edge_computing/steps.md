# Before

### Infra

``` sh
(
  cd 001_fragile_data_integrations/edge_computing/before && \
  docker compose up --build --force-recreate -d
)
```

### Run

Connect to the primary node

``` sh
psql postgres://user:password@localhost:5432/postgres 
```

Create table and insert data

``` sql
CREATE TABLE i18n (
  "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  "word" VARCHAR(255) NOT NULL,
  "language" VARCHAR(255) NOT NULL,
  "translation" VARCHAR(255) NOT NULL
);

INSERT INTO i18n ("word", "language", "translation") VALUES
  ('Madagascar Hissing Cockroach', 'en', 'Madagascar Hissing Cockroach'),
  ('Giant Burrowing Cockroach', 'en', 'Giant Burrowing Cockroach'),
  ('Death''s Head Cockroach', 'en', 'Death''s Head Cockroach'),

  ('Madagascar Hissing Cockroach', 'de', 'Zischende Kakerlake aus Madagaskar'),
  ('Giant Burrowing Cockroach', 'de', 'Riesige grabende Kakerlake'),
  ('Death''s Head Cockroach', 'de', 'Totenkopfschabe'),

  ('Madagascar Hissing Cockroach', 'es', 'Cucaracha Silbadora de Madagascar'),
  ('Giant Burrowing Cockroach', 'es', 'Cucaracha excavadora gigante'),
  ('Death''s Head Cockroach', 'es', 'Cucaracha cabeza de muerte'),

  ('Madagascar Hissing Cockroach', 'ja', 'マダガスカルのゴキブリ'),
  ('Giant Burrowing Cockroach', 'ja', '巨大な穴を掘るゴキブリ'),
  ('Death''s Head Cockroach', 'ja', '死の頭のゴキブリ');
```

Check replication

``` sh
# US data
psql "postgres://user:password@localhost:5433/postgres" \
  -c "SELECT * FROM i18n"

# JP data
psql "postgres://user:password@localhost:5434/postgres" \
  -c "SELECT * FROM i18n"
```

### Summary

* Eventually consistent for US and JP users
* US and JP users have write to the EU
* No control over what gets replicated and what doesn't
* Will have to partition tables to achieve data residency

# After

### Infra

``` sh
(
  cd 001_fragile_data_integrations/edge_computing/after && \
  docker compose up --build --force-recreate -d
)
```

### Run

Initialise the cluster

``` sh
docker exec -it crdb_eu cockroach init --insecure
docker exec -it crdb_eu cockroach sql --insecure 
```

Get cluster id

``` sql
SELECT crdb_internal.cluster_id();
```

Generate license

``` sh
crl-lic -type "Evaluation" -org "Rob Test" -months 1 94c069d8-be49-49c0-839c-bece1b05b9b2
```

Apply license

``` sql
SET CLUSTER SETTING cluster.organization = 'Rob Test';
SET CLUSTER SETTING enterprise.license = 'crl-0-ChCUwGnYvklJwIOcvs4bBbmyEOqAha0GGAIiCFJvYiBUZXN0';
```

Create table and insert data

``` sql
CREATE DATABASE store
  PRIMARY REGION "eu-central-1"
  REGIONS "us-east-1", "ap-northeast-1";

CREATE TABLE store.i18n (
  "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  "word" STRING NOT NULL,
  "language" STRING NOT NULL,
  "translation" STRING NOT NULL
) LOCALITY GLOBAL;

INSERT INTO store.i18n ("word", "language", "translation") VALUES
  ('Madagascar Hissing Cockroach', 'en', 'Madagascar Hissing Cockroach'),
  ('Giant Burrowing Cockroach', 'en', 'Giant Burrowing Cockroach'),
  ('Death''s Head Cockroach', 'en', 'Death''s Head Cockroach'),

  ('Madagascar Hissing Cockroach', 'de', 'Zischende Kakerlake aus Madagaskar'),
  ('Giant Burrowing Cockroach', 'de', 'Riesige grabende Kakerlake'),
  ('Death''s Head Cockroach', 'de', 'Totenkopfschabe'),

  ('Madagascar Hissing Cockroach', 'es', 'Cucaracha Silbadora de Madagascar'),
  ('Giant Burrowing Cockroach', 'es', 'Cucaracha excavadora gigante'),
  ('Death''s Head Cockroach', 'es', 'Cucaracha cabeza de muerte'),

  ('Madagascar Hissing Cockroach', 'ja', 'マダガスカルのゴキブリ'),
  ('Giant Burrowing Cockroach', 'ja', '巨大な穴を掘るゴキブリ'),
  ('Death''s Head Cockroach', 'ja', '死の頭のゴキブリ');
```

Check replication

``` sh
# US data
cockroach sql --url "postgres://root@localhost:26002/store?sslmode=disable" \
  -e "SELECT * FROM i18n"

# JP data
cockroach sql --url "postgres://root@localhost:26003/store?sslmode=disable" \
  -e "SELECT * FROM i18n"
```