package route

import (
	"route-finder/internal/geo"
)

type Route struct {
	To       geo.Coordinate
	Duration float64
	Distance float64
}
