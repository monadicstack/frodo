# Frodo

`frodo` is a code generator and runtime library that helps
you write RPC-style (micro) services without actually writing
any network/transport related code at all. You write your
services with business logic only and `frodo` exposes them
over RPC/REST.

At its core, it does many of the same things that gRPC does
for you with a lot less hassle. `frodo` actually parses
your Go source code files to learn everything it needs in
order to identify/expose/invoke your services.

## Getting Started

Add `frodo` to your project using `go install`. This will
give you the code generating executable `frodo` as well
as the runtime libraries used to support the RPC for your services.

```shell
go install github.com/robsignorelli/frodo
```

## Your First Service

`frodo` doesn't use .proto files or any other archaic DSL
files to work its magic. If you follow a few idiomatic
practices for service development in your code, `frodo`
will "just work".

#### Step 1: Define Your Service
Your first step is to define a .go file that simply defines
the contract for your service; the interface as well as the
inputs/outputs.

```go
// [project]/greeter/contract.go
package greeter

import (
    "context"
)

type GreeterService interface {
    SayHello(context.Context, *SayHelloRequest) (*SayHelloResponse, error)
    SayGoodbye(context.Context, *SayGoodbyeRequest) (*SayGoodbyeResponse, error)
}

type SayHelloRequest struct {
    Name string
}

type SayHelloResponse struct {
    Text string
}

type SayGoodbyeRequest struct {
    Name string
}

type SayGoodbyeResponse struct {
    Text string
}
```

At this point you haven't defined *how* this service actually
does its work. You just described what operations are
available.

We actually have enough for `frodo` to generate
every RPC artifact for you right now. We want you, however,
to think about your service, not RPC, so go ahead and
implement your service to behave how you want.

```go
// [project]/greeter/server.go
package greeter

import (
    "context"
)

type GreeterServiceServer struct {}

func (svc GreeterServiceServer) SayHello(ctx context.Context, req *SayHelloRequest) (*SayHelloResponse, error) {
    return &SayHelloResponse{Text: "Hello" + req.Name}, nil
}

func (svc GreeterServiceServer) SayGoodbye(ctx context.Context, req *SayGoodbyeRequest) (*SayGoodbyeResponse, error) {
    return &SayGoodbyeResponse{Text: "Goodbye" + req.Name}, nil
}
```

Notice that there's nothing in your service regarding
RPC, JSON, HTTP, etc. It's just some simple logic that
accepts an input value and returns an output value.

#### Step 2: Generate Your `frodo` RPC Client and Gateway

At this point, you've just written the same code that you (hopefully)
would have written even if you weren't using `frodo`. Next,
we want to auto-generate the RPC/networking code that
will let you talk to instances of this service remotely.
Run these commands in a terminal from your project's root
directory:

```shell
# Notice we're feeding it the contract, not the server code.
frodo gateway --input=greeter/contract.go
frodo client  --input=greeter/contract.go --language=go
```
This will create the directory `greeter/gen/` which includes
two new .go files; one that will run the RPC service for 
your "live" instance and another client that will let you
make calls to that instance remotely.

### Step 3: Expose Your Greeter Service

At this point you have everything you need to run a Go
program that exposes your greeter service and another that
makes calls to it. Let's start by firing up a gateway service
to expose your service to the world.

```go
package main

import (
    "net/http"

    "github.com/your/project/greeter"
    "github.com/your/project/greeter/gen"
)
func main() {
    service := greeter.GreeterServiceServer{}
    gateway := greeterrpc.NewGreeterServiceGateway(service)
    http.ListenAndServe(":9000", gateway)
}
```
Seriously. That's the whole program.

Compile and run it, and your service is now ready
to be consumed. We'll use the Go client we generated in just
a moment, but you can try this out right now by simply
using curl:

```shell
curl -d '{"Name":"Rob"}' http://localhost:9000/GreeterService.SayHello
# {"Text": "Hello Rob")

curl -d '{"Name":"Rob"}' http://localhost:9000/GreeterService.SayGoodbye
# {"Text": "Goodbye Rob")
```

#### Step 4: Consume Your Greeter Service

Making raw HTTP calls is somewhat of a pain. You've got to
deal with JSON, status codes, and so much other noise. Let's
just use the strongly-typed client that `frodo` created
back in step 2.

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/your/project/greeter"
    "github.com/your/project/greeter/gen"
)

func main() {
    // The client also implements GreeterService, so just
    // make calls on it like you would if it were a local instance.
    client := greeterrpc.NewGreeterServiceClient("http://localhost:9000")

    ctx := context.Background()

    hello, err := client.SayHello(ctx, &greeter.SayHelloRequest{Name: "Rob"})
    if err != nil {
        log.Fatalf("aww nuts: %v\n", err)
    }
    fmt.Printf("SayHello() -> %s\n", hello.Text)

    goodbye, err := client.SayGoodbye(ctx, &greeter.SayGoodbyeRequest{Name: "Rob"})
    if err != nil {
        log.Fatalf("aww nuts: %v\n", err)
    }
    fmt.Printf("SayGoodbye() -> %s\n", goodbye.Text)
}
```

Compile/run this program, and you should see the following output:

```
SayHello() -> Hello Rob
SayGoodbye() -> Goodbye Rob
```
That's it! Just write some idiomatic Go services and `frodo`
takes care of the mucky muck of dealing with marshaling,
transports, error handling, and everything else that makes
distributed services tricky.

## Why Not Just Use gRPC?

Coming soon...

## Creating a JavaScript Client

Coming soon...

## API Versioning

Coming soon...

## RESTful Endpoints

Coming soon...

## Middleware

Coming soon...

