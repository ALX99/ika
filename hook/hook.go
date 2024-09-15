package hook

import (
	"context"
)

type Factory interface {
	New(context.Context) (Hook, error)
}

type Hook interface {
	Setup(ctx context.Context, config map[string]any) error
	Teardown(ctx context.Context) error
}
