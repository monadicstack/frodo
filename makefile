build: clean-gen
	@ \
	go build -o out/exposec main.go

clean-gen:
	@ \
	rm -f example/*.gen.*.go

example-gen: build
	@ \
	out/exposec example/group_service.go

example-run: example-gen
	@ \
	go run example/cmd/main.go
