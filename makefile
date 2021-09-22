TEST_TIMEOUT=30s

#
# Builds the actual frodo CLI executable.
#
build:
	@ \
	go build -o out/frodo main.go

install: build
	@ \
 	echo "Overwriting go-installed version..." && \
 	cp out/frodo $$GOPATH/bin/frodo

#
# Uses the frodo CLI to build the gateway and clients used in our test suites to validate code generation. This will
# build the frodo tool locally and use that rather than using the 'go install'-ed one to ensure we're testing the
# updates made in this project.
#
generate-test-clients: build
	out/frodo gateway example/names/name_service.go && \
	out/frodo client example/names/name_service.go --language=go && \
	out/frodo client example/names/name_service.go --language=js && \
	out/frodo client example/names/name_service.go --language=dart

#
# Runs the all of the test suites for the entire Frodo module.
#
test: test-unit test-clients

#
# Runs the self-contained unit tests that don't require code generation or anything like that to run.
#
test-unit:
	@ \
	go test -count=1 -timeout $(TEST_TIMEOUT) -tags unit ./...

#
# Dog-foods the Frodo CLI to build clients for all of our out-of-the-box-supported languages and run them
# through test suites to make sure that they behave as expected.
#
test-clients: generate-test-clients install-deps-node
	@ \
	go test -count=1 -timeout $(TEST_TIMEOUT) -tags client ./...

#
# We don't check in node_modules/ so if you're running the JS/Node client tests, this will make sure that
# you have the 'node-fetch' module available when running the tests.
#
install-deps-node:
	@ \
	cd generate/testdata/js && \
	npm install
