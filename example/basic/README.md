# Basic Example

This is the `CalculatorService` example from the main README.
It's just an opportunity to see all the code required. If you
want to run this code you can do so from this `basic/` directory:

To run this example, just execute the following commands from
this directory. You can look at the `makefile` to see what commands
it's running under the hood... it's fairly simple.

```shell
# In terminal 1
make calculator-service

# In terminal 2
make run
```

If everything is working properly, you should see the output:

```
5 + 2 = 7
5 - 2 = 3
```

## What This Example Shows

* The basic pattern for writing idiomatic services and their functions.
* How to run your service through `frodo` to generate your API and client.
* How to use your Frodo-generated artifacts to run and consume your service.
* You can get a working API without requiring any special configurations.
