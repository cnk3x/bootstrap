package bootstrap

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"go.shu.run/bootstrap/database"
	"go.shu.run/bootstrap/logger/rus"
	"go.shu.run/bootstrap/redis"
	"go.shu.run/bootstrap/utils/json"

	"github.com/BurntSushi/toml"
	"go.uber.org/fx"
	"gopkg.in/yaml.v2"
)

//Config Config
type Config struct {
	fx.Out
	Log      rus.Config      `json:"log" toml:"log" yaml:"log"`
	Redis    redis.Config    `json:"redis" toml:"redis" yaml:"redis"`
	Database database.Config `json:"sql" toml:"sql" yaml:"sql"`
	Mux      MuxConfig       `json:"mux" toml:"mux" yaml:"mux"`

	Type string `json:"-" toml:"-" yaml:"-"`
}

//Update Update
func (cfg *Config) Update(v []byte) error {
	switch cfg.Type {
	case "yaml", "yml":
		return yaml.Unmarshal(v, &cfg)
	case "toml":
		return toml.Unmarshal(v, &cfg)
	default:
		return json.Unmarshal(v, &cfg)
	}
}

func supplyConfig() (cfg Config, err error) {
	_ = os.MkdirAll("conf", 0755)
	ext := []string{"json5", "json", "toml", "yml"}
	for i := range ext {
		fn := filepath.Join("conf", "bootstrap."+ext[i])
		v, err := ioutil.ReadFile(fn)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return cfg, err
		}

		cfg.Type = ext[i]
		if err := cfg.Update(v); err != nil {
			return cfg, err
		}
		//dig.Populate(cfg)
		//dig.Supply(cfg.Log, cfg.Redis, cfg.Sql)
		return cfg, nil
	}

	return cfg, errors.New("no conf file")
}
