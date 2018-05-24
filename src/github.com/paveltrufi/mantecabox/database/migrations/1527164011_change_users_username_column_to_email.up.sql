ALTER TABLE users RENAME username TO email;
ALTER TABLE users ALTER COLUMN email TYPE varchar(40);
