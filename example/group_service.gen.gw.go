// !!!!!!! DO NOT EDIT !!!!!!!
// Auto-generated server code from example/group_service.go
// !!!!!!! DO NOT EDIT !!!!!!!
package example

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/robsignorelli/respond"
)

func NewGroupServiceGateway(service GroupService) *GroupServiceGateway {
	gw := &GroupServiceGateway{
		service: service,
		router:  httprouter.New(),
	}

	gw.router.POST("/GroupService.CreateGroup", func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		reply := respond.To(w, req)

		request := CreateGroupRequest{}
		err := json.NewDecoder(req.Body).Decode(&request)
		if err != nil {
			reply.Fail(err)
		}

		response, err := gw.service.CreateGroup(req.Context(), &request)
		reply.Ok(response, err)
	})

	gw.router.POST("/GroupService.DeleteGroup", func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		reply := respond.To(w, req)

		request := DeleteGroupRequest{}
		err := json.NewDecoder(req.Body).Decode(&request)
		if err != nil {
			reply.Fail(err)
		}

		response, err := gw.service.DeleteGroup(req.Context(), &request)
		reply.Ok(response, err)
	})

	return gw
}

type GroupServiceGateway struct {
	// The "real" implementation of the service that this gateway delegates to.
	service GroupService
	// The underlying HTTP router/mux that wraps all of our endpoints up into a single handler.
	router *httprouter.Router
}

func (gw GroupServiceGateway) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	gw.router.ServeHTTP(w, req)
}
