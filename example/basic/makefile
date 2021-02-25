#
# Local development only. This builds and executes the service in a local process.
#
run-service: frodo
	go run calc/cmd/main.go

run-client:
	go run main.go

frodo:
	frodo gateway calc/calculator_service.go && \
	frodo client  calc/calculator_service.go
