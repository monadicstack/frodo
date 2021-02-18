package paramcount

import "context"

type FooService interface {
	Hello(context.Context, *Request, int) (*Response, error)
}

type Request struct{}
type Response struct{}
