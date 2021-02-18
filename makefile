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

#
# Runs the test suite for the module.
#
test:
	@ \
	go test -timeout 15s ./...

#
# Runs the test suite for the whole module, spitting out the the code coverage report to find gaps.
#
coverage:
	@ \
	go test -coverprofile=coverage.out -timeout 15s ./... && \
	go tool cover -func=coverage.out && \
	rm coverage.out
