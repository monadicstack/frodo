// Code generated by Frodo - DO NOT EDIT.
//
//   Timestamp: Tue, 10 May 2022 16:25:34 EDT
//   Source:    games/game_service.go
//   Generator: https://github.com/monadicstack/frodo
//
package games

import (
	"context"
	"fmt"

	"github.com/monadicstack/frodo/example/multiservice/games"
	"github.com/monadicstack/frodo/rpc"
)

// NewGameServiceClient creates an RPC client that conforms to the GameService interface, but delegates
// work to remote instances. You must supply the base address of the remote service gateway instance or
// the load balancer for that service.
//
// GameService manages the catalog info for all of the games in our system.
func NewGameServiceClient(address string, options ...rpc.ClientOption) *GameServiceClient {
	rpcClient := rpc.NewClient("GameService", address, options...)
	rpcClient.PathPrefix = "/v2"
	return &GameServiceClient{Client: rpcClient}
}

// GameServiceClient manages all interaction w/ a remote GameService instance by letting you invoke functions
// on this instance as if you were doing it locally (hence... RPC client). You shouldn't instantiate this
// manually. Instead, you should utilize the NewGameServiceClient() function to properly set this up.
type GameServiceClient struct {
	rpc.Client
}

// GetByID looks up a game record given its unique id.
func (client *GameServiceClient) GetByID(ctx context.Context, request *games.GetByIDRequest) (*games.GetByIDResponse, error) {
	if ctx == nil {
		return nil, fmt.Errorf("precondition failed: nil context")
	}
	if request == nil {
		return nil, fmt.Errorf("precondition failed: nil request")
	}

	response := &games.GetByIDResponse{}
	err := client.Invoke(ctx, "GET", "/game/:ID", request, response)
	return response, err
}

// Register adds another game record to our gaming database.
func (client *GameServiceClient) Register(ctx context.Context, request *games.RegisterRequest) (*games.RegisterResponse, error) {
	if ctx == nil {
		return nil, fmt.Errorf("precondition failed: nil context")
	}
	if request == nil {
		return nil, fmt.Errorf("precondition failed: nil request")
	}

	response := &games.RegisterResponse{}
	err := client.Invoke(ctx, "POST", "/game", request, response)
	return response, err
}

// GameServiceProxy fully implements the GameService interface, but delegates all operations to a "real"
// instance of the service. You can embed this type in a struct of your choice so you can "override" or
// decorate operations as you see fit. Any operations on GameService that you don't explicitly define will
// simply delegate to the default implementation of the underlying 'Service' value.
//
// Since you have access to the underlying service, you are able to both implement custom handling logic AND
// call the "real" implementation, so this can be used as special middleware that applies to only certain operations.
type GameServiceProxy struct {
	Service games.GameService
}

func (proxy *GameServiceProxy) GetByID(ctx context.Context, request *games.GetByIDRequest) (*games.GetByIDResponse, error) {
	return proxy.Service.GetByID(ctx, request)
}

func (proxy *GameServiceProxy) Register(ctx context.Context, request *games.RegisterRequest) (*games.RegisterResponse, error) {
	return proxy.Service.Register(ctx, request)
}
