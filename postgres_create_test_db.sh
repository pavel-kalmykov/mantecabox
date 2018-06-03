#!/bin/bash

set -e
export TEST="_test"
psql  -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" --set=db_name="$POSTGRES_DB" --set=test_db_name="$POSTGRES_DB$TEST" --set=test_user="$POSTGRES_USER" <<-EOSQL
    CREATE DATABASE :test_db_name WITH TEMPLATE :db_name OWNER :test_user
EOSQL
