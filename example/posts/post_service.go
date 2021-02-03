package posts

import (
	"context"
	chrono "time"
)

// PostService is a service that manages blog/article posts. This is just for example purposes,
// so this is not a truly exhaustive set of operations that you might want if you were *really*
// building some sort of blog/CRM engine.
//
// PATH /v2
type PostService interface {
	// GetPost fetches a Post record by its unique identifier.
	//
	// GET /post/:id
	GetPost(context.Context, *GetPostRequest) (*GetPostResponse, error)
	// CreatePost creates/stores a new Post record.
	//
	// POST /post
	CreatePost(context.Context, *CreatePostRequest) (*CreatePostResponse, error)
	// Archive effectively disables a Post from appearing in the article list.
	// PATCH /post/:id/archive
	// HTTP 202
	Archive(context.Context, *ArchiveRequest) (*ArchiveResponse, error)
}

type ShortText string

type Post struct {
	ID string
	// Title is the one-line headline for the post.
	Title ShortText
	Text  string
	// Archived determines if the post is active or not.
	Archived *bool
	Date     chrono.Time
}

type GetPostRequest struct {
	ID string
}

// GetPostResponse describes a single post in the blog.
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
