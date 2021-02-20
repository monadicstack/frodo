package resulterror

import "context"

type FooService interface {
	Hello(context.Context, *Request) (*Response, Error)
}

type Request struct{}
type Response struct{}

type Error struct{}
