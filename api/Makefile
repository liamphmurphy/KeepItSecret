.PHONY: generate test

generate:
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.5.0 -config oapi-codegen.yaml openapi/openapi.yaml

test: generate
	go test ./...
