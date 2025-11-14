path := gateways/gocql/migrations
name := $(NAME)

.PHONY: generate-migration
generate-migration:
	@migrate create -ext sql -dir $(path) -seq $(name)

.PHONY: swagger
swagger:
	@swag init -g gateways/http/server.go -o docs

.PHONY: generate
generate:
	@make swagger
	@PATH="$$(go env GOPATH)/bin:$$PATH" mockery --all
	
.PHONY: test
test:
	@go test -timeout 30s -run ./...
	
.PHONY: coverage
coverage:
	@go test ./... -coverprofile=coverage.out
		
.PHONY: lint
lint:
	@golangci-lint run ./... --fix
	
.PHONY: fmt
format:
	@gofmt -w .
	@gofumpt -w .