package scores

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/monadicstack/frodo/example/multiservice/games"
	"github.com/monadicstack/frodo/rpc/errors"
)

// ScoreServiceHandler implements all of the "real" functionality for the ScoreService.
type ScoreServiceHandler struct {
	// Games is our client to interact with the GameService.
	//
	// Frodo Notes: In cmd/main.go we'll pass "games.NewGameServiceClient()" in for this value. Both the
	// gateway and the client implement the GameService interface, so we don't actually care which
	// we are getting here. This makes it easy for you to run everything in the same process if you want
	// or distribute them across different processes/servers.
	Games games.GameService
	// Repo manages access to the underlying data store. Even when working with Frodo-powered services
	// you can still use standard Go dependency injection to get your services running.
	Repo Repo
}

// NewHighScore captures a player's high score for the given game.
func (svc *ScoreServiceHandler) NewHighScore(ctx context.Context, request *NewHighScoreRequest) (*NewHighScoreResponse, error) {
	if request.GameID == "" {
		return nil, errors.BadRequest("high scores: game id is required")
	}
	if request.PlayerName == "" {
		return nil, errors.BadRequest("high scores: player name  is required")
	}

	// First, check with the game service to make sure that this corresponds to a real game. Propagate
	// the failure from the service if it's not. Since Games is a Frodo-powered service, the 'err' you receive
	// should already have a meaningful message and status code information baked in. As long as you use the
	// Go 1.13 error wrapping verb "%w", any HTTP-status-code-having error will be preserved even though
	// we're wrapping it. So if the game service returned a 403-style forbidden error, this call will
	// result in a 403-style error as well.
	game, err := svc.Games.GetByID(ctx, &games.GetByIDRequest{ID: request.GameID})
	if err != nil {
		return nil, fmt.Errorf("unable to look up game: %w", err)
	}

	// The game is legit, so record the high score.
	score, err := svc.Repo.Create(ctx, HighScore{
		GameID:     game.ID,
		PlayerName: request.PlayerName,
		Score:      request.Score,
		Date:       time.Now(),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create score record: %w", err)
	}

	response := NewHighScoreResponse(score)
	return &response, nil
}

// HighScoresForGame fetches the top "N" high scores achieved by any player
func (svc ScoreServiceHandler) HighScoresForGame(ctx context.Context, request *HighScoresForGameRequest) (*HighScoresForGameResponse, error) {
	if request.GameID == "" {
		return nil, errors.BadRequest("high scores: game id is required")
	}

	// First, check with the game service to make sure that this corresponds to a real game. Propagate
	// the failure from the service if it's not. Since Games is a Frodo-powered service, the 'err' you receive
	// should already have a meaningful message and status code information baked in. As long as you use the
	// Go 1.13 error wrapping verb "%w", any HTTP-status-code-having error will be preserved even though
	// we're wrapping it. So if the game service returned a 403-style forbidden error, this call will
	// result in a 403-style error as well.
	game, err := svc.Games.GetByID(ctx, &games.GetByIDRequest{ID: request.GameID})
	if err != nil {
		return nil, fmt.Errorf("unable to look up game: %w", err)
	}

	gameScores, err := svc.Repo.ByGame(ctx, game.ID)
	if err != nil {
		return nil, err
	}

	// Make sure that we only grab the N highest scores of all the ones posted. Yes, this is a terrible
	// algorithm for doing this and doesn't scale, but the purpose is to show how you get to structure
	// your service logic, not how to efficiently interact w/ your data store.
	sort.Sort(gameScores)
	gameScores = gameScores.Top(request.Limit())
	return &HighScoresForGameResponse{Scores: gameScores}, nil
}
