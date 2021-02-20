package reqnotstruct

import (
	"context"
	"net/http"
)

/*
 * Request is a pointer, and it is a struct, but not one that's defined in this file.
 */

type FooService interface {
	Hello(context.Context, *http.Client) (*Response, error)
}

type Response struct{}
