run: build
	out/post-service

frodo:
	../../out/frodo gateway post_service.go && \
	../../out/frodo client  post_service.go && \
	../../out/frodo mock    post_service.go && \
	../../out/frodo docs    post_service.go

build: frodo
	go build -o out/post-service cmd/main.go

test:
	go test ./...
