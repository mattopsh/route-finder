package routes

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"route-finder/internal/geo"
)

const maxDstCount = 50

func parseRouteRequest(c *gin.Context) (geo.Coordinate, []geo.Coordinate, error) {
	srcRaw := c.Query("src")
	if srcRaw == "" {
		return geo.Coordinate{}, nil,
			fmt.Errorf("missing src parameter")
	}

	src, err := geo.ParseCoordinate(srcRaw)
	if err != nil {
		return geo.Coordinate{}, nil, err
	}

	dstRaws := c.QueryArray("dst")
	if len(dstRaws) == 0 {
		return geo.Coordinate{}, nil,
			fmt.Errorf("missing dst parameter")
	}
	if len(dstRaws) > maxDstCount {
		return geo.Coordinate{}, nil,
			fmt.Errorf("too many dst parameters: got %d, max %d", len(dstRaws), maxDstCount)
	}

	dsts := make([]geo.Coordinate, len(dstRaws))
	for i, raw := range dstRaws {
		dst, err := geo.ParseCoordinate(raw)
		if err != nil {
			return geo.Coordinate{}, nil, fmt.Errorf("invalid dst parameter at index %d: %w", i, err)
		}
		dsts[i] = dst
	}

	return src, dsts, nil
}
