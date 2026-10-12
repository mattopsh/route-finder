package routes

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"route-finder/internal/geo"
	"route-finder/internal/route"
)

type routeFinder interface {
	FindRoutes(ctx context.Context, src geo.Coordinate, dsts []geo.Coordinate) ([]route.Route, error)
}

type Handler struct {
	routes routeFinder
}

func NewHandler(routes routeFinder) *Handler {
	return &Handler{routes: routes}
}

func (h *Handler) GetRoutes(c *gin.Context) {
	src, dsts, err := parseRouteRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{
			Error: err.Error(),
		})
		return
	}

	routes, err := h.routes.FindRoutes(c.Request.Context(), src, dsts)
	if err != nil {
		status, msg := statusCodeFromError(err)
		c.JSON(status, errorResponse{Error: msg})
		return
	}

	c.JSON(http.StatusOK, newRoutesResponse(src, routes))
}
