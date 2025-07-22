#! /bin/bash

if [ "$#" -ne 2 ]; then
    echo "Usage: $0 <user_id> <auth_token>"
    exit 1
fi

USER_ID="$1"
TOKEN="$2"

curl -X GET "http://localhost:9205/api/users/$USER_ID" \
    -H "AUTH_TOKEN: $TOKEN"