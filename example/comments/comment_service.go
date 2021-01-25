package comments

import (
	"context"
	"time"
)

// CommentService manages reader-submitted comments to posts. It's where horrible people get
// to spew their vitriol all over the internet and humanity dies a little.
type CommentService interface {
	// CreateComment adds a comment to a post w/ the given details.
	CreateComment(context.Context, *CreateCommentRequest) (*CreateCommentResponse, error)
	// FindByPost lists all comments submitted to the given post.
	FindByPost(context.Context, *FindByPostRequest) (*FindByPostResponse, error)
}

type Comment struct {
	ID     string
	PostID string
	Author string
	Text   string
	Date   time.Time
}

type CreateCommentRequest struct {
	PostID string
	Author string
	Text   string
}

type CreateCommentResponse Comment

type FindByPostRequest struct {
	PostID string
}

type FindByPostResponse struct {
	Comments []Comment
}
