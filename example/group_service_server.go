package example

import (
	"context"
	"fmt"
	"time"

	"github.com/robsignorelli/frodo/rpc/metadata"
)

type GroupServiceServer struct {
}

func (svc GroupServiceServer) GetByID(ctx context.Context, request *GetByIDRequest) (*GetByIDResponse, error) {
	var someNumber int
	metadata.Value(ctx, "foo", &someNumber)
	var other GetByIDRequest
	metadata.Value(ctx, "bar", &other)
	var other2 GetByIDRequest
	metadata.Value(ctx, "bar", &other2)
	callName := ctx.Value("ServiceCall")
	return &GetByIDResponse{
		ID:          request.ID,
		Name:        fmt.Sprintf("The Bees: [id=%s][flag=%v]", request.ID, request.Flag),
		Description: fmt.Sprintf("They know the way of the bee: %v => %+v : %+v: %s", someNumber, other, other2, callName),
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
