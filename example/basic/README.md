# Basic Example

This is the `CalculatorService` example from the main README.
It's just an opportunity to see all the code required. If you
want to run this code you can do so from this `basic/` directory:

```shell
# In terminal 1
frodo gateway calc/calculator_service.go
frodo client  calc/calculator_service.go
go run calc/cmd/main.go

# In terminal 2
go run main.go
```

Or... this example also contains a `makefile` which does the same thing:

```shell
# In terminal 1
make run-server
# In terminal 2
make run-client
```

If everything is working properly, you should see the output:

```shell
5 + 2 = 7
5 - 2 = 3
```

## What This Example Shows

* The basic pattern for writing idiomatic services and their functions.
* How to run your service through `frodo` to generate your API and client.
* How to use your Frodo-generated artifacts to run and consume your service.
* You can get a working API without requiring any special configurations.
