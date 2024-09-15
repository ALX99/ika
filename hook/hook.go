package hook

import (
	"context"
)

type Hooker interface {
	Setup(ctx context.Context, config map[string]any) error
	Teardown(ctx context.Context) error
}
