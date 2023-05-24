// Code generated by Frodo - DO NOT EDIT.
//
//   Timestamp: Tue, 10 May 2022 16:25:34 EDT
//   Source:    scores/score_service.go
//   Generator: https://github.com/davidrenne/frodo
//
package scores

import (
	"context"
	"net/http"

	"github.com/davidrenne/frodo/example/multiservice/scores"
	"github.com/davidrenne/frodo/rpc"
	"github.com/monadicstack/respond"
)

// NewScoreServiceGateway accepts your "real" ScoreService instance (the thing that really does the work), and
// exposes it to other services/clients over RPC. The rpc.Gateway it returns implements http.Handler, so you
// can pass it to any standard library HTTP server of your choice.
//
//	// How to fire up your service for RPC and/or your REST API
//	service := scores.ScoreService{ /* set up to your liking */ }
//	gateway := scores.NewScoreServiceGateway(service)
//	http.ListenAndServe(":8080", gateway)
//
// The default instance works well enough, but you can supply additional options such as WithMiddleware() which
// accepts any negroni-compatible middleware handlers.
func NewScoreServiceGateway(service scores.ScoreService, options ...rpc.GatewayOption) ScoreServiceGateway {
	gw := rpc.NewGateway(options...)
	gw.Name = "ScoreService"
	gw.PathPrefix = "/v2"

	gw.Register(rpc.Endpoint{
		Method:      "GET",
		Path:        "/game/:GameID/highscore",
		ServiceName: "ScoreService",
		Name:        "HighScoresForGame",
		Handler: func(w http.ResponseWriter, req *http.Request) {
			response := respond.To(w, req)

			serviceRequest := scores.HighScoresForGameRequest{}
			if err := gw.Binder.Bind(req, &serviceRequest); err != nil {
				response.Fail(err)
				return
			}

			serviceResponse, err := service.HighScoresForGame(req.Context(), &serviceRequest)
			response.Reply(200, serviceResponse, err)
		},
	})

	gw.Register(rpc.Endpoint{
		Method:      "POST",
		Path:        "/game/:GameID/highscore",
		ServiceName: "ScoreService",
		Name:        "NewHighScore",
		Handler: func(w http.ResponseWriter, req *http.Request) {
			response := respond.To(w, req)

			serviceRequest := scores.NewHighScoreRequest{}
			if err := gw.Binder.Bind(req, &serviceRequest); err != nil {
				response.Fail(err)
				return
			}

			serviceResponse, err := service.NewHighScore(req.Context(), &serviceRequest)
			response.Reply(201, serviceResponse, err)
		},
	})

	return ScoreServiceGateway{Gateway: gw, service: service}
}

type ScoreServiceGateway struct {
	rpc.Gateway
	service scores.ScoreService
}

func (gw ScoreServiceGateway) HighScoresForGame(ctx context.Context, request *scores.HighScoresForGameRequest) (*scores.HighScoresForGameResponse, error) {
	return gw.service.HighScoresForGame(ctx, request)
}

func (gw ScoreServiceGateway) NewHighScore(ctx context.Context, request *scores.NewHighScoreRequest) (*scores.NewHighScoreResponse, error) {
	return gw.service.NewHighScore(ctx, request)
}

func (gw ScoreServiceGateway) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	gw.Gateway.ServeHTTP(w, req)
}
