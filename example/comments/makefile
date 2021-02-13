run: build
	out/comment-service

frodo:
	../../out/frodo gateway comment_service.go && \
	../../out/frodo client  comment_service.go && \
	../../out/frodo mock    comment_service.go && \
	../../out/frodo docs    comment_service.go

build: frodo
	go build -o out/comment-service cmd/main.go

test:
	go test ./...
