ALTER TABLE users ADD name VARCHAR(32) NOT NULL DEFAULT '';
ALTER TABLE users ADD is_customer BOOLEAN NOT NULL DEFAULT false;