#!/bin/bash

CONTAINER_NAME="sds-postgres"

echo "Stopping containers if they were already up..."
docker stop ${CONTAINER_NAME} || true && docker rm ${CONTAINER_NAME} || true

echo "Running Docker image..."
docker run --name sds-postgres \
         -e POSTGRES_PASSWORD=sds \
         -e POSTGRES_USER=sds \
         -e POSTGRES_DB=sds \
         -d -p 5432:5432 \
         -v $PWD/postgresql.conf:/srv/postgresql.conf \
         -v $PWD/server.crt:/srv/server.crt \
         -v $PWD/server.key:/srv/server.key \
         -v $PWD/configure_postgres.sh:/docker-entrypoint-initdb.d/configure_postgres.sh \
         -v $PWD/postgres_create_test_db.sh:/docker-entrypoint-initdb.d/postgres_create_test_db.sh \
         postgres
echo "Done"
