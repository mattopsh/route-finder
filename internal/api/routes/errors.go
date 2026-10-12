package routes

import (
	"errors"
	"net/http"

	"route-finder/internal/route"
)

func statusCodeFromError(err error) (int, string) {
	switch {
	case errors.Is(err, route.ErrNoRoute):
		return http.StatusNotFound, "no route found"
	case errors.Is(err, route.ErrUpstreamTimeout):
		return http.StatusGatewayTimeout, "upstream timeout"
	case errors.Is(err, route.ErrUpstream):
		return http.StatusBadGateway, "upstream error"
	default:
		return http.StatusInternalServerError, "internal error"
	}
}
