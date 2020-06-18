package database

import "fmt"

type Config struct {
	Server   string `json:"server" toml:"server" yaml:"server"`
	User     string `json:"user" toml:"user" yaml:"user"`
	Password string `json:"password" toml:"password" yaml:"password"`
	Database string `json:"database" toml:"database" yaml:"database"`
	Extra    string `json:"extra" toml:"extra" yaml:"extra"`
	Debug    bool   `json:"debug" toml:"debug" yaml:"debug"`
}

func (cfg Config) String() string {
	return fmt.Sprintf("%s:%s@tcp(%s)/%s%s", cfg.User, cfg.Password, cfg.Server, cfg.Database, cfg.Extra)
}
