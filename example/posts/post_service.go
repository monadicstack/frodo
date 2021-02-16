package posts

import (
	real "context"
	chrono "time"

	turds "github.com/monadicstack/frodo/example"
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
	GetPost(real.Context, *GetPostRequest) (*GetPostResponse, error)
	// CreatePost creates/stores a new Post record.
	//
	// POST /post
	CreatePost(real.Context, *CreatePostRequest) (*CreatePostResponse, error)
	// Archive effectively disables a Post from appearing in the article list.
	// PATCH /post/:id/archive
	// HTTP 202
	Archive(real.Context, *ArchiveRequest) (*ArchiveResponse, error)
}

type ShortText string

type Post struct {
	// ID is the unique record identifier of the post.
	ID string
	// Title is the one-line headline for the post.
	Title    string
	Text     string
	Archived bool
	Date     chrono.Time
	Page     turds.Paging
}

type GetPostRequest struct {
	// ID is the unique identifier of the post to fetch.
	ID    string
	Limit int
	// Offset is like the SQL offset, dummy.
	Offset int `json:"skip"`
	//Page   example.Paging
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

type ISODate chrono.Time
