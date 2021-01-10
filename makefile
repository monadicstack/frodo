build: clean-gen
	@ \
	go build -o out/frodoc main.go

clean-gen:
	@ \
	rm -f example/*.gen.*.go

example-gen: build
	@ \
	out/frodoc example/group_service.go

example-run: example-gen
	@ \
	go run example/cmd/main.go
