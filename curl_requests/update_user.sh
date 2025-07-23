#! /bin/bash

if [ "$#" -ne 2 ]; then
    echo "Usage: $0 <first_name> <age>"
    exit 1
fi

FIRST_NAME="$1"
AGE="$2"


# if AUTH_TOKEN is empty return error
if [ -z "$AUTH_TOKEN" ]; then
    echo "Error: AUTH_TOKEN is not set. Please authenticate first."
    echo "source authenticate.sh <email> <password>"
    exit 1
fi

curl -X PUT "http://localhost:9205/api/users" \
    -H "AUTH_TOKEN: $AUTH_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "first_name": "'"$FIRST_NAME"'",
        "email": "mitko@example.com",
        "age": '"$AGE"'
    }'
