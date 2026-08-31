#!/bin/sh
set -eu

curl -fsSL 'https://raw.githubusercontent.com/artie-labs/artie-api-spec/refs/tags/v1.0.88/openapi.yaml' \
  | python3 ../../scripts/normalize-openapi-for-oapi-codegen.py \
  > /tmp/artie-api-openapi.yaml

go tool oapi-codegen \
  -config ../../oapi-codegen-config.yaml \
  -generate types,client,skip-prune \
  -o client.gen.go \
  /tmp/artie-api-openapi.yaml
