package example

import (
	"context"
	"time"
)

type GroupService interface {
	// GetByID looks up a group given its unique identifier. We will return a NotFoundError
	// instead of a nil response if there is no record with that id.
	//
	// Monkeys are cool.
	//
	// GET /group/:id
	GetByID(ctx context.Context, request *GetByIDRequest) (*GetByIDResponse, error)

	// CreateGroup makes a new group... duh.
	CreateGroup(ctx context.Context, request *CreateGroupRequest) (*CreateGroupResponse, error)

	// HTTP 202
	// DELETE /group/:id
	DeleteGroup(ctx context.Context, request *DeleteGroupRequest) (*DeleteGroupResponse, error)
}

type GetByIDRequest struct {
	ID   string `json:"id"`
	Flag bool   `json:"flag"`
}

type GetByIDResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CreateGroupRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CreateGroupResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Created     time.Time `json:"created"`
}

type DeleteGroupRequest struct {
	ID   string `json:"id"`
	Hard string `json:"hard"`
}

type DeleteGroupResponse struct {
	ID   string `json:"id"`
	Hard string `json:"hard"`
}
