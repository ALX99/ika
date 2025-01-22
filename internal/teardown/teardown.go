package teardown

import (
	"context"
	"errors"
	"time"
)

type (
	TeardownFunc func(context.Context) error
	Teardowner   []TeardownFunc
)

// Add returns a new teardown function that runs both teardown functions.
func (t Teardowner) Add(teardown ...TeardownFunc) Teardowner {
	for _, td := range teardown {
		t = append(t, td)
	}
	return t
}

// Teardown runs all teardown functions.
func (t Teardowner) Teardown(ctx context.Context) error {
	var err error
	for _, teardown := range t {
		tearDownCtx, cancel := context.WithTimeoutCause(ctx, 5*time.Second, errors.New("teardown timeout exceeded"))
		err = errors.Join(err, teardown(tearDownCtx))
		cancel()
	}
	return err
}
