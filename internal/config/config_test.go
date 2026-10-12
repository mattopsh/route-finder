package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func clearEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{"APP_PORT", "OSRM_BASE_URL", "OSRM_TIMEOUT_MS", "OSRM_MAX_CONCURRNCY"} {
		prev, ok := os.LookupEnv(key)
		require.NoError(t, os.Unsetenv(key))
		t.Cleanup(func() {
			if ok {
				require.NoError(t, os.Setenv(key, prev))
				return
			}
			require.NoError(t, os.Unsetenv(key))
		})
	}
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		want    Config
		wantErr string
	}{
		{
			name: "all set",
			env: map[string]string{
				"APP_PORT":            "9090",
				"OSRM_BASE_URL":       "http://osrm.test",
				"OSRM_TIMEOUT_MS":     "1000",
				"OSRM_MAX_CONCURRNCY": "5",
			},
			want: Config{
				ListenAddr:         9090,
				OSRMBaseURL:        "http://osrm.test",
				OSRMTimeoutMs:      1000,
				OSRMMaxConcurrency: 5,
			},
		},
		{
			name: "default listen addr",
			env: map[string]string{
				"OSRM_BASE_URL":       "http://osrm.test",
				"OSRM_TIMEOUT_MS":     "1000",
				"OSRM_MAX_CONCURRNCY": "5",
			},
			want: Config{
				ListenAddr:         8080,
				OSRMBaseURL:        "http://osrm.test",
				OSRMTimeoutMs:      1000,
				OSRMMaxConcurrency: 5,
			},
		},
		{
			name: "missing base url",
			env: map[string]string{
				"OSRM_TIMEOUT_MS":     "1000",
				"OSRM_MAX_CONCURRNCY": "5",
			},
			wantErr: "OSRM_BASE_URL",
		},
		{
			name: "missing timeout",
			env: map[string]string{
				"OSRM_BASE_URL":       "http://osrm.test",
				"OSRM_MAX_CONCURRNCY": "5",
			},
			wantErr: "OSRM_TIMEOUT_MS",
		},
		{
			name: "missing concurrency",
			env: map[string]string{
				"OSRM_BASE_URL":   "http://osrm.test",
				"OSRM_TIMEOUT_MS": "1000",
			},
			wantErr: "OSRM_MAX_CONCURRNCY",
		},
		{
			name: "zero timeout",
			env: map[string]string{
				"OSRM_BASE_URL":       "http://osrm.test",
				"OSRM_TIMEOUT_MS":     "0",
				"OSRM_MAX_CONCURRNCY": "5",
			},
			wantErr: "OSRM_TIMEOUT_MS: must be positive, got 0",
		},
		{
			name: "negative timeout",
			env: map[string]string{
				"OSRM_BASE_URL":       "http://osrm.test",
				"OSRM_TIMEOUT_MS":     "-1",
				"OSRM_MAX_CONCURRNCY": "5",
			},
			wantErr: "OSRM_TIMEOUT_MS: must be positive, got -1",
		},
		{
			name: "zero concurrency",
			env: map[string]string{
				"OSRM_BASE_URL":       "http://osrm.test",
				"OSRM_TIMEOUT_MS":     "1000",
				"OSRM_MAX_CONCURRNCY": "0",
			},
			wantErr: "OSRM_MAX_CONCURRNCY: must be positive, got 0",
		},
		{
			name: "negative concurrency",
			env: map[string]string{
				"OSRM_BASE_URL":       "http://osrm.test",
				"OSRM_TIMEOUT_MS":     "1000",
				"OSRM_MAX_CONCURRNCY": "-1",
			},
			wantErr: "OSRM_MAX_CONCURRNCY: must be positive, got -1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnv(t)
			for key, value := range tt.env {
				t.Setenv(key, value)
			}

			got, err := Load()

			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				assert.Equal(t, Config{}, got)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
