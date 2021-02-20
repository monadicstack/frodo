package resnotpointer

import "context"

type FooService interface {
	Hello(context.Context, *Request) (Response, error)
}

type Request struct{}
type Response struct{}
