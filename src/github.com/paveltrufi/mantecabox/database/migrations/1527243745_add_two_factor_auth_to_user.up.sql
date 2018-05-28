ALTER TABLE users
  ADD COLUMN two_factor_auth CHAR(6),
  ADD COLUMN two_factor_time TIMESTAMP;

CREATE OR REPLACE FUNCTION trigger_set_2fa_timestamp()
  RETURNS TRIGGER AS $$
BEGIN
  IF NEW.two_factor_auth IS NOT NULL
  THEN
    NEW.two_factor_time := NOW();
  END IF;
  RETURN NEW;
END;
$$
LANGUAGE plpgsql;

CREATE TRIGGER set_2fa_timestamp
  BEFORE UPDATE
  ON users
  FOR EACH ROW
EXECUTE PROCEDURE trigger_set_2fa_timestamp();
