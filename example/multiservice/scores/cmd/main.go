package main

import (
	"net/http"

	games "github.com/monadicstack/frodo/example/multiservice/games/gen"
	"github.com/monadicstack/frodo/example/multiservice/scores"
	scoresrpc "github.com/monadicstack/frodo/example/multiservice/scores/gen"
)

func main() {
	serviceHandler := scores.ScoreServiceHandler{
		Games: games.NewGameServiceClient("http://localhost:9001"),
		Repo:  scores.NewRepo(),
	}
	gateway := scoresrpc.NewScoreServiceGateway(&serviceHandler)
	http.ListenAndServe(":9002", gateway)
}
