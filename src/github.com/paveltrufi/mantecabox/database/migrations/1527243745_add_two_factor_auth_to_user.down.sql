ALTER TABLE users
  DROP COLUMN two_factor_auth,
  DROP COLUMN two_factor_time;

DROP TRIGGER IF EXISTS set_2fa_timestamp
ON users;
DROP FUNCTION IF EXISTS trigger_set_2fa_timestamp();
