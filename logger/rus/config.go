package rus

import "github.com/BurntSushi/toml"

type Config struct {
	Caller bool   `json:"caller" toml:"caller" yaml:"caller"`
	Level  string `json:"level" toml:"level" yaml:"level"`

	// Filename is the file to write logs to.  Backup log files will be retained
	// in the same directory.  It uses <processname>-lumberjack.log in
	// os.TempDir() if empty.
	Filename string `json:"filename" toml:"filename" yaml:"filename"`

	// MaxSize is the maximum size in megabytes of the log file before it gets
	// rotated. It defaults to 100 megabytes.
	MaxSize int `json:"maxsize" toml:"maxsize" yaml:"maxsize"`

	// MaxAge is the maximum number of days to retain old log files based on the
	// timestamp encoded in their filename.  Note that a day is defined as 24
	// hours and may not exactly correspond to calendar days due to daylight
	// savings, leap seconds, etc. The default is not to remove old log files
	// based on age.
	MaxAge int `json:"maxage" toml:"maxage" yaml:"maxage"`

	// MaxBackups is the maximum number of old log files to retain.  The default
	// is to retain all old log files (though MaxAge may still cause them to get
	// deleted.)
	MaxBackups int `json:"maxbackups" toml:"maxbackups" yaml:"maxbackups"`

	// LocalTime determines if the time used for formatting the timestamps in
	// backup files is the computer's local time.  The default is to use UTC
	// time.
	LocalTime bool `json:"localtime" toml:"localtime" yaml:"localtime"`

	// Compress determines if the rotated log files should be compressed
	// using gzip. The default is not to perform compression.
	Compress bool `json:"compress" toml:"compress" yaml:"compress"`
}

var ConfigDefault = Config{
	Caller:     true,
	Level:      "trace",
	Filename:   "std",
	MaxSize:    100,
	MaxAge:     0,
	MaxBackups: 300,
	LocalTime:  true,
	Compress:   false,
}

func (cfg Config) Update(v []byte) error {
	return toml.Unmarshal(v, &cfg)
}
