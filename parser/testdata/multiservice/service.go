package multiservice

import "context"

type FooService interface {
	Hello(context.Context, *RequestA) (*ResponseA, error)
}

type BarService interface {
	Goodbye(context.Context, *RequestA) (*ResponseA, error)
	Goodbye2(context.Context, *RequestB) (*ResponseB, error)
}

type RequestA struct{}
type RequestB struct{}
type ResponseA struct{}
type ResponseB struct{}
