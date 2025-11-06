path := gateways/gocql/migrations
name := $(NAME)

.PHONY: generate-migration
generate-migration:
	@migrate create -ext sql -dir $(path) -seq $(name)
