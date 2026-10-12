package routes

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"route-finder/internal/route"
)

func TestStatusCodeFromError(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
		msg    string
	}{
		{name: "no route", err: route.ErrNoRoute, status: http.StatusNotFound, msg: "no route found"},
		{name: "upstream timeout", err: route.ErrUpstreamTimeout, status: http.StatusGatewayTimeout, msg: "upstream timeout"},
		{name: "upstream", err: route.ErrUpstream, status: http.StatusBadGateway, msg: "upstream error"},
		{name: "wrapped no route", err: fmt.Errorf("route to 13.4,52.5: %w", route.ErrNoRoute), status: http.StatusNotFound, msg: "no route found"},
		{name: "wrapped upstream timeout", err: fmt.Errorf("route to 13.4,52.5: %w", route.ErrUpstreamTimeout), status: http.StatusGatewayTimeout, msg: "upstream timeout"},
		{name: "wrapped upstream", err: fmt.Errorf("route to 13.4,52.5: %w", route.ErrUpstream), status: http.StatusBadGateway, msg: "upstream error"},
		{name: "unknown", err: errors.New("boom"), status: http.StatusInternalServerError, msg: "internal error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, msg := statusCodeFromError(tt.err)
			assert.Equal(t, tt.status, status)
			assert.Equal(t, tt.msg, msg)
		})
	}
}
