package rus

import (
	"fmt"
	"io"
	"os"

	"go.shu.run/bootstrap/logger"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

func New(cfg Config) logger.Logger {
	l := logrus.New()
	l.SetFormatter(&TextFormatter{})
	l.SetReportCaller(cfg.Caller)
	if level, err := logrus.ParseLevel(cfg.Level); err == nil {
		l.SetLevel(level)
	} else {
		l.SetLevel(logrus.TraceLevel)
	}

	if cfg.Filename != "" && cfg.Filename != "std" {
		l.SetOutput(io.MultiWriter(&lumberjack.Logger{
			Filename:   cfg.Filename,
			MaxSize:    cfg.MaxSize,
			MaxAge:     cfg.MaxAge,
			MaxBackups: cfg.MaxBackups,
			LocalTime:  cfg.LocalTime,
			Compress:   cfg.Compress,
		}, os.Stdout))
	}

	return &Rus{entry: logrus.NewEntry(l)}
}

var _ logger.Logger = (*Rus)(nil)

type Rus struct {
	entry *logrus.Entry
}

func (r *Rus) Prefix(s string) logger.Logger {
	return &Rus{entry: r.entry.WithField("prefix", s)}
}

func (r *Rus) Logf(level string, format string, args ...interface{}) {
	l := r.levelMap(level)
	if r.entry.Logger.IsLevelEnabled(l) {
		r.entry.Log(l, fmt.Sprintf(format, args...))
	}
}

func (r *Rus) Debugf(format string, args ...interface{}) {
	r.entry.Logf(logrus.DebugLevel, format, args...)
}

func (r *Rus) Infof(format string, args ...interface{}) {
	//r.entry.Logf(logrus.InfoLevel, format, args...)
	r.entry.Infof(format, args...)
}

func (r *Rus) Errorf(format string, args ...interface{}) {
	r.entry.Logf(logrus.ErrorLevel, format, args...)
}

func (r *Rus) Fatalf(format string, args ...interface{}) {
	r.entry.Logf(logrus.FatalLevel, format, args...)
	os.Exit(1)
}

func (r *Rus) Printf(format string, args ...interface{}) {
	r.entry.Logf(logrus.InfoLevel, format, args...)
}

func (r *Rus) levelMap(l string) logrus.Level {
	switch l[0] {
	case 't', 'T':
		return logrus.TraceLevel
	case 'd', 'D':
		return logrus.DebugLevel
	case 'i', 'I':
		return logrus.InfoLevel
	case 'w', 'W':
		return logrus.WarnLevel
	case 'e', 'E':
		return logrus.ErrorLevel
	case 'f', 'F':
		return logrus.FatalLevel
	case 'p', 'P':
		return logrus.PanicLevel
	default:
		return logrus.TraceLevel
	}
}
