package names

import (
	"context"
	"io"
)

type NameService interface {
	Split(ctx context.Context, req *SplitRequest) (*SplitResponse, error)
	FirstName(ctx context.Context, req *FirstNameRequest) (*FirstNameResponse, error)
	LastName(ctx context.Context, req *LastNameRequest) (*LastNameResponse, error)
	SortName(ctx context.Context, req *SortNameRequest) (*SortNameResponse, error)
	Download(ctx context.Context, req *DownloadRequest) (*DownloadResponse, error)
	DownloadExt(ctx context.Context, req *DownloadExtRequest) (*DownloadExtResponse, error)
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

type DownloadRequest struct {
	Name string
}

type DownloadResponse struct {
	reader io.ReadCloser
}

func (r DownloadResponse) Content() io.ReadCloser {
	return r.reader
}

func (r *DownloadResponse) SetContent(reader io.ReadCloser) {
	r.reader = reader
}

type DownloadExtRequest struct {
	Name string
	Ext  string
}

type DownloadExtResponse struct {
	DownloadResponse
	contentType     string
	contentFileName string
}

func (r DownloadExtResponse) ContentType() string {
	return r.contentType
}

func (r *DownloadExtResponse) SetContentType(contentType string) {
	r.contentType = contentType
}

func (r DownloadExtResponse) ContentFileName() string {
	return r.contentFileName
}

func (r *DownloadExtResponse) SetContentFileName(contentFileName string) {
	r.contentFileName = contentFileName
}
