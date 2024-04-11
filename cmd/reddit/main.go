package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/jqdurham/reddit/internal/config"
	"github.com/jqdurham/reddit/internal/logger"
	"github.com/jqdurham/reddit/internal/orchestrator"
	"github.com/jqdurham/reddit/internal/reddit"
	"github.com/jqdurham/reddit/internal/service/post"
	"golang.org/x/time/rate"
)

const rateLimiterAllowableBurst = 1

func main() {
	lvl := new(slog.LevelVar)
	logr := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: lvl}))

	file, err := os.Open("./.env")
	if err != nil {
		logr.Error(err.Error())
		exit()
	}

	cfg, err := config.Configure(file)
	if err != nil {
		logr.Error(err.Error())
		exit()
	}

	lvl.Set(cfg.LogLevel)

	ctx := logger.NewContext(context.Background(), logr)

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	rateLimiter := rate.NewLimiter(rate.Every(cfg.RateLimit), rateLimiterAllowableBurst)
	client := reddit.NewClient(cfg.ClientID, cfg.ClientSecret, http.DefaultClient, rateLimiter)

	if err := client.Login(ctx, cfg.RedditUsername, cfg.RedditPassword); err != nil {
		logr.Error(err.Error())
		exit()
	}

	postSvc := post.NewService(client, os.Stdout)

	errCh := make(chan error)

	jobs := make([]orchestrator.Job, 0, len(cfg.Subreddits)*2)
	for _, subreddit := range cfg.Subreddits {
		jobs = append(jobs, func() error {
			return postSvc.UpdateTopPosts(ctx, subreddit)
		})
		jobs = append(jobs, func() error {
			return postSvc.UpdateTopNAuthors(ctx, subreddit, cfg.TopNAuthors)
		})
	}

	orchestrator.Run(ctx, errCh, jobs...)

	select {
	case err := <-errCh:
		logr.Error(err.Error())
		exit()
	case <-ctx.Done():
		logr.Info("Shutdown signal received, exiting...")
	}
}

func exit() {
	os.Exit(1)
}
