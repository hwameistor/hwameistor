package rest

import (
	"net/http"

	"github.com/hwameistor/hwameistor/pkg/common"
)

func (rs *restServer) buildRoutes() []common.Route {

	routes := common.RestRequestRoutes{}

	routes.AddToRoutes(rs.basicRoutes())

	// add more routes

	return routes.Routes()
}

func (rs *restServer) basicRoutes() []common.Route {
	return []common.Route{
		{
			Name:        "HealthCheck",
			Method:      "GET",
			Pattern:     "/healthz",
			HandlerFunc: rs.handleHealthCheck,
		},
	}
}

func (rs *restServer) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
