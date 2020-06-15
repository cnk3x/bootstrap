package main

import (
	"context"
	"go.uber.org/fx"

	"go.shu.run/bootstrap"
	"go.shu.run/bootstrap/database"
	"go.shu.run/bootstrap/dig"
	"go.shu.run/bootstrap/logger"
	"gorm.io/gorm"
)

func main() {
	dig.Invoke(func(log logger.Logger, db *database.DB, fc fx.Lifecycle) {
		log.Prefix("DB").Errorf("打开数据库")
		fc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				go db.GetDB(ctx).AutoMigrate(&Hello{})
				return nil
			},
		})

	})
	bootstrap.Run()
}

type Hello struct {
	gorm.Model
}
