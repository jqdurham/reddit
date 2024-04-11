package test

// Errorer interface is used by tests to validate concurrent processes are running when expected.
//
//go:generate mockery --name Errorer
type Errorer interface {
	Err() error
	Err2() error
}
