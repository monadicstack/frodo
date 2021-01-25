run: build
	out/comment-service

frodo:
	frodo gateway --input=comment_service.go && \
	frodo client  --input=comment_service.go

build: frodo
	go build -o out/comment-service cmd/main.go

test:
	go test ./...
