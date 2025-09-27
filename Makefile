# This file uses go run to execute mockery, which ensures that the version
# specified in tools.go and go.mod is used, providing reproducible builds.

.PHONY: mocks generate build dev deploy-infra

# All directories under /cmd
LAMBDA_DIRS := $(shell find cmd -mindepth 1 -maxdepth 1 -type d)

mocks:
	@echo "Generating mocks..."
	@go run github.com/vektra/mockery/v2 --name=Storage --dir=./pkg/storage --output=./pkg/storage/mocks --outpkg=mocks --case=underscore
	@go run github.com/vektra/mockery/v2 --name=CronScheduler --dir=./pkg/scheduler --output=./pkg/scheduler/mocks --outpkg=mocks --case=underscore
	@go run github.com/vektra/mockery/v2 --name=DynamoDBAPI --dir=./pkg/storage/dynamodb --output=./pkg/storage/dynamodb/mocks --outpkg=mocks --case=underscore
	@echo "Mocks generated successfully."

#########################
### Deploying via SAM ###
#########################

build:
	@echo "Building SAM application..."
	sam build

dev: build
	@echo "Starting reflex for watching Go files..."
	@(export AWS_PROFILE=default && reflex -r "\.go$$" -R "\.aws-sam" -- sh -c "sam build") &
	@echo "Starting local API using AWS_PROFILE=default..."
	@(export AWS_PROFILE=default && sam local start-api)

# Deploy new infrastructure with SAM
deploy-infra: build
	@echo "Deploying infrastructure..."
	sam deploy --profile default --no-confirm-changeset

.PHONY: generate
generate:
	@echo "Generating server code from OpenAPI spec..."
	@go run -mod=mod github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --generate=types,chi-server --package=api -o pkg/api/server.gen.go api/spec.yaml
	@echo "Server code generated successfully."

.PHONY: help
help:
	@echo "Available commands:"
	@echo "  make mocks           - Generate mocks for interfaces"
	@echo "  make generate        - Generate server code from OpenAPI spec"
	@echo "  make build           - Build the SAM application"
	@echo "  make dev           - Run the API locally with hot-reloading"
	@echo "  make deploy-infra     - Deploy the stack using configuration from samconfig.toml"
	@echo "  make help            - Show this help message"