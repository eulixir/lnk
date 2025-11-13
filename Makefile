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