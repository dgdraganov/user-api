#!/bin/bash

if [ "$#" -ne 2 ]; then
    echo "Usage: $0 <email> <password>"
    exit 1
fi

EMAIL="$1"
PASSWORD="$2"

# Make the curl request with arguments
curl -X POST 'localhost:9205/api/auth' \
    -H "Content-Type: application/json" \
    -d '{
        "email": "'"$EMAIL"'",
        "password": "'"$PASSWORD"'"
    }'