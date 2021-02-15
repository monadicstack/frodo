package authorization

import (
	"context"
	"strings"
)

// The context key for when we capture the Authorization HTTP header.
type contextKeyHeader struct{}

// None is a fixed value for a missing or blank header that implies no credentials.
var None = Header{}

// WithHeader returns a new child context that contains the header value.
func WithHeader(ctx context.Context, header Header) context.Context {
	return context.WithValue(ctx, contextKeyHeader{}, &header)
}

// FromContext fetches the authorization Header value stored on this context (if any). If there is no
// header on this context, you'll simply get back a zero-value header.
func FromContext(ctx context.Context) Header {
	if ctx == nil {
		return None
	}
	header, ok := ctx.Value(contextKeyHeader{}).(*Header)
	if !ok {
		return None
	}
	return *header
}

// New creates a Header based on the HTTP Authorization header value that can be used for RPC authorization.
func New(value string) Header {
	return Header{value: strings.TrimSpace(value)}
}

// Header wraps the raw string from the HTTP Authorization header.
type Header struct {
	value string
}

// Empty returns true when the fully qualified header value is blank.
func (h Header) Empty() bool {
	return h.String() == ""
}

// NotEmpty returns true when the fully qualified header value has at least one character.
func (h Header) NotEmpty() bool {
	return h.String() != ""
}

func (h Header) String() string {
	return h.value
}
