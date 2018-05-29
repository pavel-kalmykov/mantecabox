CREATE TABLE login_attempts (
  id         BIGSERIAL PRIMARY KEY,
  created_at TIMESTAMP DEFAULT NOW() NOT NULL,
  "user"     VARCHAR(40)             NOT NULL,
  user_agent VARCHAR,
  ipv4       VARCHAR(15),
  ipv6       VARCHAR(39),
  successful BOOLEAN                 NOT NULL,
  CONSTRAINT table_name_users_email_fk FOREIGN KEY ("user") REFERENCES users (email) ON DELETE CASCADE ON UPDATE CASCADE
);
