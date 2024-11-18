build:
	go build

dev: 
	go run main.go

dev-ssl: 
	go run main.go --ssl

test:
	go test -v -cover ./...

start:
	sudo /usr/local/go/bin/go run main.go

start-ssl:
	sudo /usr/local/go/bin/go run main.go --ssl
	
# make test-server
# run test server at 8081, 8082, 8083
test-server:
	go run example_server/backend/test_backend_server.go

test-client: 
	go run example_server/client/test_client.go

test-client-ssl:
	go run example_server/client/test_client.go --ssl

.PHONY: build dev dev-ssl test-server test-client test-client-ssl test