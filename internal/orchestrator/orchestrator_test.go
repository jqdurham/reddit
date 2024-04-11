package orchestrator_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jqdurham/reddit/internal/orchestrator"
	testmocks "github.com/jqdurham/reddit/test/mocks"
	"github.com/stretchr/testify/assert"
)

var errMockedFailure = errors.New("mocked failure")

func TestRun(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		jobs   func(t *testing.T, m *testmocks.Errorer) []orchestrator.Job
		errMsg string
	}{
		{
			name: "Error on 5th iteration of single job",
			jobs: func(t *testing.T, m *testmocks.Errorer) []orchestrator.Job {
				t.Helper()
				m.On("Err").Return(nil).Times(4)
				m.On("Err").Return(errMockedFailure).Once()
				// Runner doesn't normally exit on an error,
				// and there is some delay whilst the error is sent back through errCh,
				// the context is cancelled, and that cancellation propagates.
				m.On("Err").Return(nil)
				job := func() error { return m.Err() }

				return []orchestrator.Job{job}
			},
			errMsg: errMockedFailure.Error(),
		},
		{
			name: "Error on 4th iteration of the 2nd job",
			jobs: func(t *testing.T, m *testmocks.Errorer) []orchestrator.Job {
				t.Helper()
				m.On("Err").Return(nil)
				m.On("Err2").Return(nil).Times(3)
				m.On("Err2").Return(errMockedFailure).Once()
				// Runner doesn't normally exit on an error,
				// and there is some delay whilst the error is sent back through errCh,
				// the context is cancelled, and that cancellation propagates.
				m.On("Err2").Return(nil)

				job1 := func() error { return m.Err() }
				job2 := func() error { return m.Err2() }

				return []orchestrator.Job{job1, job2}
			},
			errMsg: errMockedFailure.Error(),
		},
		{
			name: "Context cancellation exits runner",
			jobs: func(t *testing.T, m *testmocks.Errorer) []orchestrator.Job {
				t.Helper()
				m.On("Err").Return(nil)
				job := func() error { return m.Err() }

				return []orchestrator.Job{job}
			},
			errMsg: "context deadline exceeded",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			errCh := make(chan error)
			stub := testmocks.NewErrorer(t)
			jobs := tt.jobs(t, stub)

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			go orchestrator.Run(ctx, errCh, jobs...)

			if err := <-errCh; true {
				cancel()
				assert.EqualError(t, err, tt.errMsg)

				return
			}
		})
	}
}
