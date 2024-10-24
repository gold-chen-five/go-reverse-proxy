build:
	go build

dev: 
	go run main.go

test:
	go test -v -cover ./...

# make test-server port=8081
test-server:
	@echo "Running test server on port $(port)"
	go run example_server/backend/test_backend_server.go -port $(port)

test-client: 
	go run example_server/client/test_client.go

.PHONY: build dev test-server test-client test