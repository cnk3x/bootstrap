package bootstrap

import (
	"go.shu.run/bootstrap/database"
	"go.shu.run/bootstrap/dig"
	"go.shu.run/bootstrap/logger/rus"
	"go.shu.run/bootstrap/mux"

	"go.uber.org/fx"
)

func Run() {
	log := rus.New(rus.ConfigDefault)
	log.Infof("启动...")

	//if err := supplyConfig(); err != nil {
	//	log.Fatalf("%v", err)
	//}

	dig.Add(fx.Logger(log.Prefix("Fx")))
	dig.Provide(supplyConfig, rus.New)
	dig.Provide(database.New, mux.New)
	dig.Run()
}
