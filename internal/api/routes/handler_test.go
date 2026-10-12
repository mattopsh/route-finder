package routes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"route-finder/internal/geo"
	"route-finder/internal/route"
)

type mockRouteFinder struct {
	mock.Mock
}

func (m *mockRouteFinder) FindRoutes(ctx context.Context, src geo.Coordinate, dsts []geo.Coordinate) ([]route.Route, error) {
	args := m.Called(ctx, src, dsts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]route.Route), args.Error(1)
}

func serveRoutes(t *testing.T, finder routeFinder, target string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/routes", NewHandler(finder).GetRoutes)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, target, nil))
	return w
}

func TestGetRoutesSuccess(t *testing.T) {
	src := geo.Coordinate{Lng: "13.4", Lat: "52.5"}
	dstA := geo.Coordinate{Lng: "13.5", Lat: "52.6"}
	dstB := geo.Coordinate{Lng: "13.6", Lat: "52.7"}

	finder := new(mockRouteFinder)
	finder.On("FindRoutes", mock.Anything, src, []geo.Coordinate{dstA, dstB}).
		Return([]route.Route{
			{To: dstB, Duration: 10, Distance: 20},
			{To: dstA, Duration: 30, Distance: 40},
		}, nil).
		Once()

	w := serveRoutes(t, finder, "/routes?src=13.4,52.5&dst=13.5,52.6&dst=13.6,52.7")

	require.Equal(t, http.StatusOK, w.Code)

	var resp routesResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "13.4,52.5", resp.Source)
	require.Len(t, resp.Routes, 2)
	assert.Equal(t, routeItem{Destination: "13.6,52.7", Duration: 10, Distance: 20}, resp.Routes[0])
	assert.Equal(t, routeItem{Destination: "13.5,52.6", Duration: 30, Distance: 40}, resp.Routes[1])
	finder.AssertExpectations(t)
}

func TestGetRoutesRequestValidation(t *testing.T) {
	finder := new(mockRouteFinder)

	tests := []struct {
		name   string
		target string
		errMsg string
	}{
		{name: "missing src", target: "/routes?dst=13.5,52.6", errMsg: "missing src parameter"},
		{name: "missing dst", target: "/routes?src=13.4,52.5", errMsg: "missing dst parameter"},
		{name: "invalid src", target: "/routes?src=oops&dst=13.5,52.6", errMsg: `invalid coordinate "oops": want lng,lat`},
		{name: "invalid src latitude range", target: "/routes?src=13.4,91&dst=13.5,52.6", errMsg: `latitude "91" out of range [-90,90]`},
		{name: "invalid dst", target: "/routes?src=13.4,52.5&dst=oops", errMsg: `invalid dst parameter at index 0: invalid coordinate "oops": want lng,lat`},
		{name: "invalid dst at index", target: "/routes?src=13.4,52.5&dst=13.5,52.6&dst=13.7,bad", errMsg: `invalid dst parameter at index 1: invalid latitude "bad": strconv.ParseFloat: parsing "bad": invalid syntax`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := serveRoutes(t, finder, tt.target)

			require.Equal(t, http.StatusBadRequest, w.Code)

			var resp errorResponse
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
			assert.Equal(t, tt.errMsg, resp.Error)
			finder.AssertNotCalled(t, "FindRoutes", mock.Anything, mock.Anything, mock.Anything)
		})
	}
}

func dstParams(n int) string {
	params := make([]string, n)
	for i := range params {
		params[i] = "dst=13.5,52.6"
	}
	return strings.Join(params, "&")
}

func TestGetRoutesDstCountLimit(t *testing.T) {
	t.Run("rejects over limit", func(t *testing.T) {
		finder := new(mockRouteFinder)

		w := serveRoutes(t, finder, "/routes?src=13.4,52.5&"+dstParams(maxDstCount+1))

		require.Equal(t, http.StatusBadRequest, w.Code)

		var resp errorResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, fmt.Sprintf("too many dst parameters: got %d, max %d", maxDstCount+1, maxDstCount), resp.Error)
		finder.AssertNotCalled(t, "FindRoutes", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("accepts exactly max", func(t *testing.T) {
		src := geo.Coordinate{Lng: "13.4", Lat: "52.5"}
		dst := geo.Coordinate{Lng: "13.5", Lat: "52.6"}
		dsts := make([]geo.Coordinate, maxDstCount)
		for i := range dsts {
			dsts[i] = dst
		}

		finder := new(mockRouteFinder)
		finder.On("FindRoutes", mock.Anything, src, dsts).Return(make([]route.Route, maxDstCount), nil).Once()

		w := serveRoutes(t, finder, "/routes?src=13.4,52.5&"+dstParams(maxDstCount))

		require.Equal(t, http.StatusOK, w.Code)
		finder.AssertExpectations(t)
	})
}

func TestGetRoutesServiceErrors(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
		msg    string
	}{
		{name: "no route", err: route.ErrNoRoute, status: http.StatusNotFound, msg: "no route found"},
		{name: "upstream timeout", err: route.ErrUpstreamTimeout, status: http.StatusGatewayTimeout, msg: "upstream timeout"},
		{name: "upstream", err: route.ErrUpstream, status: http.StatusBadGateway, msg: "upstream error"},
		{name: "wrapped no route", err: errors.Join(errors.New("route to 13.5,52.6"), route.ErrNoRoute), status: http.StatusNotFound, msg: "no route found"},
		{name: "unknown", err: errors.New("boom"), status: http.StatusInternalServerError, msg: "internal error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			finder := new(mockRouteFinder)
			finder.On("FindRoutes", mock.Anything, mock.Anything, mock.Anything).
				Return(nil, tt.err).
				Once()

			w := serveRoutes(t, finder, "/routes?src=13.4,52.5&dst=13.5,52.6")

			require.Equal(t, tt.status, w.Code)

			var resp errorResponse
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
			assert.Equal(t, tt.msg, resp.Error)
			finder.AssertExpectations(t)
		})
	}
}
