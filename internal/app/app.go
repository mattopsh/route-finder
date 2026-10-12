package app

import (
	"net/http"
	"route-finder/internal/adapter/osrm"
	"route-finder/internal/api"
	"route-finder/internal/api/routes"
	"route-finder/internal/config"
	"route-finder/internal/route"
	"time"

	"github.com/gin-gonic/gin"
)

func New(cfg config.Config) (*gin.Engine, error) {
	osrmHttpClient := &http.Client{Timeout: time.Duration(cfg.OSRMTimeoutMs) * time.Millisecond}
	client := osrm.NewClient(cfg.OSRMBaseURL, osrmHttpClient)
	svc := route.NewService(client, cfg.OSRMMaxConcurrency)
	handler := routes.NewHandler(svc)
	return api.NewRouter(handler), nil
}
