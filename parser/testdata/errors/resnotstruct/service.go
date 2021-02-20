package resnotstruct

import (
	"context"
	"net/http"
)

type FooService interface {
	Hello(context.Context, Request) (*http.Client, error)
}

type Request struct{}
