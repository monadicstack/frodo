package names

import "context"

type NameService interface {
	Split(ctx context.Context, req *SplitRequest) (*SplitResponse, error)
	FirstName(ctx context.Context, req *FirstNameRequest) (*FirstNameResponse, error)
	LastName(ctx context.Context, req *LastNameRequest) (*LastNameResponse, error)
	SortName(ctx context.Context, req *SortNameRequest) (*SortNameResponse, error)
}

type NameRequest struct {
	Name string
}

type SplitRequest NameRequest

type SplitResponse struct {
	FirstNameResponse
	LastNameResponse
}

type FirstNameRequest struct {
	Name string
}

type FirstNameResponse struct {
	FirstName string
}

type LastNameRequest struct {
	Name string
}

type LastNameResponse struct {
	LastName string
}

type SortNameRequest struct {
	Name string
}

type SortNameResponse struct {
	SortName string
}
