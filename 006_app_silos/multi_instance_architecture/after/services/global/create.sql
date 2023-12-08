CREATE DATABASE store
  PRIMARY REGION "eu-central-1"
  REGIONS "us-east-1", "ap-northeast-1";

USE store;

CREATE TABLE products (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  market STRING NOT NULL,
  "crdb_region" CRDB_INTERNAL_REGION AS (
    CASE
      WHEN "market" IN ('de', 'es', 'uk') THEN 'eu-central-1'
      WHEN "market" IN ('mx', 'us') THEN 'us-east-1'
      WHEN "market" IN ('jp') THEN 'ap-northeast-1'
      ELSE 'eu-central-1'
    END
  ) STORED,
  name STRING NOT NULL
) LOCALITY REGIONAL BY ROW;

INSERT INTO products (market, name) VALUES
  ('de', 'Americano'),
  ('de', 'Cappuccino'),
  ('de', 'Latte'),
  ('es', 'Americano'),
  ('es', 'Cappuccino'),
  ('es', 'Latte'),
  ('uk', 'Americano'),
  ('uk', 'Cappuccino'),
  ('uk', 'Latte'),
  ('mx', 'Americano'),
  ('mx', 'Cappuccino'),
  ('mx', 'Latte'),
  ('us', 'Americano'),
  ('us', 'Cappuccino'),
  ('us', 'Latte'),
  ('jp', 'Americano'),
  ('jp', 'Cappuccino'),
  ('jp', 'Latte');


CREATE TABLE i18n(
  word STRING NOT NULL,
  lang STRING NOT NULL,
  translation STRING NOT NULL,
  
  PRIMARY KEY (word, lang),
  INDEX (lang) STORING (translation)
) LOCALITY GLOBAL;

INSERT INTO i18n (word, lang, translation) VALUES
  ('Americano', 'de', 'Americano'),
  ('Cappuccino', 'de', 'Cappuccino'),
  ('Latte', 'de', 'Latté'),
  ('Americano', 'en', 'Americano'),
  ('Cappuccino', 'en', 'Cappuccino'),
  ('Latte', 'en', 'Latte'),
  ('Americano', 'es', 'Americano'),
  ('Cappuccino', 'es', 'Capuchino'),
  ('Latte', 'es', 'Latté'),
  ('Americano', 'ja', 'アメリカーノ'),
  ('Cappuccino', 'ja', 'カプチーノ'),
  ('Latte', 'ja', 'ラテ');