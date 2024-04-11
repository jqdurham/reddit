package config_test

import (
	"bytes"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/jqdurham/reddit/internal/config"
	"github.com/stretchr/testify/require"
)

func TestConfigure(t *testing.T) {
	t.Parallel()

	requiredEnvs := `
REDDIT_CLIENT_ID=test-client-id
REDDIT_CLIENT_SECRET=test-client-secret
REDDIT_USERNAME=test-username
REDDIT_PASSWORD=test-password`

	tests := []struct {
		name    string
		envVars io.Reader
		want    *config.Config
		errMsg  string
	}{
		{
			name:    "Required parameters only",
			envVars: strings.NewReader(requiredEnvs),
			want: &config.Config{
				ClientID:       "test-client-id",
				ClientSecret:   "test-client-secret",
				RedditUsername: "test-username",
				RedditPassword: "test-password",
				Subreddits:     []string{"golang"},
				RateLimit:      time.Second,
				LogLevel:       slog.LevelInfo,
				TopNAuthors:    10,
			},
		},
		{
			name:    "Missing REDDIT_CLIENT_ID",
			envVars: &bytes.Buffer{},
			errMsg:  "missing env: REDDIT_CLIENT_ID",
		},
		{
			name:    "Missing REDDIT_CLIENT_SECRET",
			envVars: strings.NewReader(`REDDIT_CLIENT_ID=test-client-id`),
			errMsg:  "missing env: REDDIT_CLIENT_SECRET",
		},
		{
			name: "Missing REDDIT_USERNAME",
			envVars: strings.NewReader("REDDIT_CLIENT_ID=test-client-id" +
				"\nREDDIT_CLIENT_SECRET=test-client-secret"),
			errMsg: "missing env: REDDIT_USERNAME",
		},
		{
			name: "Missing REDDIT_PASSWORD",
			envVars: strings.NewReader("REDDIT_CLIENT_ID=test-client-id" +
				"\nREDDIT_CLIENT_SECRET=test-client-secret" +
				"\nREDDIT_USERNAME=test-username"),
			errMsg: "missing env: REDDIT_PASSWORD",
		},
		{
			name:    "Invalid time.Duration for REDDIT_RATE_LIMIT",
			envVars: strings.NewReader(requiredEnvs + "\nREDDIT_RATE_LIMIT=invalid"),
			errMsg:  `invalid env: REDDIT_RATE_LIMIT reason: time: invalid duration "invalid"`,
		},
		{
			name:    "Invalid integer for REDDIT_TOP_N_AUTHORS",
			envVars: strings.NewReader(requiredEnvs + "\nREDDIT_TOP_N_AUTHORS=NaN"),
			errMsg:  `invalid env: REDDIT_TOP_N_AUTHORS reason: strconv.Atoi: parsing "NaN": invalid syntax`,
		},
		{
			name:    "Invalid REDDIT_LOG_LEVEL",
			envVars: strings.NewReader(requiredEnvs + "\nREDDIT_LOG_LEVEL=NaL"),
			errMsg:  `invalid env: REDDIT_LOG_LEVEL reason: invalid env: log level reason: must be: debug, info, warn, error`,
		},
		{
			name: "All parameters",
			envVars: strings.NewReader(requiredEnvs +
				"\nREDDIT_SUBREDDITS=subreddit1,subreddit2" +
				"\nREDDIT_RATE_LIMIT=60s" +
				"\nREDDIT_LOG_LEVEL=debug" +
				"\nREDDIT_TOP_N_AUTHORS=1337"),
			want: &config.Config{
				ClientID:       "test-client-id",
				ClientSecret:   "test-client-secret",
				RedditUsername: "test-username",
				RedditPassword: "test-password",
				Subreddits:     []string{"subreddit1", "subreddit2"},
				RateLimit:      time.Minute,
				LogLevel:       slog.LevelDebug,
				TopNAuthors:    1337,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := config.Configure(tt.envVars)
			if tt.errMsg != "" {
				require.EqualError(t, err, tt.errMsg)
				require.Nil(t, got)

				return
			}

			require.NoError(t, err)
			require.EqualValues(t, tt.want, got)
		})
	}
}
