#! /bin/bash

if [ "$#" -ne 1 ]; then
    echo "Usage: $0 <token>"
    exit 1
fi

TOKEN="$1"

curl -X POST http://localhost:9205/api/users/upload \
  -H "AUTH_TOKEN: $TOKEN" \
  -F "file=@./test.txt"