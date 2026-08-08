-include .env

IMAGE_TAG ?= rvpay-go-clients:local

install-tools:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(shell go env GOPATH)/bin v1.64.8
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@1.26.0
	go install go.uber.org/mock/mockgen@v0.6.0

generate:
	@echo "Generating code..."
	go generate ./...

lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@1.64.8 run ./...

test:
	@echo "Running tests in all subdirectories..."
	@find . -name "*_test.go" -exec dirname {} \; | sort | uniq | xargs -I {} go test ./... --cover {}

build:
	@echo "Building the project..."
	docker build -f ./Dockerfile -t $(or $(full_image_tag),$(IMAGE_TAG)) ..

publish:
	@echo "Publishing the project..."
	docker push $(full_image_tag)

run:
	@echo "Running the project..."
	go run ./cmd/grpc-service/...

create-migration:
	migrate create -seq -ext sql -dir db/migrations $(name)

rundb:
	docker run --name clients-postgres -p $(DB_PORT):5432 -e POSTGRES_USER=$(DB_USER) -e POSTGRES_PASSWORD=$(DB_PASSWORD) -e POSTGRES_DB=$(DB_NAME) -d postgres:16-alpine