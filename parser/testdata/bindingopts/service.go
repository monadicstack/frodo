package bindingopts

import "context"

type BindingService interface {
	BindIt(context.Context, *Request) (*Response, error)
}

type Request struct {
	ID        string `json:"record_id"`
	Name      string `json:"Name"`
	OmitMe    string `json:"-"`
	IncludeMe string `json:"include,omitempty"`
}

type Response struct{}
