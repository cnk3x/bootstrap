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

func (l *l) LogMode(_ dl.LogLevel) dl.Interface {
	return l
}

func (l *l) Info(_ context.Context, s string, i ...interface{}) {
	l.Logf("info", s, i...)
}

func (l *l) Warn(_ context.Context, s string, i ...interface{}) {
	l.Logf("warn", s, i...)
}

func (l *l) Error(_ context.Context, s string, i ...interface{}) {
	l.Logf("error", s, i...)
}

func (l *l) Trace(_ context.Context, begin time.Time, fc func() (string, int64), err error) {
	if err != nil {
		f, n := fc()
		l.Logf("trace", "(%s:%d) %s -> %v", f, n, err, time.Now().Sub(begin).String())
	}
}
