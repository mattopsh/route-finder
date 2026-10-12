package osrm

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"route-finder/internal/geo"
	"route-finder/internal/route"
)

type mockHTTPClient struct {
	mock.Mock
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	resp, _ := args.Get(0).(*http.Response)
	return resp, args.Error(1)
}

func newHTTPResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

const okBody = `{"code":"Ok","routes":[{"duration":465.2,"distance":1879.4}]}`

var (
	srcCoord = geo.Coordinate{Lng: "13.388860", Lat: "52.517037"}
	dstCoord = geo.Coordinate{Lng: "13.397634", Lat: "52.529407"}
)

func TestGetRouteResponseCodes(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		body    string
		wantErr error
	}{
		{name: "ok", status: http.StatusOK, body: okBody},
		{name: "ok without routes", status: http.StatusOK, body: `{"code":"Ok","routes":[]}`, wantErr: route.ErrUpstream},
		{name: "no route", status: http.StatusBadRequest, body: `{"code":"NoRoute","message":"No route found"}`, wantErr: route.ErrNoRoute},
		{name: "invalid url", status: http.StatusBadRequest, body: `{"code":"InvalidUrl"}`, wantErr: route.ErrUpstream},
		{name: "invalid service", status: http.StatusBadRequest, body: `{"code":"InvalidService"}`, wantErr: route.ErrUpstream},
		{name: "invalid version", status: http.StatusBadRequest, body: `{"code":"InvalidVersion"}`, wantErr: route.ErrUpstream},
		{name: "invalid options", status: http.StatusBadRequest, body: `{"code":"InvalidOptions"}`, wantErr: route.ErrUpstream},
		{name: "invalid query", status: http.StatusBadRequest, body: `{"code":"InvalidQuery"}`, wantErr: route.ErrUpstream},
		{name: "too big", status: http.StatusBadRequest, body: `{"code":"TooBig"}`, wantErr: route.ErrUpstream},
		{name: "disabled dataset", status: http.StatusBadRequest, body: `{"code":"DisabledDataset"}`, wantErr: route.ErrUpstream},
		{name: "unknown code", status: http.StatusBadRequest, body: `{"code":"SomethingNew"}`, wantErr: route.ErrUpstream},
		{name: "unexpected http status", status: http.StatusServiceUnavailable, body: `{"code":"Ok"}`, wantErr: route.ErrUpstream},
		{name: "undecodable body", status: http.StatusBadRequest, body: `<html>`, wantErr: route.ErrUpstream},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpClient := new(mockHTTPClient)
			httpClient.On("Do", mock.Anything).Return(newHTTPResponse(tt.status, tt.body), nil).Once()
			client := NewClient("http://osrm.test", httpClient)

			got, err := client.GetRoute(context.Background(), srcCoord, dstCoord)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, dstCoord, got.To)
				assert.Equal(t, 465.2, got.Duration)
				assert.Equal(t, 1879.4, got.Distance)
			}
			httpClient.AssertExpectations(t)
		})
	}
}

func TestGetRouteErrorMessages(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		body       string
		wantErrMsg string
	}{
		{name: "includes code and message", status: http.StatusBadRequest, body: `{"code":"NoRoute","message":"No route found"}`, wantErrMsg: "code=NoRoute message=No route found"},
		{name: "includes code without message", status: http.StatusBadRequest, body: `{"code":"InvalidQuery"}`, wantErrMsg: "code=InvalidQuery"},
		{name: "ok without routes describes empty result", status: http.StatusOK, body: `{"code":"Ok","routes":[]}`, wantErrMsg: "code=Ok but no routes returned"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpClient := new(mockHTTPClient)
			httpClient.On("Do", mock.Anything).Return(newHTTPResponse(tt.status, tt.body), nil).Once()
			client := NewClient("http://osrm.test", httpClient)

			_, err := client.GetRoute(context.Background(), srcCoord, dstCoord)

			require.Error(t, err)
			assert.ErrorContains(t, err, tt.wantErrMsg)
			httpClient.AssertExpectations(t)
		})
	}
}

func TestGetRouteBuildsRequestURL(t *testing.T) {
	var captured *http.Request
	httpClient := new(mockHTTPClient)
	httpClient.On("Do", mock.Anything).
		Run(func(args mock.Arguments) {
			captured = args.Get(0).(*http.Request)
		}).
		Return(newHTTPResponse(http.StatusOK, okBody), nil).
		Once()
	client := NewClient("http://osrm.test/", httpClient)

	_, err := client.GetRoute(context.Background(), srcCoord, dstCoord)

	require.NoError(t, err)
	require.NotNil(t, captured)
	assert.Equal(t, http.MethodGet, captured.Method)
	assert.Equal(t, "osrm.test", captured.Host)
	assert.Equal(t, "/route/v1/driving/13.388860,52.517037;13.397634,52.529407", captured.URL.Path)
	assert.Equal(t, "false", captured.URL.Query().Get("overview"))
	httpClient.AssertExpectations(t)
}

func TestGetRoutePropagatesContext(t *testing.T) {
	type ctxKey struct{}

	var captured *http.Request
	httpClient := new(mockHTTPClient)
	httpClient.On("Do", mock.Anything).
		Run(func(args mock.Arguments) {
			captured = args.Get(0).(*http.Request)
		}).
		Return(newHTTPResponse(http.StatusOK, okBody), nil).
		Once()
	client := NewClient("http://osrm.test", httpClient)
	ctx := context.WithValue(context.Background(), ctxKey{}, "present")

	_, err := client.GetRoute(ctx, srcCoord, dstCoord)

	require.NoError(t, err)
	require.NotNil(t, captured)
	assert.Equal(t, "present", captured.Context().Value(ctxKey{}))
	httpClient.AssertExpectations(t)
}

func TestGetRouteErrorsFromHTTPClient(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		wantErr error
	}{
		{name: "deadline exceeded", err: fmt.Errorf("do request: %w", context.DeadlineExceeded), wantErr: route.ErrUpstreamTimeout},
		{name: "os deadline exceeded", err: fmt.Errorf("do request: %w", os.ErrDeadlineExceeded), wantErr: route.ErrUpstreamTimeout},
		{name: "connection refused", err: errors.New("dial tcp: connection refused"), wantErr: route.ErrUpstream},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpClient := new(mockHTTPClient)
			httpClient.On("Do", mock.Anything).Return(nil, tt.err).Once()
			client := NewClient("http://osrm.test", httpClient)

			got, err := client.GetRoute(context.Background(), srcCoord, dstCoord)

			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, route.Route{}, got)
			httpClient.AssertExpectations(t)
		})
	}
}

func TestGetRouteBuildRequestError(t *testing.T) {
	httpClient := new(mockHTTPClient)
	client := NewClient("://invalid", httpClient)

	got, err := client.GetRoute(context.Background(), srcCoord, dstCoord)

	assert.ErrorIs(t, err, route.ErrUpstream)
	assert.Equal(t, route.Route{}, got)
	httpClient.AssertNotCalled(t, "Do", mock.Anything)
}
