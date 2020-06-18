package redis

import (
	"go.shu.run/bootstrap/logger"

	"github.com/go-redis/redis/v8"
)

type Config struct {
	Addr     string `json:"addr" toml:"addr" yaml:"addr"`
	Password string `json:"password" toml:"password" yaml:"password"`
	DB       int    `json:"db" toml:"db" yaml:"db"`
}

//NewRedis 新建路由提供程序
func NewRedis(l logger.Logger, cfg Config) *redis.Client {
	log := l.Prefix("RDS")
	log.Infof("Redis: 注册")

	return redis.NewClient(&redis.Options{Addr: cfg.Addr, Password: cfg.Password, DB: cfg.DB})
}
