package rus

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

var loc = time.FixedZone("Asia/Shanghai", 8*3600)

type TextFormatter struct {
}

func (f *TextFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	f.print(b, entry)

	b.WriteByte('\n')
	return b.Bytes(), nil
}

func (f *TextFormatter) print(b *bytes.Buffer, entry *logrus.Entry) {
	entry.Message = strings.TrimSuffix(entry.Message, "\n")

	caller := ""
	if entry.HasCaller() {
		caller = fmt.Sprintf(" %-20s", fmt.Sprintf("(%s:%d)", filepath.Base(entry.Caller.File), entry.Caller.Line))
	}

	var prefix string
	if prefixI, ok := entry.Data["prefix"]; ok {
		if prefix, _ = prefixI.(string); prefix != "" {
			prefix = strings.ToUpper(prefix + ".")
		}
	}

	fmt.Fprintf(b, "%s%s <%s>%s ⇨ %-44s ",
		prefix,
		strings.ToUpper(entry.Level.String())[0:4],
		entry.Time.In(loc).Format("2006/01/02 15:04:05"),
		caller,
		entry.Message,
	)

	for k, v := range entry.Data {
		if k != "prefix" {
			fmt.Fprintf(b, " %s=", k)
			f.appendValue(b, v)
		}
	}
}

func (f *TextFormatter) appendValue(b *bytes.Buffer, value interface{}) {
	stringVal, ok := value.(string)
	if !ok {
		stringVal = fmt.Sprint(value)
	}

	if !f.needsQuoting(stringVal) {
		b.WriteString(stringVal)
	} else {
		b.WriteString(fmt.Sprintf("%q", stringVal))
	}
}

func (f *TextFormatter) needsQuoting(text string) bool {
	for _, ch := range text {
		if !((ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '-' || ch == '.' || ch == '_' || ch == '/' || ch == '@' || ch == '^' || ch == '+') {
			return true
		}
	}
	return false
}
