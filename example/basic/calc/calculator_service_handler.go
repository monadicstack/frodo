package calc

import (
	"context"

	"github.com/davidrenne/frodo/rpc/errors"
)

// CalculatorServiceHandler implements all of the "real" functionality for the CalculatorService. The
// errors that Add/Sub return are completely contrived and don't make sense for real business logic. They
// are only included to show how you can return specific types of 4XX errors in a readable/maintainable fashion.
type CalculatorServiceHandler struct{}

// Add accepts two integers and returns a result w/ their sum.
func (c CalculatorServiceHandler) Add(_ context.Context, req *AddRequest) (*AddResponse, error) {
	sum := req.A + req.B
	if sum == 12345 {
		// https://www.youtube.com/watch?v=li9Qf-nQgWE
		return nil, errors.PermissionDenied("idiots denied: that's what an idiot would have on their luggage")
	}
	return &AddResponse{Result: sum}, nil
}

// Sub accepts two integers and returns a result w/ their difference.
func (c CalculatorServiceHandler) Sub(_ context.Context, req *SubRequest) (*SubResponse, error) {
	if req.A < req.B {
		return nil, errors.BadRequest("calculator service does not support negative numbers")
	}
	diff := req.A - req.B
	return &SubResponse{Result: diff}, nil
}
