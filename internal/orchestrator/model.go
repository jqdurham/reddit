package orchestrator

// Job represents a unit of work to run continuously.
type Job func() error
