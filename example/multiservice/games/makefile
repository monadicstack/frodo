#
# Local development only. This builds and executes the service in a local process.
#
run: build
	out/game

#
# Runs game_service.go through the 'frodo' code generator to spit out
# the latest and greatest RPC client/gateway code.
#
frodo:
	frodo gateway game_service.go && \
	frodo client  game_service.go

#
# If you add Frodo-based "//go:generate" comments to your service, generate your Frodo
# artifacts using that method instead.
#
generate:
	go generate .

#
# Rebuilds the binary for this service. We will "re-frodo" the service declaration beforehand
# so that any modifications to your service are always reflected in your client/gateway code
# without you having to think about it.
#
build: frodo
	go build -o out/game cmd/main.go

#
# This target hacks the Gibson; what do you think 'test' does? It runs through all of
# the test suites for this service.
#
test:
	go test ./...
