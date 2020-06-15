package rus

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

var _ logrus.Hook = (*NotifyHook)(nil)

func newHook() logrus.Hook {
	name, _ := os.Executable()
	name = filepath.Base(name)
	ext := filepath.Ext(name)
	name = "logs/" + strings.TrimSuffix(name, ext) + ".log"
	return &NotifyHook{Logger: &lumberjack.Logger{
		Filename: name,
	}}
}

type NotifyHook struct {
	*lumberjack.Logger
}

func (m *NotifyHook) Levels() []logrus.Level {
	return []logrus.Level{logrus.ErrorLevel}
}

func (m *NotifyHook) Fire(entry *logrus.Entry) error {
	notify, find := entry.Data["notify"]
	if !find {
		return nil
	}
	if yes, ok := notify.(bool); !ok || !yes {
		return nil
	}

	go func(entry *logrus.Entry) {
		v, err := entry.Logger.Formatter.Format(entry)
		if err != nil {
			return
		}
		m.Logger.Write(v)
	}(entry)

	return nil
}
