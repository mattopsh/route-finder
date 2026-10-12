package osrm

import (
	"route-finder/internal/geo"
	"route-finder/internal/route"
)

const (
	codeOK              = "Ok"
	codeInvalidUrl      = "InvalidUrl"
	codeInvalidService  = "InvalidService"
	codeInvalidVersion  = "InvalidVersion"
	codeInvalidOptions  = "InvalidOptions"
	codeInvalidQuery    = "InvalidQuery"
	codeInvalidValue    = "InvalidValue"
	codeNoSegment       = "NoSegment"
	codeTooBig          = "TooBig"
	codeDisabledDataset = "DisabledDataset"
	codeNoRoute         = "NoRoute"
)

func toRoute(to geo.Coordinate, r routeResult) route.Route {
	return route.Route{
		To:       to,
		Duration: r.Duration,
		Distance: r.Distance,
	}
}
