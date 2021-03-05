package games

import (
	"context"
)

// GameService manages the catalog info for all of the games in our system.
//
// PATH /v2
type GameService interface {
	// GetByID looks up a game record given its unique id.
	//
	// GET /game/:ID
	GetByID(context.Context, *GetByIDRequest) (*GetByIDResponse, error)

	// Register adds another game record to our gaming database.
	//
	// HTTP 201
	// POST /game
	Register(context.Context, *RegisterRequest) (*RegisterResponse, error)
}

// Game is the snapshot of data that represents a single game that we support in our gaming cloud.
type Game struct {
	// ID is the unique identifier for the game.
	ID string `json:"id"`
	// Name is the title of the game (e.g. "Minecraft")
	Name string `json:"name"`
	// Description contains the long-form details of the game.
	Description string `json:"description"`
	// Publisher indicates who made/released the game.
	Publisher string `json:"publisher"`
	// Year is when the game was initially released.
	Year int `json:"year"`
}

// GetByIDRequest contains all of the inputs for the GetByID operation.
type GetByIDRequest struct {
	// ID is the unique identifier of the game to look up.
	ID string `json:"id"`
}

// GetByIDResponse is the record info for the game that we found.
//
// Frodo Notes: Notice that you don't need need to define a whole new struct. You can
// have any of your request/response types simply alias some shared type. You don't need
// to repeat yourself over and over. For explicitness, it's still good to create this alias
// rather than having GetByID() return a Game type.
type GetByIDResponse Game

// RegisterRequest contains all of the inputs used to populate a new game record.
type RegisterRequest struct {
	// Name is the title of the game (e.g. "Minecraft")
	Name string `json:"name"`
	// Description contains the long-form details of the game.
	Description string `json:"description"`
	// Publisher indicates who made/released the game.
	Publisher string `json:"publisher"`
	// Year is when the game was initially released.
	Year int `json:"year"`
}

// RegisterResponse is the complete game record as it was saved, with the ID it was assigned.
type RegisterResponse Game
