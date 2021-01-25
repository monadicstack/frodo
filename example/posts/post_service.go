package posts

import (
	"context"
	"time"
)

// PostService is a service that manages blog/article posts. This is just for example purposes,
// so this is not a truly exhaustive set of operations that you might want if you were *really*
// building some sort of blog/CRM engine.
type PostService interface {
	// GetPost fetches a Post record by its unique identifier.
	GetPost(context.Context, *GetPostRequest) (*GetPostResponse, error)
	// CreatePost creates/stores a new Post record.
	CreatePost(context.Context, *CreatePostRequest) (*CreatePostResponse, error)
	// Archive effectively disables a Post from appearing in the article list.
	Archive(context.Context, *ArchiveRequest) (*ArchiveResponse, error)
}

type Post struct {
	ID       string
	Title    string
	Text     string
	Archived bool
	Date     time.Time
}

type GetPostRequest struct {
	ID string
}

type GetPostResponse Post

type CreatePostRequest struct {
	Title string
	Text  string
}

type CreatePostResponse Post

type ArchiveRequest struct {
	ID string
}

type ArchiveResponse struct {
}
