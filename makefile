build: clean-gen
	@ \
	go build -o out/frodo main.go

clean-gen:
	@ \
	rm -f example/*.gen.*.go

example-gen: build
	@ \
	out/frodo gateway --input=example/group_service.go && \
	out/frodo client  --input=example/group_service.go --language=go && \
	out/frodo client  --input=example/group_service.go --language=js

example-run: example-gen
	@ \
	go run example/cmd/main.go
