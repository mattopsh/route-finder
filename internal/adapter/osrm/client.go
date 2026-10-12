package osrm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"route-finder/internal/geo"
	"route-finder/internal/route"
)

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	baseUrl    string
	httpClient httpClient
}

func NewClient(baseUrl string, httpClient httpClient) *Client {
	return &Client{baseUrl: baseUrl, httpClient: httpClient}
}

func (c *Client) GetRoute(ctx context.Context, from, to geo.Coordinate) (route.Route, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.buildUrl(from, to), nil)
	if err != nil {
		return route.Route{}, fmt.Errorf("%w: build request: %v", route.ErrUpstream, err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, os.ErrDeadlineExceeded) {
			return route.Route{}, fmt.Errorf("%w: %v", route.ErrUpstreamTimeout, err)
		}
		return route.Route{}, fmt.Errorf("%w: %v", route.ErrUpstream, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadRequest {
		return route.Route{}, fmt.Errorf("%w: unexpected http status %d", route.ErrUpstream, resp.StatusCode)
	}

	var body routeResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return route.Route{}, fmt.Errorf("%w: decode response (status %d): %v", route.ErrUpstream, resp.StatusCode, err)
	}

	return c.buildResponse(to, body)
}

func (c *Client) buildUrl(from, to geo.Coordinate) string {
	return strings.TrimSuffix(c.baseUrl, "/") + "/route/v1/driving/" + from.String() + ";" + to.String() + "?overview=false"
}

func (c *Client) buildResponse(to geo.Coordinate, body routeResponse) (route.Route, error) {
	switch body.Code {
	case codeOK:
		if len(body.Routes) == 0 {
			return route.Route{}, fmt.Errorf("%w: code=Ok but no routes returned", route.ErrUpstream)
		}
		return toRoute(to, body.Routes[0]), nil
	case codeNoRoute:
		return route.Route{}, fmt.Errorf("%w: %s", route.ErrNoRoute, body.describe())
	case codeNoSegment, codeInvalidValue, codeInvalidUrl, codeInvalidService, codeInvalidVersion,
		codeInvalidOptions, codeInvalidQuery, codeTooBig, codeDisabledDataset:
		return route.Route{}, fmt.Errorf("%w: %s", route.ErrUpstream, body.describe())
	default:
		return route.Route{}, fmt.Errorf("%w: %s", route.ErrUpstream, body.describe())
	}
}
