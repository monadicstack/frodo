package noservices

import "context"

/*
 * The service interface must actually end in "Service" to count, so we won't sweep up "Blah".
 */

type Hello interface {
	Say(context.Context, *Request) (*Response, error)
}

type Request struct {
	ID string
}

type Response struct {
	ID   string
	Name string
	Age  int
}
