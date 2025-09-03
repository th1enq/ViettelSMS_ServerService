swagger:
	@echo "Generating Swagger documentation..."
	@swag init -g cmd/server/main.go -o docs

COVER_CORE_PKGS=\
	./internal/usecase/server

COVERPKG_LIST=$(shell echo $(COVER_CORE_PKGS) | tr ' ' ',')

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -v -coverpkg=$(COVERPKG_LIST) $(COVER_CORE_PKGS) -coverprofile=coverage.out
	@go tool cover -func=coverage.out | tee coverage.txt
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
