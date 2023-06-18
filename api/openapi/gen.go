package openapi

//go:generate ./build_openapi.sh

//go:generate go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen --config config.yaml openapi_bundle.yaml
