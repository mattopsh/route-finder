package api

import (
	"route-finder/internal/api/routes"

	"github.com/gin-gonic/gin"
)

func NewRouter(h *routes.Handler) *gin.Engine {
	r := gin.Default()
	err := r.SetTrustedProxies(nil)
	if err != nil {
		return nil
	}
	r.GET("/routes", h.GetRoutes)
	return r
}
