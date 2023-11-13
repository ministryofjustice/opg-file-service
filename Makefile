all: build scan test down

up:
	docker compose up -d --build file-service

build:
	docker compose build file-service

test-results:
	mkdir -p -m 0777 test-results .gocache .trivy-cache

setup-directories: test-results

scan: setup-directories
	docker compose run --rm trivy image --format table --exit-code 0 311462405659.dkr.ecr.eu-west-1.amazonaws.com/file-service:latest
	docker compose run --rm trivy image --format sarif --output /test-results/trivy.sarif --exit-code 1 311462405659.dkr.ecr.eu-west-1.amazonaws.com/file-service:latest

test: setup-directories
	docker compose run --rm test-runner

swagger-generate: # Generate API swagger docs from inline code annotations using Go Swagger (https://goswagger.io/)
	docker compose run --rm swagger-generate

swagger-up docs: # Serve swagger API docs on port 8383
	docker compose up -d --force-recreate swagger-ui
	@echo "Swagger docs available on http://localhost:8383/"
down:
	docker compose down
