package basic

import "context"

type DudeService interface {
	Bowl(context.Context, *BowlRequest) (*BowlResponse, error)
}

type BowlRequest struct {
	BowlerID string
}

type BowlResponse struct {
	BowlerID string
	Pins     int
}
