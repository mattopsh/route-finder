package routes

import (
	"route-finder/internal/geo"
	"route-finder/internal/route"
)

type routesResponse struct {
	Source string      `json:"source"`
	Routes []routeItem `json:"routes"`
}

type routeItem struct {
	Destination string  `json:"destination"`
	Duration    float64 `json:"duration"`
	Distance    float64 `json:"distance"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func newRoutesResponse(src geo.Coordinate, routes []route.Route) routesResponse {
	resp := routesResponse{
		Source: src.String(),
		Routes: make([]routeItem, len(routes)),
	}

	for i, r := range routes {
		resp.Routes[i] = routeItem{
			Destination: r.To.String(),
			Duration:    r.Duration,
			Distance:    r.Distance,
		}
	}

	return resp
}
