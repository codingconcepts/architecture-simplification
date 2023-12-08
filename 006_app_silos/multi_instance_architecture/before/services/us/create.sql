CREATE TABLE products (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(255) NOT NULL
);

INSERT INTO products (id, name) VALUES
  ('ac9384f7-12f7-4431-8a78-c9ccc6d321af', 'Americano'),
  ('bf6569f0-08fc-4a01-a3a6-2d353cdda01d', 'Cappuccino'),
  ('c9803ecd-04f2-44e4-87ff-e3e5725f93bd', 'Latte');


CREATE TABLE i18n(
  word VARCHAR(255) NOT NULL,
  lang VARCHAR(255) NOT NULL,
  translation VARCHAR(255) NOT NULL,
  
  PRIMARY KEY (word, lang)
);
CREATE INDEX ON i18n(lang) INCLUDE (translation);

INSERT INTO i18n (word, lang, translation) VALUES
  ('Americano', 'de', 'Americano'),
  ('Cappuccino', 'de', 'Cappuccino'),
  ('Latte', 'de', 'Latté'),
  ('Americano', 'en', 'Americano'),
  ('Cappuccino', 'en', 'Cappuccino'),
  ('Latte', 'en', 'Latte'),
  ('Americano', 'es', 'Americano'),
  ('Cappuccino', 'es', 'Capuchino'),
  ('Latte', 'es', 'Latté');