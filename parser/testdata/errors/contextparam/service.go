package contextparam

type FooService interface {
	Hello(string, *Request) (*Response, error)
}

type Request struct{}
type Response struct{}
