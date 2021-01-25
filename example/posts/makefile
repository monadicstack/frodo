run: build
	out/post-service

frodo:
	frodo gateway --input=post_service.go && \
	frodo client  --input=post_service.go

build: frodo
	go build -o out/post-service cmd/main.go

test:
	go test ./...
