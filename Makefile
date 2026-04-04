include .env
export

export PROJECT_ROOT=$(shell pwd)

env-up:
	docker-compose up -d crypto-scanner-postgres

env-down:
	docker-compose down crypto-scanner-postgres