#!/bin/sh

set -e

npx @redocly/cli lint openapi.yaml
npx @redocly/cli bundle openapi.yaml -o openapi_bundle.yaml
npx @redocly/cli build-docs openapi_bundle.yaml -o openapi.html
