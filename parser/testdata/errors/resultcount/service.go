package resultcount

import "context"

type FooService interface {
	Hello(context.Context, *Request) (*Response, error, int)
}

type Request struct{}
type Response struct{}
