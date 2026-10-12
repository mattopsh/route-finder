package geo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseCoordinate(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Coordinate
		wantErr string
	}{
		{
			name:  "valid",
			input: "13.388860,52.517037",
			want:  Coordinate{Lng: "13.388860", Lat: "52.517037"},
		},
		{
			name:  "integers",
			input: "13,52",
			want:  Coordinate{Lng: "13", Lat: "52"},
		},
		{
			name:  "negative values",
			input: "-73.985130,40.758896",
			want:  Coordinate{Lng: "-73.985130", Lat: "40.758896"},
		},
		{
			name:  "boundary values",
			input: "180,90",
			want:  Coordinate{Lng: "180", Lat: "90"},
		},
		{
			name:  "negative boundary values",
			input: "-180,-90",
			want:  Coordinate{Lng: "-180", Lat: "-90"},
		},
		{
			name:    "empty",
			input:   "",
			wantErr: `invalid coordinate "": want lng,lat`,
		},
		{
			name:    "missing comma",
			input:   "13.388860",
			wantErr: `invalid coordinate "13.388860": want lng,lat`,
		},
		{
			name:    "non-numeric longitude",
			input:   "abc,52.5",
			wantErr: `invalid longitude "abc"`,
		},
		{
			name:    "non-numeric latitude",
			input:   "13.4,xyz",
			wantErr: `invalid latitude "xyz"`,
		},
		{
			name:    "latitude with comma",
			input:   "13.4,52.5,1",
			wantErr: `invalid latitude "52.5,1"`,
		},
		{
			name:    "longitude out of range",
			input:   "181,52.5",
			wantErr: `longitude "181" out of range [-180,180]`,
		},
		{
			name:    "negative longitude out of range",
			input:   "-181,52.5",
			wantErr: `longitude "-181" out of range [-180,180]`,
		},
		{
			name:    "latitude out of range",
			input:   "13.4,91",
			wantErr: `latitude "91" out of range [-90,90]`,
		},
		{
			name:    "negative latitude out of range",
			input:   "13.4,-91",
			wantErr: `latitude "-91" out of range [-90,90]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCoordinate(tt.input)

			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				assert.Equal(t, Coordinate{}, got)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCoordinateString(t *testing.T) {
	c := Coordinate{Lng: "13.388860", Lat: "52.517037"}
	assert.Equal(t, "13.388860,52.517037", c.String())
}
