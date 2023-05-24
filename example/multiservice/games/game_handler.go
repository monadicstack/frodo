package games

import (
	"context"
	"fmt"

	"github.com/davidrenne/frodo/rpc/errors"
)

// GameServiceHandler implements all of the "real" functionality for the GameService.
type GameServiceHandler struct {
	// Repo manages access to the underlying data store. Even when working with Frodo-powered services
	// you can still use standard Go dependency injection to get your services running.
	Repo Repo
}

// GetByID looks up a game record given its unique id.
func (svc *GameServiceHandler) GetByID(ctx context.Context, req *GetByIDRequest) (*GetByIDResponse, error) {
	if req.ID == "" {
		return nil, errors.BadRequest("id is required")
	}

	game, err := svc.Repo.GetByID(ctx, req.ID)
	if err != nil {
		return nil, fmt.Errorf("unable to look up game record: %w", err)
	}

	response := GetByIDResponse(game)
	return &response, nil
}

// Register adds another game record to our gaming database.
func (svc *GameServiceHandler) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	if req.Name == "" {
		return nil, errors.BadRequest("create: name is required")
	}

	game, err := svc.Repo.Create(ctx, Game{
		Name:        req.Name,
		Description: req.Description,
		Publisher:   req.Publisher,
		Year:        req.Year,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create game record: %w", err)
	}

	response := RegisterResponse(game)
	return &response, err
}
