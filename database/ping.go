package database

import (
	"context"
	"gorm.io/gorm"
)

type iPingC interface {
	PingContext(ctx context.Context) error
}

type iPing interface {
	Ping() error
}

func ping(ctx context.Context, gdb *gorm.DB) error {
	if ping, ok := gdb.ConnPool.(iPingC); ok {
		return ping.PingContext(ctx)
	}

	if ping, ok := gdb.ConnPool.(iPing); ok {
		return ping.Ping()
	}

	return ErrNotOpen
}
