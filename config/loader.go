package config

import (
	"context"
)

type Loader interface {
	Watch(ctx context.Context, cfg Config) error
}
