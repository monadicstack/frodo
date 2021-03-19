package calc

import (
	"context"
	"net/http"
	"time"
)

// CalculatorService provides the ability to add and subtract at WEB SCALE!
type CalculatorService interface {
	// Add accepts two integers and returns a result w/ their sum.
	Add(context.Context, *AddRequest) (*AddResponse, error)
	// Sub accepts two integers and returns a result w/ their difference.
	Sub(context.Context, *SubRequest) (*SubResponse, error)
}

type PagingStuff struct {
	Text   string
	Tokens []string
	// FixedTokens is 4 strings... please work.
	FixedTokens [4]string
	LookerThing
	Stuff     map[string]uint64
	Limit     uint16 `json:"take"`
	Offset    uint64 `json:"skip"`
	Enabled   bool
	Timeout   *time.Duration
	Time      time.Time
	C         http.Client
	fooBarBaz time.Time
}

type LookerThing map[int]string

// AddRequest wrangles the two integers you plan to add together.
type AddRequest struct {
	// A is the first number to add.
	A int
	// B is the other number to add.
	B int
	// Paging is garbage
	Paging *PagingStuff
	nope   string
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
	B int
}

// SubResponse contains the result from subtracting two numbers.
type SubResponse struct {
	// Result is the difference you're returning.
	Result int
}
