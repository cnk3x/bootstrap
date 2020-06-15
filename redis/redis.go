package redis

import (
	"context"

	"github.com/go-redis/redis/v8"
	"go.shu.run/bootstrap/logger"
	"go.uber.org/fx"
)

const (
	sqlRedisKey = "redis.dsn"
)

//NewRedis 新建路由提供程序
func NewRedis(l logger.Logger) *redis.Client {
	log := l.Prefix("RDS")
	log.Infof("Redis: 注册")

	var (
		//dsn  = "" //url.Parse(app.GetString(sqlRedisKey))
		addr = "" //dsn.Host
		pwd  = "" //dsn.User.Username()
		dbi  = 0  //strconv.Atoi(strings.Trim(dsn.Path, "/"))
	)

	return redis.NewClient(&redis.Options{Addr: addr, Password: pwd, DB: dbi})
}

//StartRedis 启动redis
func StartRedis(lc fx.Lifecycle, l logger.Logger, ring *redis.Client) {
	log := l.Prefix("RDS")
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ret, err := ring.Ping(ctx).Result()
			if err != nil {
				log.Errorf("Redis: 打开 (出错): %v", err)
				return err
			}
			log.Infof("Redis: 已连接: %s", ret)
			return nil
		},
		OnStop: func(_ context.Context) error {
			defer log.Infof("Redis: 已关闭")
			log.Infof("Redis: 关闭")
			if err := ring.Close(); err != nil {
				log.Errorf("Redis: 关闭, 出错: %v", err)
			}
			return nil
		},
	})
}
