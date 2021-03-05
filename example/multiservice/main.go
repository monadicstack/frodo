package main

import (
	"context"
	"fmt"
	"log"

	"github.com/monadicstack/frodo/example/multiservice/games"
	gamesrpc "github.com/monadicstack/frodo/example/multiservice/games/gen"
	"github.com/monadicstack/frodo/example/multiservice/scores"
	scoresrpc "github.com/monadicstack/frodo/example/multiservice/scores/gen"
)

func main() {
	ctx := context.Background()
	gameClient := gamesrpc.NewGameServiceClient("http://localhost:9001")
	scoreClient := scoresrpc.NewScoreServiceClient("http://localhost:9002")

	// Operation 1: Just look up a game that's already in the database.
	game1, err := gameClient.GetByID(ctx, &games.GetByIDRequest{ID: "1"})
	exitOnError(err)
	fmt.Printf("Game 1 = %s\n", game1.Name)

	// Operation 2: Add a game to the catalog.
	game2, err := gameClient.Register(ctx, &games.RegisterRequest{
		Name:      "The Witcher 3: Wild Hunt",
		Publisher: "CD Projekt RED",
		Year:      2015,
	})
	exitOnError(err)
	fmt.Printf("Game 2 = %s\n", game2.Name)

	// Operation 3: Post a high score for super mario bros.
	score, err := scoreClient.NewHighScore(ctx, &scores.NewHighScoreRequest{
		GameID:     game1.ID,
		PlayerName: "Red Luigi",
		Score:      9393311,
	})
	exitOnError(err)
	fmt.Printf("New Score = %d\n", score.Score)

	// Operation 4: Fetch the top 3 scores for Super Mario Bros (should include the one we just posted).
	highScores, err := scoreClient.HighScoresForGame(ctx, &scores.HighScoresForGameRequest{
		GameID:  game1.ID,
		HowMany: 3,
	})
	exitOnError(err)
	for i, highScore := range highScores.Scores {
		fmt.Printf("High Score %d = %d (%s)\n", i+1, highScore.Score, highScore.PlayerName)
	}
}

func exitOnError(err error) {
	if err != nil {
		log.Fatalf(err.Error())
	}
}
