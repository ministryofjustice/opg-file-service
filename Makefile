all: build test

build:
	docker-compose build file_service

test: ## Run all test suites
	docker-compose --project-name file-service-test \
		-f docker-compose.yml -f docker-compose.test.yml \
		up -d localstack
	docker-compose --project-name file-service-test \
		-f docker-compose.yml -f docker-compose.test.yml \
		run --rm wait-for-it -address=localstack:4566 --timeout=30
	docker-compose --project-name file-service-test \
		-f docker-compose.yml -f docker-compose.test.yml \
		build file_service_test
	docker-compose --project-name file-service-test \
		-f docker-compose.yml -f docker-compose.test.yml \
		run --rm file_service_test make go-test
	docker-compose --project-name file-service-test down

go-test:
	go mod download
	gotestsum --junitfile unit-tests.xml --format short-verbose -- -coverprofile=./cover.out ./...

swagger-generate: # Generate API swagger docs from inline code annotations using Go Swagger (https://goswagger.io/)
	docker-compose --project-name file-service-docs-generate \
    -f docker-compose.yml run --rm swagger-generate
	docker-compose --project-name file-service-docs-generate down

swagger-up: # Serve swagger API docs on port 8383
	docker-compose --project-name file-service-docs \
    -f docker-compose.yml up -d --force-recreate swagger-ui

swagger-down:
	docker-compose --project-name file-service-docs down

docs: # Alias for make swagger-up (Generate API swagger docs)
	make swagger-up
