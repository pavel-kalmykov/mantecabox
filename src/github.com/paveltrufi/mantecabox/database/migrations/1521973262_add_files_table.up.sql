CREATE TABLE IF NOT EXISTS files (
  id                BIGSERIAL,
  created_at        TIMESTAMP   NOT NULL DEFAULT NOW(),
  updated_at        TIMESTAMP   NOT NULL DEFAULT NOW(),
  deleted_at        TIMESTAMP,
  name              VARCHAR     NOT NULL,
  owner             VARCHAR(20) NOT NULL CONSTRAINT files_owner_fk REFERENCES users ON DELETE CASCADE,
  "group"           VARCHAR,
  user_readable     BOOLEAN     NOT NULL DEFAULT TRUE,
  user_writable     BOOLEAN     NOT NULL DEFAULT TRUE,
  user_executable   BOOLEAN     NOT NULL DEFAULT FALSE,
  group_readable    BOOLEAN     NOT NULL DEFAULT TRUE,
  group_writable    BOOLEAN     NOT NULL DEFAULT FALSE,
  group_executable  BOOLEAN     NOT NULL DEFAULT FALSE,
  other_readable    BOOLEAN     NOT NULL DEFAULT FALSE,
  other_writable    BOOLEAN     NOT NULL DEFAULT FALSE,
  other_executable  BOOLEAN     NOT NULL DEFAULT FALSE,
  platform_creation VARCHAR
);

/*cambiamos el nombre del trigger de usuarios para evitar confusión*/
ALTER TRIGGER set_timestamp
ON users
RENAME TO set_users_timestamp;

CREATE TRIGGER set_files_timestamp
  BEFORE UPDATE
  ON files
  FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();
