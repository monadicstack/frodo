package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/davidrenne/frodo/example/basic/calc"
	calcrpc "github.com/davidrenne/frodo/example/basic/calc/gen"
	"github.com/davidrenne/frodo/example/multiservice/games"
	gamesrpc "github.com/davidrenne/frodo/example/multiservice/games/gen"
	"github.com/davidrenne/frodo/example/multiservice/scores"
	scoresrpc "github.com/davidrenne/frodo/example/multiservice/scores/gen"
	"github.com/davidrenne/frodo/rpc"
)

func main() {
	// Start both gateways, running both on port 9004
	go startGateways()
	time.Sleep(2 * time.Second)

	// In the 'multiservice/' example the clients actually used different ports because the
	// services were run in separate processes. Here, they're both run in the same one.
	ctx := context.Background()
	calcClient := calcrpc.NewCalculatorServiceClient("http://localhost:9004")
	gameClient := gamesrpc.NewGameServiceClient("http://localhost:9004")
	scoreClient := scoresrpc.NewScoreServiceClient("http://localhost:9004")

	sum, err := calcClient.Add(ctx, &calc.AddRequest{A: 10, B: 3})
	exitOnError(err)
	fmt.Printf("10 + 3 = %d\n", sum.Result)

	game1, err := gameClient.GetByID(ctx, &games.GetByIDRequest{ID: "1"})
	exitOnError(err)
	fmt.Printf("Game 1 = %s\n", game1.Name)

	highScores, err := scoreClient.HighScoresForGame(ctx, &scores.HighScoresForGameRequest{
		GameID:  game1.ID,
		HowMany: 3,
	})
	exitOnError(err)
	for i, highScore := range highScores.Scores {
		fmt.Printf("High Score %d = %d (%s)\n", i+1, highScore.Score, highScore.PlayerName)
	}
}

func startGateways() {
	// Create your live service instances how you normally would. Notice that for the score service's
	// dependency on the game service we pass in the real game service handler rather than a game
	// service client. Both the client and handler implement GameService so they are 100% interchangeable!
	calcService := calc.CalculatorServiceHandler{}
	gameService := games.GameServiceHandler{
		Repo: games.NewRepo(),
	}
	scoreService := scores.ScoreServiceHandler{
		Games: &gameService,
		Repo:  scores.NewRepo(),
	}

	// Combine the gateways for all of our services into a single gateway that routes requests to all of them.
	gateway := rpc.Compose(
		calcrpc.NewCalculatorServiceGateway(&calcService).Gateway,
		scoresrpc.NewScoreServiceGateway(&scoreService).Gateway,
		gamesrpc.NewGameServiceGateway(&gameService).Gateway,
	)
	http.ListenAndServe(":9004", gateway)
}

func exitOnError(err error) {
	if err != nil {
		log.Fatalf(err.Error())
	}
}
