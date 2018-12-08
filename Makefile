
test:
	GOCACHE=off go test ./coap
	GOCACHE=off go test ./example-server

build:
	go build ./coap
	go build  -o bin/example-server ./example-server

run-example-server:
	go run ./example-server