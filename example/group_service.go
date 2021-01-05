package example

import (
	"context"
	"time"
)

type GroupService interface {
	CreateGroup(ctx context.Context, request *CreateGroupRequest) (*CreateGroupResponse, error)
	DeleteGroup(ctx context.Context, request *DeleteGroupRequest) (*DeleteGroupResponse, error)
}

type CreateGroupRequest struct {
	Name        string
	Description string
}

type CreateGroupResponse struct {
	ID          string
	Name        string
	Description string
	Created     time.Time
}

type DeleteGroupRequest struct {
	ID string
}

type DeleteGroupResponse struct {
	ID string
}
