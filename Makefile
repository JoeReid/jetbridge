.PHONY: generate-protos
generate-protos:
	@echo "Generating code..."
	cd proto && npx @bufbuild/buf generate
	go generate ./...
