BINARY := ./build
CMD := ./cmd/proxy/main.go
IMAGE := go-rate-limiter:latest

help:
	@echo "Available targets:"
	@echo "  help         - Show this help"
	@echo "  run          - Run the server with 'go run'"
	@echo "  build        - Build the production binary into $(BINARY)"
	@echo "  test         - Run unit tests"
	@echo "  race         - Run tests with the race detector"
	@echo "  fmt          - Run 'go fmt' on the project"
	@echo "  vet          - Run 'go vet'"
	@echo "  lint         - Run golangci-lint if installed (optional)"
	@echo "  docker-build - Build a docker image ($(IMAGE))"
	@echo "  docker-run   - Run the docker image with UPSTREAM_URL env (example)"
	@echo "  clean        - Remove build output"

dev:
	APP_ENV=development air

build:
	@echo "Building binary -> $(BINARY)"
	@mkdir -p $(BINARY)
	go build -o $(BINARY) $(CMD)

test:
	@echo "Running tests"
	go test ./... -v

race:
	@echo "Running tests with race detector"
	go test ./... -race -v

fmt:
	@echo "gofmt"
	go fmt ./...

vet:
	@echo "go vet ./..."
	go vet ./...

docker-build:
	@echo "Building docker image $(IMAGE)"
	docker build -t $(IMAGE) .

docker-run:
	@echo "Running docker image $(IMAGE) (example)"
	@echo "Make sure to set UPSTREAM_URL to your upstream service"
	docker run --rm -e UPSTREAM_URL="http://host.docker.internal:9000" -p 8080:8080 $(IMAGE)