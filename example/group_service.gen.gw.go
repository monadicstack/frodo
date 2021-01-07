// !!!!!!! DO NOT EDIT !!!!!!!
// Auto-generated server code from example/group_service.go
// !!!!!!! DO NOT EDIT !!!!!!!
package example

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/robsignorelli/expose/gateway"
	"github.com/robsignorelli/respond"
)

func NewGroupServiceGateway(service GroupService, options ...gateway.Option) *GroupServiceGateway {
	gw := &GroupServiceGateway{
		HTTPGateway: gateway.New(options...),
		Service:     service,
	}

	gw.Router.POST("/GroupService.CreateGroup", func(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
		response := respond.To(w, req)

		serviceRequest := CreateGroupRequest{}
		if err := gw.Binder.Bind(req, params, &serviceRequest); err != nil {
			response.Fail(err)
			return
		}

		serviceResponse, err := gw.Service.CreateGroup(req.Context(), &serviceRequest)
		response.Reply(200, serviceResponse, err)
	})

	gw.Router.DELETE("/group/:ID", func(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
		response := respond.To(w, req)

		serviceRequest := DeleteGroupRequest{}
		if err := gw.Binder.Bind(req, params, &serviceRequest); err != nil {
			response.Fail(err)
			return
		}

		serviceResponse, err := gw.Service.DeleteGroup(req.Context(), &serviceRequest)
		response.Reply(202, serviceResponse, err)
	})

	return gw
}

type GroupServiceGateway struct {
	gateway.HTTPGateway
	// The "real" implementation of the service that this gateway delegates to.
	Service GroupService
}

func (gw GroupServiceGateway) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	gw.Middleware.ServeHTTP(w, req, gw.Router.ServeHTTP)
}
