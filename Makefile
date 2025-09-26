# This file uses go run to execute mockery, which ensures that the version
# specified in tools.go and go.mod is used, providing reproducible builds.

.PHONY: mocks

mocks:
	@echo "Generating mocks..."
	@go run github.com/vektra/mockery/v2 --name=Storage --dir=./pkg/storage --output=./pkg/storage/mocks --outpkg=mocks --case=underscore
	@go run github.com/vektra/mockery/v2 --name=Scheduler --dir=./pkg/scheduler --output=./pkg/scheduler/mocks --outpkg=mocks --case=underscore
	@echo "Mocks generated successfully."

.PHONY: help
help:
	@echo "Available commands:"
	@echo "  make mocks    - Generate mocks for interfaces"
	@echo "  make help     - Show this help message"
