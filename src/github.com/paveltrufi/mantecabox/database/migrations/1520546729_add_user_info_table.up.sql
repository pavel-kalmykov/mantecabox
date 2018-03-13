CREATE TABLE userinfo
(
  uid        SERIAL                 NOT NULL,
  username   CHARACTER VARYING(100) NOT NULL,
  departname CHARACTER VARYING(500) NOT NULL,
  Created    DATE,
  CONSTRAINT userinfo_pkey PRIMARY KEY (uid)
) WITH (OIDS = FALSE
);
