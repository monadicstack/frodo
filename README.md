[![Go Report Card](https://goreportcard.com/badge/github.com/monadicstack/frodo)](https://goreportcard.com/report/github.com/monadicstack/frodo)

# Frodo

Frodo is a code generator and runtime library that helps
you write RPC-enabled (micro) services and APIs. It parses
the interfaces/structs/comments in your service code to 
generate all of your client/server communication code.

* No .proto files. Your services are just idiomatic Go code.
* Auto-generate APIs that play nicely with `net/http`, middleware, and other standard library compatible API solutions.  
* Auto-generate RPC-style clients in multiple languages like Go and JavaScript.
* Auto-generate strongly-typed mock implementations of your service for unit testing.
* Create OpenAPI documentation so others know how to interact with your API (if they can't use the client).

Frodo automates all the boilerplate associated with service
communication, data marshaling, routing, error handling, etc. You
get to focus on writing business logic and features while Frodo gives
you all of that other stuff to turn it into a distributed system for free.

Tools like gRPC solve similar problems by giving you a complex
airplane cockpit filled with knobs and dials most of us don't want/need.
Frodo is the autopilot button that gets most of us where we need to go
with as little fuss as possible.

## Table of Contents

* [Getting Started](https://github.com/monadicstack/frodo#getting-started)
* [Example](https://github.com/monadicstack/frodo#example)
* [Customize HTTP Route, Status, etc](https://github.com/monadicstack/frodo#doc-options-custom-urls-status-etc)
* [Error Handling](https://github.com/monadicstack/frodo#error-handling)
* [HTTP Redirects](https://github.com/monadicstack/frodo#http-redirects)
* [Middleware](https://github.com/monadicstack/frodo#middleware)
* [Request Scoped Metadata](https://github.com/monadicstack/frodo#request-scoped-metadata)
* [Create a JavaScript Client](https://github.com/monadicstack/frodo#creating-a-javascript-client)
* [Authorization](https://github.com/monadicstack/frodo#authorization)
* [Composing Gateways](https://github.com/monadicstack/frodo#composing-gateways)
* [Mocking Services](https://github.com/monadicstack/frodo#mocking-services)
* [Generating OpenAPI Documentation](https://github.com/monadicstack/frodo#generate-openapiswagger-documentation-experimental)
* [Go Generate Support](https://github.com/monadicstack/frodo#go-generate-support)
* [Bring Your Own Templates](https://github.com/monadicstack/frodo#bring-your-own-templates)
* [New Service Scaffolding](https://github.com/monadicstack/frodo#create-a-new-service-w-frodo-create)
* [Why Not gRPC?](https://github.com/monadicstack/frodo#why-not-just-use-grpc) (motivation for this project)

## Getting Started

*Frodo requires Go 1.16+ as it uses `fs.FS` and `//go:embed` to load templates.*

```shell
go install github.com/monadicstack/frodo
```
This will fetch the `frodo` code generation executable as well
as the runtime libraries that allow your services to
communicate with each other.


## Example

#### Step 1: Describe Your Service

Your first step is to write a .go file that just defines
the contract for your service; the interface as well as the
inputs/outputs.

```go
// calculator_service.go
package calc

import (
    "context"
)

type CalculatorService interface {
    Add(context.Context, *AddRequest) (*AddResponse, error)
    Sub(context.Context, *SubRequest) (*SubResponse, error)
}

type AddRequest struct {
    A int
    B int
}

type AddResponse struct {
    Result int
}

type SubRequest struct {
    A int
    B int
}

type SubResponse struct {
    Result int
}
```

You haven't actually defined *how* this service gets
this work done; just which operations are available.

We actually have enough for `frodo` to
generate your RPC/API code already, but we'll hold off
for a moment. Frodo frees you up to focus on building
features, so let's actually implement service; no networking,
no marshaling, no status stuff, just logic to make your
service behave properly.

```go
// calculator_service_handler.go
package calc

import (
    "context"
)

type CalculatorServiceHandler struct {}

func (svc CalculatorServiceHandler) Add(ctx context.Context, req *AddRequest) (*AddResponse, error) {
    result := req.A + req.B
    return &AddResponse{Result: result}, nil
}

func (svc CalculatorServiceHandler) Sub(ctx context.Context, req *SubRequest) (*SubResponse, error) {
    result := req.A - req.B
    return &SubResponse{Result: result}, nil
}
```

#### Step 2: Generate Your RPC Client and Gateway

At this point, you've just written the same code that you (hopefully)
would have written even if you weren't using Frodo. Next,
we want to auto-generate two things:

* A "gateway" that allows an instance of your CalculatorService
  to listen for incoming requests (via an HTTP API).
* A "client" struct that communicates with that API to get work done.

Just run these two commands in a terminal:

```shell
# Feed it the service interface code, not the handler.
frodo gateway calculator_service.go
frodo client  calculator_service.go
```

#### Step 3: Run Your Calculator API Server

Let's fire up an HTTP server on port 9000 that makes your service
available for consumption (you can choose any port you want, obviously).  

```go
package main

import (
    "net/http"

    "github.com/your/project/calc"
    calcrpc "github.com/your/project/calc/gen"
)

func main() {
    service := calc.CalculatorServiceHandler{}
    gateway := calcrpc.NewCalculatorServiceGateway(service)
    http.ListenAndServe(":9000", gateway)
}
```
Seriously. That's the whole program.

Compile and run it, and your service/API is now ready
to be consumed. We'll use the Go client we generated in just
a moment, but you can try this out right now by simply
using curl:

```shell
curl -d '{"A":5, "B":2}' http://localhost:9000/CalculatorService.Add
# {"Result":7}
curl -d '{"A":5, "B":2}' http://localhost:9000/CalculatorService.Sub
# {"Result":3}
```

#### Step 4: Consume Your Calculator Service

While you can use raw HTTP to communicate with the service,
let's use our auto-generated client to hide the gory
details of JSON marshaling, status code translation, and
other noise.

The client actually implements CalculatorService
just like the server/handler does. As a result the RPC-style
call will "feel" like you're executing the service work
locally, when in reality the client is actually making API
calls to the server running on port 9000.

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/your/project/calc"
    "github.com/your/project/calc/gen"
)

func main() {
    ctx := context.Background()
    client := calcrpc.NewCalculatorServiceClient("http://localhost:9000")

    add, err := client.Add(ctx, &calc.AddRequest{A:5, B:2})
    if err != nil {
        log.Fatalf(err.Error())
    }
    fmt.Println("5 + 2 = ", add.Result)

    sub, err := client.Sub(ctx, &calc.SubRequest{A:5, B:2})
    if err != nil {
        log.Fatalf(err.Error())
    }
    fmt.Println("5 - 2 = ", sub.Result)
}
```

Compile/run this program, and you should see the following output:

```
5 + 2 = 7
5 - 2 = 3
```
That's it!

For more examples of how to write services that let Frodo take
care of the RPC/API boilerplate, take a look in the [example/](https://github.com/monadicstack/frodo/tree/main/example)
directory of this repo.

## Doc Options: Custom URLs, Status, etc

Frodo gives you a remote service/API that "just works" out of the
box. You can, however customize the API routes for individual operations,
set a prefix for all routes in a service, and more using "Doc Options"...
worst Spider-Man villain ever.

Here's an example with all the available options. They are all
independent, so you can specify a custom status without specifying
a custom route and so on.

```go
// CalculatorService provides some basic arithmetic operations.
//
// VERSION 0.1.3
// PATH /v1
type CalculatorService interface {
    // Add calculates the sum of A + B.
    //
    // HTTP 202
    // GET /addition/:A/:B
    Add(context.Context, *AddRequest) (*AddResponse, error)

    // Sub calculates the difference of A - B.
    //
    // GET /subtraction/:A/:B
    Sub(context.Context, *SubRequest) (*SubResponse, error)
}
```

#### Service: PATH

This prepends your custom value on every route in the API. It applies
to the standard `ServiceName.FunctionName` routes as well as custom routes
as we'll cover in a moment. 

Your generated API and RPC clients will be auto-wired to use the prefix "v1" under the
hood, so you don't need to change your code any further. If you want
to hit the raw HTTP endpoints, however, here's how they look now:

```shell
curl -d '{"A":5, "B":2}' http://localhost:9000/v1/CalculatorService.Add
# {"Result":7}

curl -d '{"A":5, "B":2}' http://localhost:9000/v1/CalculatorService.Sub
# {"Result":3}
```

#### Service: VERSION

Annotate your service to track its current version. This doesn't affect
the behavior of your service in any way. It's currently only used if you
generate documentation using `frodo docs`

#### Function: GET/POST/PUT/PATCH/DELETE

You can replace the default `POST ServiceName.FunctionName` route for any
service operation with the route of your choice. In the example, the path parameters `:A` and `:B`
will be bound to the equivalent A and B attributes on the request struct.

Here are the updated curl calls after we generate the new
gateway code. Notice it's also taking into account the service's PATH
prefix as well:

```shell
curl http://localhost:9000/v1/addition/5/2
# {"Result":7}
curl http://localhost:9000/v1/subtraction/5/2
# {"Result":3}
```

#### Function: HTTP

This lets you have the API return a non-200 status code on success.
For instance, the Add function's route will return a "202 Accepted"
status when it responds with the answer instead of "200 OK".

## Error Handling

By default, if your service call returns a non-nil error, the
resulting RPC/HTTP request will have a 500 status code. You
can, however, customize that status code to correspond to the type
of failure (e.g. 404 when something was not found).

The easiest way to do this is to just use the `rpc/errors`
package when you encounter a failure case:

```go
import (
    "github.com/monadicstack/frodo/rpc/errors"
)

func (svc UserService) Get(ctx context.Context, req *GetRequest) (*GetResponse, error) {
    if req.ID == "" {
        return nil, errors.BadRequest("id is required")
    }
    user, err := svc.Repo.GetByID(req.ID)
    if err != nil {
    	return nil, err
    }
    if user == nil {
        return nil, errors.NotFound("user not found: %s", req.ID)
    }
    return &GetResponse{User: user}, nil
}
```

In this case, the caller will receive an HTTP 400 if they
didn't provide an id, a 404 if there is no user with that
id, and a 500 if any other type of error occurs.

While the error categories in Frodo's errors package is
probably good enough for most people, take a look at the
documentation for [github.com/monadicstack/respond](https://github.com/monadicstack/respond#how-does-it-know-which-4xx5xx-status-to-use)
to see how you can roll your own custom errors, but still
drive which 4XX/5XX status your service generates.

## HTTP Redirects

It's fairly common to have a service call that does some work
to locate a resource, authorize it, and then redirect to
S3, CloudFront, or some other CDN to actually serve up
the raw asset.

In Frodo, it's pretty simple. If your XxxResponse struct implements
the `respond.Redirector` interface from [github.com/monadicstack/respond](https://github.com/monadicstack/respond)
then the gateway will respond with a 307-style redirect
to the URL of your choice:

```go
// In video_service.go, this implements the Redirector interface.
type DownloadResponse struct {
    Bucket string
    Key    string	
}

func (res DownloadResponse) Redirect() string {
    return fmt.Sprintf("https://%s.s3.amazonaws.com/%s",
        res.Bucket,
        res.Key)
}


// In video_service_handler.go, this will result in a 307-style
// redirect to the URL returned by "response.Redirect()"
func (svc VideoServiceHandler) Download(ctx context.Context, req *DownloadRequest) (*DownloadResponse, error) {
    file := svc.Repo.Get(req.FileID)
    return &DownloadResponse{
        Bucket: file.Bucket,
        Key:    file.Key, 
    }, nil
}
```

## Middleware

Your RPC gateway is just an `http.Handler`, so you can plug
and play your favorite off-the-shelf middleware. Here's an
example using [github.com/urfave/negroni](https://github.com/urfave/negroni)

```go
func main() {
    service := calc.CalculatorServiceHandler{}
    gateway := calcrpc.NewCalculatorServiceGateway(service,
        rpc.WithMiddleware(
            negroni.NewLogger().ServeHTTP,
            NotOnMonday,
        ))

    http.ListenAndServe(":9000", gateway)
}

func NotOnMonday(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
    if time.Now().Weekday() == time.Monday {
        http.Error(w, "garfield says no math on mondays", 403)
        return
    }
    next(w, req)
}
```

You might think to yourself... wait a minute; I thought the gateway
*was* an HTTP handler, so couldn't I just wrap the gateway in middleware
like this?

```go
gateway := calcrpc.NewCalculatorServiceGateway(service)
handler := negroni.New(
    negroni.NewLogger(),
    negroni.WrapFunc(NotOnMonday),
)
handler.UseHandler(gateway)
http.ListenAndServe(":9000", handler)
```

You absolutely can, and it will work great... mostly. Frodo's gateway
performs a few book-keeping tasks before it executes your
middleware and the eventual service function. One of these tasks is
restoring request-scoped metadata headers passed from the caller (next section).
If you don't actually need that information, then this works. If you want
the full arsenal of Frodo functionality in your middleware functions,
be sure to use `.WithMiddleware()` like in the first example.

## Request Scoped Metadata

When you make an RPC call from Service A to Service B, none
of the values stored on the `context.Context` will be
available to you when are in Service B's handler. There are
instances, however, where it's useful to have values follow
every hop from service to service; request ids, tracing info, etc.

Frodo places a special bag of values called "metadata" onto
the context which **will** follow you as you go from service
to service:

```go
func (a ServiceA) Foo(ctx context.Context, r *FooRequest) (*FooResponse, error) {
    // "Hello" will NOT follow you when you call Bar(),
    // but "DontPanic" will. Notice that the metadata
    // value does not need to be a string like in gRPC.
    ctx = context.WithValue(ctx, "Hello", "World")
    ctx = metadata.WithValue(ctx, "DontPanic", 42)

    serviceB.Bar(ctx, &BarRequest{})
}

func (b ServiceB) Bar(ctx context.Context, r *BarRequest) (*BarResponse, error) {
    a, okA := ctx.Value("A").(string)

    b := 0
    okB = metadata.Value(ctx, "DontPanic", &b)
    
    // At this point:
    // a == ""   okA == false
    // b == 42   okB == true
}
```

If you're wondering why `metadata.Value()` looks more like
`json.Unarmsahl()` than `context.Value()`, it has to
do with a limitation of reflection in Go. When the values
are sent over the network from Service A to Service B, we
lose all type information. We need the type info `&b` gives
us in order to properly restore the original value, so Frodo
follows the idiom established by many
of the decoders in the standard library.

## Creating a JavaScript Client

The `frodo` tool can actually generate a JS client that you
can add to your frontend code to hide the complexity of making
all the API calls to your backend service. Without any plugins
or fuss, we can create a JS client of the same
CalculatorService from earlier...

```shell
frodo client calc/calculator_service.go --language=js
```

This will create the file `calculator_service.gen.client.js`
which you can include with your frontend codebase. Using it
should look similar to the Go client we saw earlier:

```js
import {CalculatorService} from 'lib/calculator_service.gen.client';

// The service client is a class that exposes all of the
// operations as 'async' functions that resolve with the
// result of the service call.
const service = new CalculatorService('http://localhost:9000')
const add = await service.Add({A:5, B:2})
const sub = await service.Sub({A:5, B:2})

// Should print:
// Add(5, 2) = 7
// Sub(5, 2) = 3
console.info('Add(5, 2) = ' + add.Result)
console.info('Sub(5, 2) = ' + sub.Result)
```

Another subtle benefit of using Frodo's client is that all of your
service/function documentation follows you in the generated code.
It's included in the JSDoc of the client so all of your service/API
documentation should be available to your IDE even when writing
your frontend code.

#### Node Support

You can actually use this client in your Node
server-side code as well to call service functions in your Go API.
The client uses the 'fetch' API to handle the HTTP layer. In the browser,
it will just use the window-scoped 'fetch' instance, but you can supply
your own to the constructor of your service client:

```js
const fetch = require('node-fetch');

// Just inject your 'fetch' implementation to the construtor and everything
// should work exactly the same.
const service = new CalculatorService('http://localhost:9000', {fetch})
const add = await service.Add({A:5, B:2})
const sub = await service.Sub({A:5, B:2})
```

## Authorization

Since you probably want your services to do some sort of authentication
and authorization, Frodo helps you manage the HTTP `Authorization` header.
When the gateway accepts a request, it stores the header in a context value
that you can fetch in middleware or your handler, so you can decide how
to utilize that info.

```go
import (
    "github.com/monadicstack/frodo/rpc/authorization"
    "github.com/monadicstack/frodo/rpc/errors"
)

func (a *ServiceA) Hello(ctx contex.Context, req *HelloRequest) (*HelloResponse, error) {
    auth := authorization.FromContext(ctx)
    if auth.Empty() {
        return nil, errors.BadCredentials("missing authorization header")
    }
    if auth.String() == "Donny" {
        return nil, errors.PermissionDenied("you're out of your element")	
    }
    return &HelloResponse{Text: "Hello"+req.Name}, nil
}
```
Ignore the awful security of hardcoding valid/invalid credentials;
the value for `auth` should be the value
of the `Authorization` header on the incoming HTTP request. The whole
idea is that your business logic exists independent of HTTP-related stuff,
so Frodo takes that HTTP-provided data and puts it on the context. This
allows your handler to deal with credentials in a transport-independent fashion.


Frodo, however, just makes sure that you have the value that was given. With that
value in hand, you can feed that to your favorite OAuth2, JWT, or whatever
library/middleware to do something meaningful with it. 

### Supplying Authorization Credentials

In the previous example we assumed that the caller supplied authorization
credentials and just retrieved/used them. Since authorization is just a value stored
on the context, you can supply them fairly easily:

```go
// Supply "Authorization: Token 12345" when calling the Hello endpoint
auth := authorization.New("Token 12345") 
ctx = authorization.WithHeader(ctx, auth)
clientA.Hello(ctx, &servicea.HelloRequest{Name: "Bob"})
```

This works when you are making the initial call, but to make life easier,
Frodo will also include that authorization on every other RPC-driven service call you make in that
request scope:


```go
func (a *ServiceA) Hello(ctx contex.Context, req *HelloRequest) (*HelloResponse, error) {
    fmt.Printf("Authorization A: %v\n", authorization.FromContext(ctx))
    clientB.EatTacos(ctx, &serviceb.EatTacosRequest{})
    ...
}

func (b *ServiceB) EatTacos(ctx contex.Context, req *EatTacosRequest) (*EatTacosResponse, error) {
    fmt.Printf("Authorization B: %v\n", authorization.FromContext(ctx))
    clientC.Goodbye(ctx, &serviceb.EatTacosRequest{})
    ...
}

func (c *ServiceC) Goodbye(ctx contex.Context, req *GoodbyeRequest) (*GoodbyeResponse, error) {
    fmt.Printf("Authorization C: %v\n", authorization.FromContext(ctx))
    ...
}

// Output
// Authorization A: Token 12345
// Authorization B: Token 12345
// Authorization C: Token 12345
```

Again, ignore the poor architecture - if you have this many dependent
calls, you probably need some asynchronous pub/sub in your life. The
point of the example, however, is that the credentials "Token 12345" was
only explicitly provided to the initial call to `Hello()`, but it was automatically propagated to service B and C because
you threaded the context through the whole thing.

### Authorization Using the JavaScript Client

When making the original call to `ServiceA.Hello()`, the JS client
doesn't utilize a "context" like Go does. To stay idiomatic to JS,
you supply it via an options argument when making your
service call:

```js
client = new ServiceAClient('...');
client.Hello({ Name: 'Bob' }, {
    authorization: 'Token 12345'
});
```

## Composing Gateways

The default behavior for your service gateways is that they will each
run in their own HTTP server, likely in their own processes. There are
a few instances, however, where you might decide that you want to run
multiple services in the same server/process. The most common is probably
for local development. It can be a pain to start/stop 15 different processes
for each of your services, so maybe you just want everything to run
in a single one. This way when you make a change to any service you just
down/up one process and see your changes.

You can use the `rpc.Compose()` function to take any N service gateways
and create a single gateway that serves up all of those services/operations.

```go
// Create gateways for each service like you normally would.
userGateway := users.NewUserServiceGateway(userService)
groupGateway := groups.NewGroupServiceGateway(groupService)
projectGateway := projects.NewProjectServiceGateway(projectService)

// Wrap them in a composed gateway that routes requests to all three.
gateway := rpc.Compose(
    userGateway,
    groupGateway,
    projectGateway,
)
http.listenAndService(":8080", gateway)
```
All 3 services will be listening on port 8080, so
you can access them via their Frodo clients; just give them all the
same address:

```go
userClient := users.NewUserServiceClient("http://localhost:8080")
groupClient := groups.NewGroupServiceClient("http://localhost:8080")
projectClient := projects.NewProjectServiceClient("http://localhost:8080")
```

If you plan to just hit the API endpoints directly, the
base address is the same, but the request paths should still correspond
to the original gateways:

```
curl -d '{"ID":"123"}' http://localhost:8080/UserService.GetByID
curl -d '{"Name":"Foo"}' http://localhost:8080/GroupService.CreateGroup
curl -d '{"Flag":true}' http://localhost:8080/ProjectService.ArchiveProject
```

## Mocking Services

When you write tests that rely on your services, Frodo can generate mock instances of your
them so that you can customize their behaviors:

```shell
$ frodo mock calculator_service.go
```

Now, you can do the following in your tests:

```go
import (
    "context"
    "fmt"

    "github.com/example/calc"
    mocks "github.com/example/calc/gen"
)

func TestSomethingThatDependsOnAddFailure(t *testing.T) {
    // You can program behaviors for Add(). If the test code calls Sub()
    // it will panic since you didn't define a behavior for that operation.
    svc := mocks.MockCalculatorService{
        AddFunc: func(ctx context.Context, req *calc.AddRequest) (*calc.AddResponse, error) {
            return nil, fmt.Errorf("barf...")
        },	
    }

    // Feed your mock service to the thing you're testing
    something := NewSomething(svc)
    _, err := something.BlahBlah(100)
    assertError(err)
    ...

    // You can also verify invocations on your service:
    assertEquals(0, svc.Calls.Sub.Times)
    assertEquals(5, svc.Calls.Add.Times)
    assertEquals(1, svc.Calls.Add.TimesFor(calc.Request{A: 4, B: 2}))
    assertEquals(2, svc.Calls.Add.TimesMatching(func(r calc.Request) bool {
        return r.A > 2
    }))
}
```

Frodo's mocks are not as fully featured as other Go mocking frameworks
out there, but it's good enough for most standard use cases. Your
services are just interfaces, so it's easy enough to bring your own
mocking framework if this won't work for you.

## Generate OpenAPI/Swagger Documentation (Experimental)

Definitely a work in progress, but in addition to generating
your backend and frontend assets, Frodo can generate OpenAPI 3.0 YAML
files to describe your API. It uses the name/type information from
your Go code as well as the GoDoc comments that you (hopefully)
write. Document your code in Go and you can get online API docs for free:

```shell
$ frodo client calculator_service.go --language=openapi
  # or
$ frodo client calculator_service.go --language=swagger
```

Now you can feed the file `gen/calculator_service.gen.swagger.yaml`
to your favorite Swagger tools. You can try it out by just pasting
the output on https://editor.swagger.io.

Not gonna lie... this is still a work in progress. I've still
got some issues to work out with nested request/response structs.
It spits out enough good stuff that it should describe your services
better than no documentation at all, though.

## Go Generate Support

If you prefer to stick to the standard Go toolchain for generating
code, you can use `//go:generate` comments to hook Frodo code
generation into your build process. Here's how you can set up your
service to generate the gateway, mock service, Go client, and JS client. 

```go
import (
   ...
)

//go:generate frodo gateway $GOFILE
//go:generate frodo client  $GOFILE
//go:generate frodo client  $GOFILE --language=js
//go:generate frodo mock    $GOFILE

type CalculatorService interface {
    ...
}
```

Now when you want to re-create your RPC artifacts, you can run a
single command (assuming you're already in the 'calc' directory):

```shell
$ go generate .
```

## Bring Your Own Templates

As Frodo matures, we will try to maintain a large number of templates for
clients in multiple popular languages (feel free to submit a PR if your
language of choice is not currently supported). If you have more specialized
needs, you can actually bring your own custom code templates to all CLI code
generation sub-commands (client, gateway, mock, and docs).

For instance, if you have your own JavaScript client template that meets your
needs better than the one that ships with Frodo, you can do the following:

```shell
frodo client calculator_service.go \
  --language=js \
  --template=mytemplates/myclient.js.tmpl
```

The path to the template can be either relative to where you're running the
command or an absolute path to a template on your hard drive. Either way, just
make sure that your template expects the root value to be a Frodo `*parser.Context`.

## Create a New Service w/ `frodo create`

This is 100% optional. As we saw in the initial example,
you can write all of your Go code starting with empty
files and have a fully distributed service in a few lines of code.

The `frodo` tool, however, has a command that generates a
lot of that boilerplate for you so that you can get straight
to solving your customers' problems.

Let's assume that you want to make a new service called
`UserService`, you can execute any  of the following commands:

```shell
frodo create user
  # or
frodo create User
  # or
frodo create UserService
```

This will create a new package in your project with all of the
following assets created:

```
[project]
  user/
    makefile
    user_service.go
    user_service_handler.go
    cmd/
      main.go
    gen/
      user_service.gen.gateway.go
      user_service.gen.client.go
```

The service will have a dummy `Create()` function
just so that there's *something* defined. You should replace
that with your own functions and implement them to make the service
do something useful.

The makefile has some convenience targets for building/running/testing
your new service as you make updates. The `build` target even
makes sure that your latest service updates get re-frodo'd so
your gateway/client are always in sync.

## Why Not Just Use gRPC?

Simply put... complexity. gRPC and grpc-gateway solve a lot of hard problems
related to distributed systems at massive scale, but those solutions come at
the cost of simplicity. Countless hours have been lost debugging issues
resolving missing dependencies in proto files, or trying to get a load balancing
solution to work, or figuring out whether to write logic in HTTP middleware or
a gRPC interceptor (just to name a few common pain points).

If many of us are honest with ourselves, a lot of
what gRPC and its ecosystem offers falls into the [YAGNI](https://en.wikipedia.org/wiki/You_aren%27t_gonna_need_it)
realm (you ain't gonna need it).

Simple solutions that create JSON-based HTTP APIs are good enough
for most of us. In those cases gRPC tends to introduce a lot of complexity without
giving as much in return. Frodo strives to have a much better developer
experience that takes you from your first line of code to a solid
set of services/APIs that are easy to maintain, test, deploy, and scale.

For the vast majority of projects, JSON (un)marshaling is probably not a performance bottleneck for you. Protobufs
are a great solution for a very specific problem, but it's one that
very, very few applications actually have. I've seen REST APIs handle hundreds
of millions of requests per day using Go's `encoding/json` package. While
cool tech, protobufs just do not solve a problem that most of us have, so
we might was well avoid the complexity that comes along with them.

Even if you get things working on your development machine, gRPC introduces
a series of other problems that you're left having to solve such as
load balancing. You could do client-side load balancing, but that involves
you having to also figure out etcd, Consul, or some other service
discovery solution. You could do server-side load balancing, but
you're going to need something like Linkerd, Envoy, or some other
not-so-simple service mesh solution. Frodo's services work the same as any other
API you might build by hand with the standard library, Gin, Chi, Echo, etc., so
the load balancers provided by AWS/GCP/Azure
just work.

Ultimately, Frodo tries to take as much the "good" from gRPC and its
ecosystem while eliminating as much of the "bad" as possible. Frodo
focuses more on the developer experience and less on giving you options
galore. It tries to provide simple solutions and sane defaults for common
problems so you spend less time figuring out how to make Frodo work and
more time solving your users' problems.

Here are a few conscious deviations from how gRPC does things:

* No proto files or other quirky DSLs to learn. Just write a Go interfaces/structs
  and Frodo will figure out automatically.
* Setup is as easy as one `go install` to get every feature. gRPC requires you to manually install protoc
  then fetch 3 or 4 grpc-related/Go dependencies, and not all have properly
  adopted Go modules yet.
* The CLI is much less complex. Even a simple gRPC service
  with an API gateway requires around 9 or 10 arguments to `protoc` in order
  to function. Contrast that with `frodo gateway foo/service.go`.
* If you don't like the CLI, you can hook into `go:generate` instead.
* In gRPC, the client it generates does not implement the service interface. It's
  really close but not enough for a strongly typed language like Go. Frodo
  makes it so that clients/gateways both implement your service interface. This
  gives you more flexibility to swap interacting with a local vs remote
  instance of the service w/ no code changes.
* The RPC layer is just JSON over HTTP. Your frontend can consume
  your services the exact same way that other backend services do.
* You've got an entire ecosystem of off-the-shelf solutions for middleware
  for logging, security, etc regardless of
  where the request comes from. With gRPC, the rules are
  different if the request came from another service vs
  through your API gateway (e.g. from your frontend).
* In gRPC, you have to jump through some [crazy hoops](https://grpc-ecosystem.github.io/grpc-gateway/docs/mapping/customizing_your_gateway/) if you 
  want anything other than a status of 200 for success 500 for failure.
  With Frodo, you can use idiomatic errors and one-line changes
  to customize this behavior.
* Since Frodo uses standard-library HTTP, traditional load balancing
  solutions like nginx or your cloud provider's load balancer
  gets the job done. No need to introduce etcd, Consul, Linkerd, Envoy,
  or any other technology into your architecture.
* Better metadata. Request-scoped metadata in gRPC is basically a map
  of string values. This forces you to marshal/unmarshal other types yourself.
  Frodo's metadata lets you pass around any type of data you want as
  you hop from service to service and will handle all that noise for you.
* Frodo has a stronger focus on generated code that is actually
  readable. If you want to treat Frodo RPC like a black box,
  you can. If you want to peek under the hood, however, you can
  do so with only minimal tears, hopefully.
