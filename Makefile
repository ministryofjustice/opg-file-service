test: ## Run all test suites
	docker-compose --project-name file-service-test \
	-f docker-compose.yml -f docker-compose.test.yml \
	run --rm file_service_test make go-test
	docker-compose --project-name file-service-test down

go-test:
	go mod download
	gotestsum --format short-verbose -- -coverprofile=../cover.out ./...

swagger-generate: # Generate API swagger docs from inline code annotations using Go Swagger (https://goswagger.io/)
	GO111MODULE=off swagger generate spec -o ./swagger.yaml --scan-models

swagger: # Serve swagger API docs on port 8383
	docker-compose --project-name file-service-docs \
    -f docker-compose.yml up -d --force-recreate swagger

docs: # Alias for make swagger (Generate API swagger docs)
	make swagger