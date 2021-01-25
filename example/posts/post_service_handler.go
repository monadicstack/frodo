package posts

import (
	"context"
	"fmt"
	"time"

	"github.com/robsignorelli/frodo/rpc/errors"
)

// PostServiceHandler implements all of the "real" functionality for the PostService. In a
// truly "real" service, you'd probably have a DB connection or Repo as a dependency to help
// store posts long-term or allow you to run multiple post services, but the point of this example
// is to show how frodo-powered services interact with each other - not how to structure code
// when dealing with long-term storage solutions.
type PostServiceHandler struct {
	posts []*Post
}

func (p *PostServiceHandler) GetPost(_ context.Context, request *GetPostRequest) (*GetPostResponse, error) {
	for _, post := range p.posts {
		if post.ID == request.ID {
			response := GetPostResponse(*post)
			return &response, nil
		}
	}
	return nil, errors.NotFound("post not found: %s", request.ID)
}

func (p *PostServiceHandler) CreatePost(_ context.Context, request *CreatePostRequest) (*CreatePostResponse, error) {
	if request.Title == "" {
		return nil, errors.BadRequest("post title is required")
	}
	if request.Text == "" {
		return nil, errors.BadRequest("post text is required")
	}

	newPost := &Post{
		ID:       fmt.Sprintf("%d", len(p.posts)+1),
		Title:    request.Title,
		Text:     request.Text,
		Archived: false,
		Date:     time.Now(),
	}
	p.posts = append(p.posts, newPost)

	response := CreatePostResponse(*newPost)
	return &response, nil
}

func (p *PostServiceHandler) Archive(_ context.Context, request *ArchiveRequest) (*ArchiveResponse, error) {
	for _, post := range p.posts {
		if post.ID == request.ID {
			post.Archived = true
			return &ArchiveResponse{}, nil
		}
	}
	return nil, errors.NotFound("post not found: %s", request.ID)
}
