// !!!!!!! DO NOT EDIT !!!!!!!
// Auto-generated server code from example/group_service.go
// !!!!!!! DO NOT EDIT !!!!!!!
package example

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/robsignorelli/expose/binding"
	"github.com/robsignorelli/respond"
)

func NewGroupServiceGateway(service GroupService) *GroupServiceGateway {
	gw := &GroupServiceGateway{
		service: service,
		router:  httprouter.New(),
	}

	gw.router.POST("/GroupService.CreateGroup", func(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
		response := respond.To(w, req)

		serviceRequest := CreateGroupRequest{}
		if err := binding.Bind(req, params, &serviceRequest); err != nil {
			response.Fail(err)
			return
		}

		serviceResponse, err := gw.service.CreateGroup(req.Context(), &serviceRequest)
		response.Ok(serviceResponse, err)
	})

	gw.router.DELETE("/group/:ID", func(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
		response := respond.To(w, req)

		serviceRequest := DeleteGroupRequest{}
		if err := binding.Bind(req, params, &serviceRequest); err != nil {
			response.Fail(err)
			return
		}

		serviceResponse, err := gw.service.DeleteGroup(req.Context(), &serviceRequest)
		response.Ok(serviceResponse, err)
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
