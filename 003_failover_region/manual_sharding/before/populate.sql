INSERT INTO global.customer (id, email, country)
SELECT
  id,
  email,
  'uk'
FROM customer;
