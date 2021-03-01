#
# This runs the calculator service in a local process exposing it on port 9000.
#
calculator-service: frodo
	go run calc/cmd/main.go

#
# This runs the sample consumer program that utilizes the service client to make
# requests to the calculator service. This will fail if you do not have 'calculator-service'
# already running.
#
run:
	go run main.go

#
# Generates all of our RPC gateway/client artifacts from the service definitions.
#
frodo:
	frodo gateway calc/calculator_service.go && \
	frodo client  calc/calculator_service.go
