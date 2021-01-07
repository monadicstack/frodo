package example

import (
	"context"
	"time"
)

type GroupService interface {
	// CreateGroup makes a new group... duh.
	CreateGroup(ctx context.Context, request *CreateGroupRequest) (*CreateGroupResponse, error)

	// DeleteGroup smokes the group if it exists.
	//
	// DELETE /group/:ID
	// HTTP 202
	DeleteGroup(ctx context.Context, request *DeleteGroupRequest) (*DeleteGroupResponse, error)
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
