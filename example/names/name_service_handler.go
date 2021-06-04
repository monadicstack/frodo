package names

import (
	"bytes"
	"context"
	"io"
	"strings"

	"github.com/monadicstack/frodo/rpc/authorization"
	"github.com/monadicstack/frodo/rpc/errors"
)

// NameServiceHandler provides the reference implementation of the NameService.
type NameServiceHandler struct {
}

// Split separates a first and last name.
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

// FirstName extracts just the first name from a full name string.
func (svc NameServiceHandler) FirstName(ctx context.Context, req *FirstNameRequest) (*FirstNameResponse, error) {
	res, err := svc.Split(ctx, &SplitRequest{Name: req.Name})
	if err != nil {
		return nil, err
	}
	return &res.FirstNameResponse, nil
}

// LastName extracts just the last name from a full name string.
func (svc NameServiceHandler) LastName(ctx context.Context, req *LastNameRequest) (*LastNameResponse, error) {
	res, err := svc.Split(ctx, &SplitRequest{Name: req.Name})
	if err != nil {
		return nil, err
	}
	return &res.LastNameResponse, nil
}

// SortName establishes the "phone book" name for the given full name.
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

// Download returns a raw CSV file containing the parsed name.
func (svc NameServiceHandler) Download(ctx context.Context, req *DownloadRequest) (*DownloadResponse, error) {
	split, err := svc.Split(ctx, &SplitRequest{Name: req.Name})
	if err != nil {
		return nil, err
	}

	buf := bytes.Buffer{}
	buf.WriteString("first,last\n")
	buf.WriteString(split.FirstName)
	buf.WriteString(",")
	buf.WriteString(split.LastName)
	return &DownloadResponse{reader: io.NopCloser(&buf)}, nil
}

// DownloadExt returns a raw CSV file containing the parsed name.
func (svc NameServiceHandler) DownloadExt(ctx context.Context, req *DownloadExtRequest) (*DownloadExtResponse, error) {
	split, err := svc.Split(ctx, &SplitRequest{Name: req.Name})
	if err != nil {
		return nil, err
	}

	buf := bytes.Buffer{}
	buf.WriteString("first,last\n")
	buf.WriteString(split.FirstName)
	buf.WriteString(",")
	buf.WriteString(split.LastName)

	res := DownloadExtResponse{}
	res.reader = io.NopCloser(&buf)
	res.contentType = "text/" + req.Ext
	res.contentFileName = "name." + req.Ext
	return &res, nil
}
