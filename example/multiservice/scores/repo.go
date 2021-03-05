package scores

import (
	"context"
	"time"

	"github.com/monadicstack/frodo/rpc/errors"
)

// Repo manages access to the data store where we keep high score data.
type Repo interface {
	// Create adds a high score record to the store.
	Create(ctx context.Context, score HighScore) (HighScore, error)
	// ByGame searches for a list of all scores posted for the given game.
	ByGame(ctx context.Context, gameID string) (HighScoreList, error)
}

// NewRepo constructs a new mock repo that stores high score data in-memory. Not what you'd
// use in production, but it serves us well for this example code.
func NewRepo() Repo {
	now := time.Now()
	return &mockRepo{
		scores: []HighScore{
			{GameID: "1", PlayerName: "Dog Man", Score: 20332, Date: now.Add(-50 * time.Hour)},
			{GameID: "1", PlayerName: "Wolverine", Score: 999, Date: now.Add(-40 * time.Hour)},
			{GameID: "1", PlayerName: "Dog Man", Score: 899, Date: now.Add(-30 * time.Hour)},
			{GameID: "3", PlayerName: "Wolverine", Score: 6, Date: now.Add(-20 * time.Hour)},
			{GameID: "3", PlayerName: "Cyclops", Score: 3, Date: now.Add(-15 * time.Hour)},
			{GameID: "3", PlayerName: "Cyclops", Score: 2, Date: now.Add(-10 * time.Hour)},
			{GameID: "3", PlayerName: "Wolverine", Score: 8, Date: now.Add(-5 * time.Hour)},
		},
	}
}

type mockRepo struct {
	scores HighScoreList
}

func (m *mockRepo) Create(_ context.Context, score HighScore) (HighScore, error) {
	if score.GameID == "" {
		return score, errors.BadRequest("high score: game id is required")
	}
	if score.PlayerName == "" {
		return score, errors.BadRequest("high score: player name is required")
	}

	m.scores = append(m.scores, score)
	return score, nil
}

func (m mockRepo) ByGame(_ context.Context, gameID string) (HighScoreList, error) {
	results := HighScoreList{}
	for _, score := range m.scores {
		if score.GameID == gameID {
			results = append(results, score)
		}
	}
	return results, nil
}

// HighScoreList is a slice of high scores with sorting and sub-slicing behavior baked in.
type HighScoreList []HighScore

// Len returns the number of scores in this list.
func (scores HighScoreList) Len() int {
	return len(scores)
}

// Less returns true when the score at index 'i' is actually greater than the one at
// index 'j' since we sort the list from highest score to lowest.
func (scores HighScoreList) Less(i, j int) bool {
	return scores[i].Score > scores[j].Score
}

// Swap provides 'sort' package support by switching the values at indices i and j.
func (scores HighScoreList) Swap(i, j int) {
	tmp := scores[i]
	scores[i] = scores[j]
	scores[j] = tmp
}

// Top returns the first "n" high scores in the list. If the list is shorter than 'howMany'
// then this will return the entire list.
func (scores HighScoreList) Top(howMany int) HighScoreList {
	if howMany >= len(scores) {
		return scores
	}
	return scores[0:howMany]
}
