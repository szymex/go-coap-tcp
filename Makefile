
test:
	GOCACHE=off go test ./coap
	GOCACHE=off go test ./example-server
	GOCACHE=off go test ./coap-cli

build:
	go build ./coap
	go build  -o bin/example-server ./example-server
	go build  -o bin/coap-cli ./coap-cli

run-example-server:
	go run ./example-server