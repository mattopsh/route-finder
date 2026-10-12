package route

import (
	"cmp"
	"context"
	"fmt"
	"slices"

	"golang.org/x/sync/errgroup"

	"route-finder/internal/geo"
)

type routeFinder interface {
	GetRoute(ctx context.Context, from, to geo.Coordinate) (Route, error)
}

type Service struct {
	routeFinder               routeFinder
	routeFinderMaxConcurrency int
}

func NewService(routeFinder routeFinder, routeFinderMaxConcurrency int) *Service {
	return &Service{routeFinder: routeFinder, routeFinderMaxConcurrency: routeFinderMaxConcurrency}
}

func (s *Service) FindRoutes(ctx context.Context, src geo.Coordinate, dsts []geo.Coordinate) ([]Route, error) {
	routes, err := s.fetchRoutes(ctx, src, dsts)
	if err != nil {
		return nil, err
	}

	sortRoutes(routes)

	return routes, nil
}

func (s *Service) fetchRoutes(ctx context.Context, src geo.Coordinate, dsts []geo.Coordinate) ([]Route, error) {
	routes := make([]Route, len(dsts))

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(s.routeFinderMaxConcurrency)

	for i, dst := range dsts {
		g.Go(func() error {
			route, err := s.routeFinder.GetRoute(ctx, src, dst)
			if err != nil {
				return fmt.Errorf("route to %s: %w", dst, err)
			}

			routes[i] = route
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return routes, nil
}

func sortRoutes(routes []Route) {
	slices.SortFunc(routes, func(a, b Route) int {
		if c := cmp.Compare(a.Duration, b.Duration); c != 0 {
			return c
		}
		return cmp.Compare(a.Distance, b.Distance)
	})
}
