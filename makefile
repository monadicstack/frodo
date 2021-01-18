build: clean-gen
	@ \
	go build -o out/frodoc main.go

clean-gen:
	@ \
	rm -f example/*.gen.*.go

example-gen: build
	@ \
	out/frodoc gateway --input=example/group_service.go && \
	out/frodoc client  --input=example/group_service.go --language=go && \
	out/frodoc client  --input=example/group_service.go --language=js

example-run: example-gen
	@ \
	go run example/cmd/main.go
