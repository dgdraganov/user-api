#! /bin/bash

if [ "$#" -ne 5 ]; then
    echo "Usage: $0 <first_name> <last_name> <email> <age> <password>"
    exit 1
fi

FIRST_NAME="$1"
LAST_NAME="$2"
EMAIL="$3"
AGE="$4"
PASSWORD="$5"

curl -X POST "http://localhost:9205/api/users" \
    -H "Content-Type: application/json" \
    -d '{
        "first_name": "'"$FIRST_NAME"'",
        "last_name": "'"$LAST_NAME"'",
        "email": "'"$EMAIL"'",
        "age": '"$AGE"',
        "password": "'"$PASSWORD"'"
    }'