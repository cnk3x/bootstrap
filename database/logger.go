package database

import (
	"context"
	"time"

	"go.shu.run/bootstrap/logger"
	dl "gorm.io/gorm/logger"
)

var _ dl.Interface = (*l)(nil)

type l struct {
	logger.Logger
}

func (l *l) LogMode(level dl.LogLevel) dl.Interface {
	return l
}

func (l *l) Info(ctx context.Context, s string, i ...interface{}) {
	l.Infof(s, i...)
}

func (l *l) Warn(ctx context.Context, s string, i ...interface{}) {
	l.Infof(s, i...)
}

func (l *l) Error(ctx context.Context, s string, i ...interface{}) {
	l.Infof(s, i...)
}

func (l *l) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if err != nil {
		f, n := fc()
		l.Errorf("(%s:%d) %s -> %v", f, n, err, time.Now().Sub(begin).String())
	}
}
