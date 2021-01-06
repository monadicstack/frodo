package example

import (
	"context"
	"time"
)

type GroupServiceServer struct {
}

func (svc GroupServiceServer) CreateGroup(ctx context.Context, req *CreateGroupRequest) (*CreateGroupResponse, error) {
	return &CreateGroupResponse{
		ID:          "123",
		Name:        req.Name,
		Description: req.Description,
		Created:     time.Now(),
	}, nil
}

func (svc GroupServiceServer) DeleteGroup(ctx context.Context, req *DeleteGroupRequest) (*DeleteGroupResponse, error) {
	return &DeleteGroupResponse{ID: req.ID, Hard: req.Hard}, nil
}
