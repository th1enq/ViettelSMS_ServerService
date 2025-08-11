#!/bin/bash

data=$(base64 servers.xlsx)

curl -X POST \
  'http://localhost/api/v1/servers/import' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d @- <<EOF
{
  "contentType": "text/xlsx",
  "data": "$data"
}
EOF
