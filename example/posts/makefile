SERVICE_NAME=post
BINARY_NAME=$(SERVICE_NAME)-service

run: build
	out/$(BINARY_NAME)

frodo:
	../../out/frodo gateway --input=$(SERVICE_NAME)_service.go && \
	../../out/frodo client  --input=$(SERVICE_NAME)_service.go --language=go

build: frodo
	go build -o out/$(BINARY_NAME) cmd/main.go

test:
	go test ./...
