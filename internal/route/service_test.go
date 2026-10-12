package route

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"route-finder/internal/geo"
)

type mockRouteFinder struct {
	mock.Mock
}

func (m *mockRouteFinder) GetRoute(ctx context.Context, from, to geo.Coordinate) (Route, error) {
	args := m.Called(ctx, from, to)
	return args.Get(0).(Route), args.Error(1)
}

func coord(lng, lat string) geo.Coordinate {
	return geo.Coordinate{Lng: lng, Lat: lat}
}

func TestFindRoutesSortsByDuration(t *testing.T) {
	// GIVEN
	src := coord("13.4", "52.5")
	dstA := coord("13.5", "52.6")
	dstB := coord("13.6", "52.7")
	dstC := coord("13.7", "52.8")

	finder := new(mockRouteFinder)
	finder.On("GetRoute", mock.Anything, src, dstA).Return(Route{To: dstA, Duration: 30, Distance: 100}, nil)
	finder.On("GetRoute", mock.Anything, src, dstB).Return(Route{To: dstB, Duration: 10, Distance: 300}, nil)
	finder.On("GetRoute", mock.Anything, src, dstC).Return(Route{To: dstC, Duration: 20, Distance: 200}, nil)
	svc := NewService(finder, 3)

	// WHEN
	got, err := svc.FindRoutes(context.Background(), src, []geo.Coordinate{dstA, dstB, dstC})

	// THEN
	require.NoError(t, err)
	require.Len(t, got, 3)
	assert.Equal(t, dstB, got[0].To)
	assert.Equal(t, dstC, got[1].To)
	assert.Equal(t, dstA, got[2].To)
	finder.AssertExpectations(t)
	finder.AssertNumberOfCalls(t, "GetRoute", 3)
}

func TestFindRoutesSortsByDistanceWhenDurationTies(t *testing.T) {
	// GIVEN
	src := coord("13.4", "52.5")
	dstA := coord("13.5", "52.6")
	dstB := coord("13.6", "52.7")
	dstC := coord("13.7", "52.8")

	finder := new(mockRouteFinder)
	finder.On("GetRoute", mock.Anything, src, dstA).Return(Route{To: dstA, Duration: 10, Distance: 5}, nil)
	finder.On("GetRoute", mock.Anything, src, dstB).Return(Route{To: dstB, Duration: 10, Distance: 1}, nil)
	finder.On("GetRoute", mock.Anything, src, dstC).Return(Route{To: dstC, Duration: 10, Distance: 3}, nil)
	svc := NewService(finder, 3)

	// WHEN
	got, err := svc.FindRoutes(context.Background(), src, []geo.Coordinate{dstA, dstB, dstC})

	// THEN
	require.NoError(t, err)
	require.Len(t, got, 3)
	assert.Equal(t, dstB, got[0].To)
	assert.Equal(t, dstC, got[1].To)
	assert.Equal(t, dstA, got[2].To)
	finder.AssertExpectations(t)
}

func TestFindRoutesEmptyDsts(t *testing.T) {
	// GIVEN
	finder := new(mockRouteFinder)
	svc := NewService(finder, 3)

	// WHEN
	got, err := svc.FindRoutes(context.Background(), coord("13.4", "52.5"), nil)

	// THEN
	require.NoError(t, err)
	assert.Empty(t, got)
	finder.AssertNotCalled(t, "GetRoute", mock.Anything, mock.Anything, mock.Anything)
}

func TestFindRoutesWrapsFinderError(t *testing.T) {
	// GIVEN
	src := coord("13.4", "52.5")
	dst := coord("13.5", "52.6")

	finder := new(mockRouteFinder)
	finder.On("GetRoute", mock.Anything, src, dst).Return(Route{}, ErrNoRoute).Once()
	svc := NewService(finder, 3)

	// WHEN
	got, err := svc.FindRoutes(context.Background(), src, []geo.Coordinate{dst})

	// THEN
	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorIs(t, err, ErrNoRoute)
	assert.ErrorContains(t, err, "route to 13.5,52.6")
	finder.AssertExpectations(t)
}

func TestFindRoutesAllOrNothingOnFailure(t *testing.T) {
	// GIVEN
	src := coord("13.4", "52.5")
	dstOK := coord("13.5", "52.6")
	dstFail := coord("13.6", "52.7")

	finder := new(mockRouteFinder)
	finder.On("GetRoute", mock.Anything, src, dstOK).Return(Route{To: dstOK, Duration: 10, Distance: 20}, nil)
	finder.On("GetRoute", mock.Anything, src, dstFail).Return(Route{}, ErrUpstream)
	svc := NewService(finder, 3)

	// WHEN
	got, err := svc.FindRoutes(context.Background(), src, []geo.Coordinate{dstOK, dstFail})

	// THEN
	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorIs(t, err, ErrUpstream)
	finder.AssertExpectations(t)
}
