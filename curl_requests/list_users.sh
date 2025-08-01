#! /bin/bash


# Check if both arguments are provided
if [ "$#" -ne 2 ]; then
    echo "Usage: $0 <page> <page_size>"
    exit 1
fi

PAGE="$1"
PAGESIZE="$2"

# if AUTH_TOKEN is empty return error
if [ -z "$AUTH_TOKEN" ]; then
    echo "Error: AUTH_TOKEN is not set. Please authenticate first."
    echo "source authenticate.sh <email> <password>"
    exit 1
fi

# Make the curl request with arguments
curl -X GET "localhost:9205/api/users?page=$PAGE&page_size=$PAGESIZE" \
    -H "Content-Type: application/json" \
    -H "AUTH_TOKEN: $AUTH_TOKEN"



