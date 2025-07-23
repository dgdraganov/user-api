#! /bin/bash

if [ "$#" -ne 1 ]; then
    echo "Usage: $0 <user_id>"
    exit 1
fi

USER_ID="$1"

# if AUTH_TOKEN is empty return error
if [ -z "$AUTH_TOKEN" ]; then
    echo "Error: AUTH_TOKEN is not set. Please authenticate first."
    echo "source authenticate.sh <email> <password>"
    exit 1
fi

curl -X DELETE "http://localhost:9205/api/users/$USER_ID" \
    -H "AUTH_TOKEN: $AUTH_TOKEN" \
    -H "Content-Type: application/json" 
