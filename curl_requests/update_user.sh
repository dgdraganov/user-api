#! /bin/bash

if [ "$#" -ne 3 ]; then
    echo "Usage: $0 <user_id> <first_name> <age>"
    exit 1
fi

USER_ID="$1"
FIRST_NAME="$2"
AGE="$3"


# if AUTH_TOKEN is empty return error
if [ -z "$AUTH_TOKEN" ]; then
    echo "Error: AUTH_TOKEN is not set. Please authenticate first."
    echo "source authenticate.sh <email> <password>"
    exit 1
fi

curl -X PUT "http://localhost:9205/api/users/$USER_ID" \
    -H "AUTH_TOKEN: $AUTH_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "first_name": "'"$FIRST_NAME"'",
        "age": '"$AGE"'
    }'
