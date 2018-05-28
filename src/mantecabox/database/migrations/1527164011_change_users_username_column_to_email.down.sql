ALTER TABLE users
  RENAME email TO username;
ALTER TABLE users
  ALTER COLUMN username TYPE varchar(40);
