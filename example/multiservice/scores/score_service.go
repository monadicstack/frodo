package scores

import (
	"context"
	"time"
)

// ScoreService is a shared leaderboard service that tracks the high scores that people have
// achieved while playing various games.
//
// PREFIX /v2
type ScoreService interface {
	// NewHighScore captures a player's high score for the given game.
	//
	// HTTP 201
	// POST /game/:GameID/highscore
	NewHighScore(context.Context, *NewHighScoreRequest) (*NewHighScoreResponse, error)

	// HighScoresForGame fetches the top "N" high scores achieved by any player
	// for the specified game. If you don't specify the HowMany value, this will default
	// to returning the top 5 scores.
	//
	// Frodo Notes: The request has 2 attributes. The GameID field will be populated via the
	// path parameters, but HowMany will be specified via the query string. The auto-generated
	// clients will do this by default under the hood, but if you use curl/Postman, that's how
	// you would supply that (e.g. "/v2/game/2/highscore?howMany=3" for the top 3).
	//
	// GET /game/:GameID/highscore
	HighScoresForGame(context.Context, *HighScoresForGameRequest) (*HighScoresForGameResponse, error)
}

// HighScore tracks a single high score that a player achieved for a specific game. A single player can have
// multiple instances of this for the same game since they can continue to beat their scores.
type HighScore struct {
	// GameID is the identifier of the game that we're tracking the score for.
	GameID string `json:"gameID"`
	// PlayerName is the handle of the player who got the score. In a *real* system, this would
	// likely be a PlayerID instead, but to keep this example a bit more simple, I didn't want to
	// add a third service into the mix. But you would treat it exactly like the game service if
	// you wanted to track player data separately.
	PlayerName string `json:"playerName"`
	// Score is the numeric score that the player attained. This score may mean different things for
	// different games. For Tetris, it is the actual score, but for something like Dark Souls maybe it
	// is how many New Game + runs did you complete before dying.
	Score uint64 `json:"score"`
	// Date is the date/time that the player attained this score.
	Date time.Time `json:"date"`
}

// NewHighScoreRequest captures all of the data fields required to post a new high score.
type NewHighScoreRequest struct {
	// GameID is the identifier of the game that we're tracking the score for.
	GameID string `json:"gameID"`
	// PlayerName is the handle of the player who got the score. In a *real* system, this would
	// likely be a PlayerID instead, but to keep this example a bit more simple, I didn't want to
	// add a third service into the mix. But you would treat it exactly like the game service if
	// you wanted to track player data separately.
	PlayerName string `json:"playerName"`
	// Score is the numeric score that the player attained. This score may mean different things for
	// different games. For Tetris, it is the actual score, but for something like Dark Souls maybe it
	// is how many New Game + runs did you complete before dying.
	Score uint64 `json:"score"`
}

// NewHighScoreResponse is the complete record of the high score that you just posted.
type NewHighScoreResponse HighScore

// HighScoresForGameRequest contains the attributes used to filter high scores for a specific game.
type HighScoresForGameRequest struct {
	// GameID is the identifier of the game that we're looking up scores for.
	GameID string `json:"gameID"`
	// HowMany limits the results to just the top N scores. (Default=5)
	HowMany int `json:"howMany"`
}

// Limit decides if we should use the default list size of 5 or if we should use
// the 'HowMany' you specified in the request to cap the resulting high score list.
func (req HighScoresForGameRequest) Limit() int {
	if req.HowMany == 0 {
		return 5
	}
	return req.HowMany
}

// HighScoresForGameResponse contains the list of high scores for the given game.
type HighScoresForGameResponse struct {
	Scores []HighScore `json:"scores"`
}
