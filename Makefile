build:
	go build

dev: 
	go run main.go

test:
	go test -v -cover ./...

start:
	sudo /usr/local/go/bin/go run main.go
# make test-server
# run test server at 8081, 8082, 8083
test-server:
	go run example_server/backend/test_backend_server.go

test-client: 
	go run example_server/client/test_client.go

.PHONY: build dev test-server test-client test