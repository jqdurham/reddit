package config

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	ClientID, ClientSecret,
	RedditUsername, RedditPassword string
	Subreddits  []string
	RateLimit   time.Duration
	LogLevel    slog.Level
	TopNAuthors int
}

func Configure(envVars io.Reader) (*Config, error) {
	var (
		clientID, secret,
		username, password, topNAuthors,
		subreddits, rateLimit, logLevel string
		vars map[string]string
		err  error
	)

	if vars, err = godotenv.Parse(envVars); err != nil {
		return nil, fmt.Errorf("parse configuration: %w", err)
	}

	if clientID, err = getRequiredEnv(vars, "REDDIT_CLIENT_ID"); err != nil {
		return nil, err
	}

	if secret, err = getRequiredEnv(vars, "REDDIT_CLIENT_SECRET"); err != nil {
		return nil, err
	}

	if username, err = getRequiredEnv(vars, "REDDIT_USERNAME"); err != nil {
		return nil, err
	}

	if password, err = getRequiredEnv(vars, "REDDIT_PASSWORD"); err != nil {
		return nil, err
	}

	subreddits = getOptionalEnv(vars, "REDDIT_SUBREDDITS", "golang")
	logLevel = getOptionalEnv(vars, "REDDIT_LOG_LEVEL", "info")

	rateLimit = getOptionalEnv(vars, "REDDIT_RATE_LIMIT", "1s")
	freq, err := time.ParseDuration(rateLimit)
	if err != nil {
		return nil, NewInvalidConfigInputError("REDDIT_RATE_LIMIT", err.Error())
	}

	topNAuthors = getOptionalEnv(vars, "REDDIT_TOP_N_AUTHORS", "10")
	num, err := strconv.Atoi(topNAuthors)
	if err != nil {
		return nil, NewInvalidConfigInputError("REDDIT_TOP_N_AUTHORS", err.Error())
	}

	level, err := toLevel(logLevel)
	if err != nil {
		return nil, NewInvalidConfigInputError("REDDIT_LOG_LEVEL", err.Error())
	}

	return &Config{
		ClientID:       clientID,
		ClientSecret:   secret,
		RedditUsername: username,
		RedditPassword: password,
		Subreddits:     strings.Split(subreddits, ","),
		RateLimit:      freq,
		LogLevel:       level,
		TopNAuthors:    num,
	}, nil
}

func getRequiredEnv(vars map[string]string, env string) (string, error) {
	if v, ok := os.LookupEnv(env); ok {
		return v, nil
	}
	v, ok := vars[env]
	if !ok {
		return "", NewMissingConfigInputError(env)
	}

	return v, nil
}

func getOptionalEnv(vars map[string]string, env, def string) string {
	if v, ok := os.LookupEnv(env); ok {
		return v
	}
	if v, ok := vars[env]; ok {
		return v
	}

	return def
}

func toLevel(level string) (slog.Level, error) {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	}

	return slog.LevelInfo, NewInvalidConfigInputError("log level", "must be: debug, info, warn, error")
}
