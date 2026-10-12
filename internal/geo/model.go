package geo

import (
	"fmt"
	"strconv"
	"strings"
)

type Coordinate struct {
	Lat string
	Lng string
}

func (c Coordinate) String() string {
	return c.Lng + "," + c.Lat
}

func ParseCoordinate(s string) (Coordinate, error) {
	parts := strings.SplitN(s, ",", 2)
	if len(parts) != 2 {
		return Coordinate{}, fmt.Errorf("invalid coordinate %q: want lng,lat", s)
	}

	lng, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return Coordinate{}, fmt.Errorf("invalid longitude %q: %w", parts[0], err)
	}
	if lng < -180 || lng > 180 {
		return Coordinate{}, fmt.Errorf("longitude %q out of range [-180,180]", parts[0])
	}

	lat, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return Coordinate{}, fmt.Errorf("invalid latitude %q: %w", parts[1], err)
	}
	if lat < -90 || lat > 90 {
		return Coordinate{}, fmt.Errorf("latitude %q out of range [-90,90]", parts[1])
	}

	return Coordinate{Lng: parts[0], Lat: parts[1]}, nil
}
