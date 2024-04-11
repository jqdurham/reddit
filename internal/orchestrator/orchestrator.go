package orchestrator

import (
	"context"

	"github.com/jqdurham/reddit/internal/logger"
)

// Run executes provided Jobs perpetually, sending errors back to caller.
func Run(ctx context.Context, errCh chan error, jobs ...Job) {
	logr := logger.FromContext(ctx)

	for i := range jobs {
		go func(ctx context.Context, job Job) {
			for {
				select {
				case <-ctx.Done():
					errCh <- ctx.Err()
				default:
					if err := job(); err != nil {
						logr.Error(err.Error())
						errCh <- err
					}
				}
			}
		}(ctx, jobs[i])
	}
}
