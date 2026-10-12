package route

import "errors"

var (
	ErrNoRoute         = errors.New("no route found")
	ErrUpstream        = errors.New("upstream error")
	ErrUpstreamTimeout = errors.New("upstream timeout")
)
