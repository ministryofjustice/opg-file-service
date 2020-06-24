test: ## Run all test suites
	docker-compose --project-name file-service-test \
	-f docker-compose.yml -f docker-compose.test.yml \
	run --rm file_service_test make go-test
	docker-compose --project-name file-service-test down

go-test:
	go mod download
	gotestsum --format short-verbose -- -coverprofile=../cover.out ./...