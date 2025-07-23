#!/bin/bash

if [ "$#" -ne 2 ]; then
    echo "Usage: $0 <email> <password>"
    exit 1
fi

EMAIL="$1"
PASSWORD="$2"

# Make the curl request with arguments
RESPONSE=$(curl -X POST 'localhost:9205/api/auth' \
    -H "Content-Type: application/json" \
    -d '{
        "email": "'"$EMAIL"'",
        "password": "'"$PASSWORD"'"
    }')

echo $RESPONSE

# {"token":"eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6Im1pdGtvQGV4YW1wbGUuY29tIiwiZXhwIjoxNzUzMzYzNTEzLCJpYXQiOjE3NTMyNzcxMTMsInN1YiI6ImJjOTgyYWUyLTM5ZmEtNDkyYi1hMzhlLTU0OWYxMDFiMGRiMCJ9.mczf22UutOYIv0Wscbh3CSfgoTLkvhPNHvmgpFajmDb8yB2JHZhiNbVMTEbl6fNxqMVi8ne7qmU_ICARsqp0qQ"}
# get the token from the response
export AUTH_TOKEN=$(echo $RESPONSE | jq -r '.token')