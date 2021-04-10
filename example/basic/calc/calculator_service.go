package calc

import (
	"context"
)

// CalculatorService provides the ability to add and subtract at WEB SCALE!
type CalculatorService interface {
	// Add accepts two integers and returns a result w/ their sum.
	Add(context.Context, *AddRequest) (*AddResponse, error)
	// Sub accepts two integers and returns a result w/ their difference.
	// GET /sub/:A/:B/please
	Sub(context.Context, *SubRequest) (*SubResponse, error)
}

// AddRequest wrangles the two integers you plan to add together.
type AddRequest struct {
	// A is the first number to add.
	A int
	// B is the other number to add.
	B        int
	Criteria PagingStuff `json:"criteria"`
}

// AddResponse contains the result from adding two numbers.
type AddResponse struct {
	// Result is the sum you're returning.
	Result int
}

// SubRequest wrangles the two integers you plan to subtract.
type SubRequest struct {
	// A is the "minuend" in the subtraction operation.
	A int
	// B is the "subtrahend" in the subtraction operation.
	B        int
	Criteria PagingStuff `json:"criteria"`
}

// SubResponse contains the result from subtracting two numbers.
type SubResponse struct {
	// Result is the difference you're returning.
	Result int
}

type PagingStuff struct {
	Offset  int `json:"skip"`
	Limit   int
	Tokens  []string
	Person  Name
	People  []Name
	People2 [4]Name
}

type Name struct {
	ID    string
	First string
	Last  string
}
