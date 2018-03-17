/* Como esta va a ser la primera tabla con timestamps, nos creamos aquí la función que se usará en los triggers */
CREATE OR REPLACE FUNCTION trigger_set_timestamp()
  RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$
LANGUAGE plpgsql;

CREATE TABLE IF NOT EXISTS users (
  created_at TIMESTAMP   NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP   NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMP,
  username   VARCHAR(20)
    CONSTRAINT users_pk PRIMARY KEY
    CONSTRAINT username_not_empty CHECK (length(username) > 0),
  password   VARCHAR(60) NOT NULL
    CONSTRAINT password_not_empty CHECK (length(password) > 0)
);

CREATE TRIGGER set_timestamp
  BEFORE UPDATE
  ON users
  FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();
