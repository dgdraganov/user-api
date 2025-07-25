#! /bin/bash

if [ "$#" -ne 1 ]; then
    echo "Usage: $0 <file_path>"
    exit 1
fi

# if AUTH_TOKEN is empty return error
if [ -z "$AUTH_TOKEN" ]; then
    echo "Error: AUTH_TOKEN is not set. Please authenticate first."
    echo "source authenticate.sh <email> <password>"
    exit 1
fi

FILE_PATH="$1"

curl -X POST http://localhost:9205/api/users/file \
    -H "AUTH_TOKEN: $AUTH_TOKEN" \
    -F "file=@$FILE_PATH"