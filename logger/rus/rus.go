package rus

import (
	"github.com/sirupsen/logrus"
	"go.shu.run/bootstrap/logger"
)

func Provide() logger.Logger {
	l := logrus.New()
	l.SetFormatter(&TextFormatter{})
	l.SetReportCaller(true)
	l.SetLevel(logrus.DebugLevel)
	l.Hooks.Add(newHook())
	return &Rus{Entry: logrus.NewEntry(l)}
}

type Rus struct {
	*logrus.Entry
}

func (r *Rus) Prefix(s string) logger.Logger {
	return &Rus{Entry: r.Entry.WithField("prefix", s)}
}
