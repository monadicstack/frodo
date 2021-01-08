package example

import (
	"context"
	"fmt"
	"time"
)

type GroupServiceServer struct {
}

func (svc GroupServiceServer) GetByID(ctx context.Context, request *GetByIDRequest) (*GetByIDResponse, error) {
	return &GetByIDResponse{
		ID:          request.ID,
		Name:        "The Bees",
		Description: "They know the way of the bee.",
	}, nil
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
	fmt.Println(">>>>>>> DELETING GROUP:", req.ID, "[", req.Hard, "]")
	return &DeleteGroupResponse{ID: req.ID, Hard: req.Hard}, nil
}
