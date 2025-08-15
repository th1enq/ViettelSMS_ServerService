swagger:
	@echo "Generating Swagger documentation..."
	@swag init -g cmd/server/main.go -o docs