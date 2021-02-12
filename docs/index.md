# Frodo

Frodo is a code generator and runtime library that helps
you write RPC-enabled (micro) services and APIs. It parses
the interfaces/structs/comments in your service code to
generate all of your client/server communication code.

## Features

* No .proto files. Your services are just idiomatic Go code.
* Write Go code that solves a problem and immediately start
  communicating over HTTP/JSON to your frontend and other backend services.
* Generate robust HTTP APIs/gateways to expose your services. 
* Ggenerate clients in multiple languages to consume your services
  without the need for communication/transport code.
* Generate complete OpenAPI documentation to describe your
  APIs, operations, schemas, and such.
* They're just HTTP-based APIs, so deploying to your cloud
  infrastructure or Kubernetes doesn't require additional complexity. 
* Insert off-the-shelf middleware functions into your service's flow.
* Customize RESTful URLS for your API or stick with
  RPC-style URLs; your choice.
* Easily control HTTP status codes for successes and failures.
* Automatic data binding so that raw request data is converted
  into Go structs consumed by your services.
* Seamlessly pass request-scoped metadata between your services.
* Embeddable decorators for all of your services so that you can
  separate service concerns like security and business logic more easily.

## Motivations (Problems Solved)

It's all too common in service-based architectures and APIs
to have your code get bogged down in the minutia of how
to convert HTTP data into Go data that you can process and then
converting it back to HTTP data when you want to respond. I've
seen plenty of 30-line HTTP handler functions that are 3 lines
of "logic" and 27 lines of query string, header, and status
code management. It distracts from the true purpose of the
code and hurts maintainability. It also tends to be very
boilerplate and repetitive code that just takes time away from
you solving real problems for your users.

gRPC and `grpc-gateway` have grown in popularity with the Go community as they
address this problem.

The issue? It adds too much complexity.

There's the 

## Source

[GitHub Repo](https://github.com/robsignorelli/frodo)
