#!/bin/sh

SCRIPT="[wait-for-postgres.sh]"
EXPECTED="1-"
export PGPASSWORD=root
POSTGRES_STATUS=$(psql -h 127.0.0.1 -tA -U root -c "SELECT 1" | tr '\n' '-')

# TODO ADD TIMEOUT

until [ "$POSTGRES_STATUS" = "$EXPECTED" ]; do
    POSTGRES_STATUS=$(psql -h 127.0.0.1 -tA -U root -c "SELECT 1" | tr '\n' '-')
    echo "$SCRIPT polling postgres status, expected: $EXPECTED received: $POSTGRES_STATUS"
    sleep 1
done

echo "$SCRIPT Postgres Initialized"