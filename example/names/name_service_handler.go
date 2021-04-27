package names

import (
	"context"
	"strings"

	"github.com/monadicstack/frodo/rpc/authorization"
	"github.com/monadicstack/frodo/rpc/errors"
)

type NameServiceHandler struct {
}

func (svc NameServiceHandler) Split(ctx context.Context, req *SplitRequest) (*SplitResponse, error) {
	if authorization.FromContext(ctx).String() == "Donny" {
		return nil, errors.PermissionDenied("donny, you're out of your element")
	}
	if req.Name == "" {
		return nil, errors.BadRequest("name is required")
	}
	// This is a really dumb algorithm, but that's not the point of this service. We just need something to test
	// client/server communication.
	tokens := strings.Split(req.Name, " ")
	response := SplitResponse{}
	response.FirstName = tokens[0]
	response.LastName = strings.Join(tokens[1:], " ")
	return &response, nil
}

func (svc NameServiceHandler) FirstName(ctx context.Context, req *FirstNameRequest) (*FirstNameResponse, error) {
	res, err := svc.Split(ctx, &SplitRequest{Name: req.Name})
	if err != nil {
		return nil, err
	}
	return &res.FirstNameResponse, nil
}

func (svc NameServiceHandler) LastName(ctx context.Context, req *LastNameRequest) (*LastNameResponse, error) {
	res, err := svc.Split(ctx, &SplitRequest{Name: req.Name})
	if err != nil {
		return nil, err
	}
	return &res.LastNameResponse, nil
}

func (svc NameServiceHandler) SortName(ctx context.Context, req *SortNameRequest) (*SortNameResponse, error) {
	res, err := svc.Split(ctx, &SplitRequest{Name: req.Name})
	if err != nil {
		return nil, err
	}

	if res.LastName == "" {
		return &SortNameResponse{SortName: strings.ToLower(res.FirstName)}, nil
	}
	return &SortNameResponse{SortName: strings.ToLower(res.LastName + ", " + res.FirstName)}, nil
}
