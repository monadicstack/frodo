# Frodo Architecture

As a Frodo user, you may not really care how the magic code generation
and runtime nuts and bolts work. You probably just want to get something
running, so executing `frodo xxx my_service.go` is all you usually need to do.

If, however, you want to contribute, fork this repo, or build your own
custom templates, it might be helpful to know how everything works under the hood.

## Key Functions/Structs

If you want to skip the exposition and just explore, here are some of the most
important places in the code and what purpose they serve:

* `parser.Context` - Provides a simplified view of the parsed service code including
  info from the Abstract Syntax Tree, GoDoc, and type parsing that goes on. It is
  a single snapshot of everything Frodo cares about your service and is the input
  value when evaluating code generation templates.
* `parser.Parse()` - This accepts the user's service definition file, parses it,
  and spits out a `parser.Context` with all the info Frodo cares about.
* `generate.FileTemplate` - It's basically a pointer to a code template that you
  want to generate. By using the `io.FS` interface we can easily change it from
  pointing to an embedded file as the template to an `os.DirFS` that lets users
  supply their own templates.
* `generate.File()` - This accepts a parser context (i.e. service snapshot data),
  and a file template (i.e. what template am I generating). It then evaluates that
  Go template w/ the context to spit out the generated code file. When generating
  another Go file, it will also run it through `gofmt` to make it pretty.
* `generate/templates` - This is where we store the Go text template files for
  every asset that Frodo supports out of the box.
* `cli/generate_xxx.go` - Each of these handle the logic for the `frodo` tool
  subcommands. For instance `frodo gateway` is handled by `generate_gateway.go`,
  `frodo client` is handled by `generate_client.go`, and so on.

## High Level Workflow

At its core, Frodo is pretty simple. When you execute a CLI command like
`frodo gateway my_service.go`, we perform a few basic tasks.

#### Step 1: Parse Your Code Into An Abstract Syntax Tree

Frodo uses the standard library's `go/ast` package to generate an Abstract
Syntax Tree (AST) of your source file (the one with the service interface).
Basically, Frodo parses it to the same few data structures used by the Go
compiler - I didn't reinvent that wheel.

In addition to exposing their compiler tools in the standard library, the Go
team also exposed GoDoc processing via the `go/doc` package. Frodo
runs your code file through that parser as well to get access to the comments
in your code. That's how we are later able to parse Doc Options such
as `// HTTP 202` and `// VERSION 1.0.2` and so on.

#### Step 2: Populate a `parser.Context`

While the AST has the info we need, it would be an absolute nightmare to try
to walk this tree in a Go template (we'll get here in the next step). You have
to perform a lot of type assertions to determine if you're processing a struct
or an interface or whatever; and depending on which one you've got, you'll
process things very differently.

To make life easier, the parser normalizes the AST, the GoDoc info, and the
parsed `go.mod` file into a `parser.Context`. It's a much more streamlined
structure that better fits the mental model of a service and its inputs/outputs.

#### Step 3: Evaluate The Go Template to Output The Code File

Now that we've got a `parser.Context` that represents your service, the
module it belongs to, and all the input/output types, we can generate the
desired code file. Frodo just uses standard Go templates to generate a large string
that it will write to the appropriate file. Again, we've tried to stick as
close to the standard library for everything.

You'll find the built-in templates in the `[frodo]/generate/templates`
directory. The CLI lets you supply your own templates if you want, but it
will ultimately get processed exactly the same. Frodo will parse the
template using the standard `text/template` package (not `html/template`).
Then it will `Eval()` the template passing in the `parser.Context`
as the input data value.

## The Frodo "Runtime"

Basically this refers to the `rpc.Client` and `rpc.Gateway` structs. Regardless
of what your service is or what it does, there's certain functionality that gets
baked into your generated RPC code:

* Support for middleware functions
* Authorization management
* Metadata support and propagation
* HTTP route management
* Error handling
* Transport
* Data binding
* yada, yada, yada...

Originally, Frodo just baked all of that stuff directly into the code that it
generated for the Go client/gateway, but I quickly moved all of that common stuff to the
`rpc.Client` and `rpc.Gateway` structs.

There are pro's and con's to both approaches, but I preferred having the ability
to apply fixes/updates to most of the core RPC functionality without having to
have you re-generate all of your service artifacts. You just rev up your Frodo
version and re-deploy.

### rpc.Gateway

This bit of magic helps to expose your service to other consumers. For simplicity,
Frodo only uses the standard library's HTTP server/handler to allow remote clients
to make calls to the service. At its core, the Gateway is just an `http.Handler`.
All the routing, transport, etc gets hidden behind a single `ServeHTTP()` function.
That's why we can just feed it to `http.ListenAndServe()` in order to have a
working service.

Under the hood, Frodo generates an HTTP route for every single method on your
service interface. For instance, if you're generating the gateway for the
`GreeterService` and it has 2 methods; `SayHello()` and `SayGoodbye()`, the
gateway will expose 2 routes:

* `POST /GreeterService.SayHello`
* `POST /GreeterService.SayGoodbye`

Both include JSON marshaling, error handling, HTTP header/status/body management,
and so on. Ultimately, when an HTTP server accepts a request to one of those
two endpoints, the gateway is wired to call the appropriate function. It will
decode the JSON body onto an instance of `SayHelloRequest` or `SayGoodbyeRequest`
and pass it to the "real" service to get the work done. The `SayXxxResponse` that
those functions return will be JSON-ified and sent back as the HTTP response.

### rpc.Client

One of Frodo's main tenets is that both your client and server should implement
the service interface. This allows you to seamlessly swap implementations
in your application code to go from local to remote as you wish. That being the
case, when you generate a service client, it should be an implementation of the
service (i.e. has all the same methods) but each method delegates to some
remote instance (i.e. makes HTTP requests to the gateway to get the work done).

If you look at the generated code for any client, you'll see that the clients
are just some sugar around an HTTP client that hits specific endpoints based
on which function you're calling. While the Go client also contains extra features
such as metadata and contexts, all the supported language clients wrap an HTTP
client and handle JSON marshaling, error handling, etc.

## How We Test Clients In Other Languages

Most Frodo functionality is easily handled via unit tests. I tried to make
nicely isolated components that can be tested without too much fuss. One exception
to this pattern is testing the parser. We get better coverage and more confidence
in the behavior by just having a large set of real Go code files that define 
services and structs in all sorts of wacky ways. Then we just validate the context
that is crapped out of the other end.

Testing the auto-generated code for other languages, however, is a bit more tricky.
To support this, each language/client has a separate "runner" program in that
language whose job it is to instantiate the generated service client and make
calls to the gateway running in your Go test code.

These runner programs all accept a single argument; the test case that we're
trying to exercise in our test. The runner will create a client and make one
or more calls that utilize that functionality. The results of each service
interaction are then written to stdout with info about whether that call
succeeded or failed as well as the result/error that the call generated as
a nice JSON payload.

For instance, when testing the JS/node client, the `run_client.js` runner
behaves like this:

```shell
# Runs 5 client functions that all succeed.
$ node run_client.js Success
OK {"FirstName":"Jeff","LastName":"Lebowski"}
OK {"FirstName":"Jeff"}
OK {"LastName":"Lebowski"}
OK {"SortName":"lebowski, jeff"}
OK {"SortName":"dude"}
```

or...

```shell
# Runs 4 client functions that have bad authorization and should fail.
$ node run_client.js AuthFailureCall
FAIL {"status":403, "message": "donny, you're out of your element"}
FAIL {"status":403, "message": "donny, you're out of your element"}
FAIL {"status":403, "message": "donny, you're out of your element"}
FAIL {"status":403, "message": "donny, you're out of your element"}
```

With this helpful runner program, our Go test code can follow this
basic flow:

* (SetupTest) Fire up the gateway for the proper service.
* Execute the language-specific runner program w/ the test case argument.
* Capture/parse stdout so that we can make sense of what succeeded/failed in the other language.
* Make assertions based on the output so that we know when the behavior meets/fails our expectations.  
* (TearDownTest) Shut down the gateway for the proper service.

## Random Special Cases and Gotchas

* Doc options does not allow OPTIONS HTTP method. One of Frodo's goals is to
  make it easy to consume your services anywhere; this includes your web
  frontend. That being the case, you're likely going to want to throw CORS
  into the mix. As a result when Frodo registers your service method with the
  gateway, it will register the POST (or whatever you configured) as well as
  an OPTIONS for the same path. Now when you add your off-the-shelf CORS
  middleware, the gateway's internal router will already let that request through
  far enough to even hit your middleware. There are more details in the comments
  for the `Register()` method on `rpc.Gateway`.
* When scraping GoDoc comments for Doc Options, we need to pull them from two
  different parsing locations. For some reason we have
  access to the comments on interfaces/types are available when running the
  source through `go/doc`, but comments on methods are only available on the AST.
  I haven't been able to find a single documentation processing scheme that gets
  ALL GoDoc comments at once, so the `parser.ParseDocumentation` function
  actually traverses two separate trees to grab the comments for all the types
  of things you can apply comments to.
