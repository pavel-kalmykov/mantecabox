CREATE TABLE IF NOT EXISTS files (
  id                BIGSERIAL,
  created_at        TIMESTAMP   NOT NULL DEFAULT NOW(),
  updated_at        TIMESTAMP   NOT NULL DEFAULT NOW(),
  deleted_at        TIMESTAMP,
  name              VARCHAR     NOT NULL,
  owner             VARCHAR(20) NOT NULL CONSTRAINT files_owner_fk REFERENCES users ON DELETE CASCADE,
  "group"           VARCHAR     NOT NULL,
  user_readable     BIT         NOT NULL,
  user_writable     BIT         NOT NULL,
  user_executable   BIT         NOT NULL,
  group_readable    BIT         NOT NULL,
  group_writable    BIT         NOT NULL,
  group_executable  BIT         NOT NULL,
  other_readable    BIT         NOT NULL,
  other_writable    BIT         NOT NULL,
  other_executable  BIT         NOT NULL,
  platform_creation VARCHAR     NOT NULL
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
