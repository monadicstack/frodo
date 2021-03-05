package games

import (
	"context"
	"fmt"
	"time"

	"github.com/monadicstack/frodo/rpc/errors"
)

// Repo manages access to the data store where we keep game catalog data.
type Repo interface {
	// GetByID fetches a game record given its unique identifier. If the game is not found,
	// you will receive nil and a 404-style error value back.
	GetByID(ctx context.Context, id string) (Game, error)
	// Create adds another game record to the store.
	Create(ctx context.Context, game Game) (Game, error)
}

// NewRepo constructs a new mock repo that stores high score data in-memory. Not what you'd
// use in production, but it serves us well for this example code.
func NewRepo() Repo {
	games := []Game{
		{
			ID:          "1",
			Name:        "Super Mario Bros.",
			Description: "Luigi and Red Luigi try to save the Princess.",
			Publisher:   "Nintendo",
			Year:        1985,
		},
		{
			ID:          "2",
			Name:        "Metroid",
			Description: "Objectify a strong woman. If you beat the game quickly enough your reward is seeing her in a bikini.",
			Publisher:   "Nintendo",
			Year:        1986,
		},
		{
			ID:          "3",
			Name:        "Dark Souls",
			Description: "Git gud.",
			Publisher:   "From Software",
			Year:        2011,
		},
	}
	return &mockRepo{Games: games}
}

type mockRepo struct {
	Games []Game
}

func (m mockRepo) GetByID(_ context.Context, id string) (Game, error) {
	for _, game := range m.Games {
		if game.ID == id {
			return game, nil
		}
	}
	return Game{}, errors.NotFound("game not found: %s", id)
}

func (m *mockRepo) Create(_ context.Context, game Game) (Game, error) {
	if game.ID != "" {
		return game, errors.BadRequest("create: game already exists")
	}
	if game.Name == "" {
		return game, errors.BadRequest("create: name is required")
	}

	// I realize this is *terrible* client-side ID generation, but that's not really the purpose of this
	// example. It's more about showing you how to wire up your services and less about the finder details of
	// interacting with your data store.
	game.ID = fmt.Sprintf("%v", time.Now().UnixNano())
	m.Games = append(m.Games, game)
	return game, nil
}
