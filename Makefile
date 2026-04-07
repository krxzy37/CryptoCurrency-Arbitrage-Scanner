include .env
export

export PROJECT_ROOT=$(shell pwd)

env-up:
	@docker-compose up -d crypto-scanner-postgres

env-down:
	@docker-compose down crypto-scanner-postgres

env-cleanup:
	@read -p "clean volume files? risk of data loss. [y/n] " ans; \
	if [ "$$ans" = "y" ]; then \
	  docker-compose down crypto-scanner-postgres && \
	  rm -rf out/pgdata && \
	  echo "environment files have been cleared"; \
	else \
	 	echo "cleaning cancelled"; \
	fi
migrate-create:
	@if [-z "$(seq)"]; then \
  		echo "variable sec is not set. like: make migrate-create seq=init"; \
  		exit 1; \
  	fi;

	docker-compose run --rm crypto-scanner-postgres-migrate \
		create \
		-ext sql \
		-dir //migrations \
		-seq "$(seq)"

env-port-forward:
	@docker-compose up -d port-forwarder

env-port-close:
	@docker-compose down port-forwarder

migrate-up:
	@make migrate-action action=up

migrate-down:
	@make migrate-action action=down

migrate-action:
	@if [ -z "$(action)" ]; then \
		echo "variable action is not set. like: make migrate-action action=up"; \
		exit 1; \
	fi; \
	docker compose run --rm crypto-scanner-postgres-migrate \
		-path //migrations \
		-database "postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@crypto-scanner-postgres:5432/$(POSTGRES_DB)?sslmode=disable" \
		"$(action)"