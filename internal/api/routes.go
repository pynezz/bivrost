package api

import (
	"github.com/pynezz/bivrost/internal/config"
)

// path: internal/api/routes.go

func (a *app) addRoute(route string, cfg *config.Config) {
	return Route{
		Method:  *a.App.Get(route),
		Path:    route,
		Handler: getThreatsHandler,
	}
}
