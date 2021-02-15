package comments

import (
	"context"
	"fmt"
	"time"

	"github.com/monadicstack/frodo/example/posts"
	"github.com/monadicstack/frodo/rpc/errors"
)

// CommentServiceHandler implements all of the "real" functionality for the CommentService. Just like
// with the PostService, we're doing a lot of hand-waving for the storage of comment data since the point
// is just to show you how frodo-powered services communicate over RPC.
type CommentServiceHandler struct {
	// PostService is an RPC client to the posts service used to fetch status info about
	// posts that we're trying to add comments to.
	PostService posts.PostService
	// comments is our "database" of submitted comments.
	comments []*Comment
}

func (svc *CommentServiceHandler) CreateComment(ctx context.Context, request *CreateCommentRequest) (*CreateCommentResponse, error) {
	if request.PostID == "" {
		return nil, errors.BadRequest("comment post id is required")
	}
	if request.Text == "" {
		return nil, errors.BadRequest("comment text is required")
	}
	if request.Author == "" {
		return nil, errors.BadRequest("comment author is required")
	}

	// We don't allow you to comment on posts that have been archived. Here we make our RPC
	// call to the PostService. Notice that because both the client and handler implement PostService
	// we don't actually know which we're dealing with here. That's great for testing/mocking
	// as well as giving you flexibility to run multiple services in the same process or
	// distributing them across your infrastructure. This code remains 100% the same.
	post, err := svc.PostService.GetPost(ctx, &posts.GetPostRequest{ID: request.PostID})
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve post: %w", err)
	}
	if post.Archived {
		return nil, errors.BadRequest("unable to comment on archived post")
	}

	newComment := &Comment{
		ID:     fmt.Sprintf("%d", len(svc.comments)+1),
		PostID: request.PostID,
		Author: request.Author,
		Text:   request.Text,
		Date:   time.Now(),
	}
	svc.comments = append(svc.comments, newComment)

	response := CreateCommentResponse(*newComment)
	return &response, nil
}

func (svc *CommentServiceHandler) FindByPost(ctx context.Context, request *FindByPostRequest) (*FindByPostResponse, error) {
	if request.PostID == "" {
		return nil, errors.BadRequest("comment post id is required")
	}

	// Make sure that the post exists before allowing comments on it. The service should return
	// some sort of `errors.NotFound()` error if it doesn't exist.
	_, err := svc.PostService.GetPost(ctx, &posts.GetPostRequest{ID: request.PostID})
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve post: %w", err)
	}

	var results []Comment
	for _, comment := range svc.comments {
		if comment.PostID == request.PostID {
			results = append(results, *comment)
		}
	}
	return &FindByPostResponse{Comments: results}, nil
}
