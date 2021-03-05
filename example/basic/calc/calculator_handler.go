package calc

import (
	"context"
)

// CalculatorServiceHandler implements all of the "real" functionality for the CalculatorService.
type CalculatorServiceHandler struct{}

func (c CalculatorServiceHandler) Add(_ context.Context, req *AddRequest) (*AddResponse, error) {
	sum := req.A + req.B
	return &AddResponse{Result: sum}, nil
}

func (c CalculatorServiceHandler) Sub(_ context.Context, req *SubRequest) (*SubResponse, error) {
	diff := req.A - req.B
	return &SubResponse{Result: diff}, nil
}
