package common

import (
	"net/http"
)

// Route defines a rest api
type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

// RestRequestRoutes is the routes for Rest Request
type RestRequestRoutes struct {
	routes []Route
}

// Routes gets routes
func (r *RestRequestRoutes) Routes() []Route {
	return r.routes
}

// AddToRoutes adds more routes
func (r *RestRequestRoutes) AddToRoutes(routes []Route) {
	r.routes = append(r.routes, routes...)
}
