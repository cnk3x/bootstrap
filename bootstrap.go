package bootstrap

import (
	"go.shu.run/bootstrap/database"
	"go.shu.run/bootstrap/dig"
	"go.shu.run/bootstrap/logger/rus"
	"go.shu.run/bootstrap/mux"

	"go.uber.org/fx"
)

//Run Run the bootstrap
func Run() {
	log := rus.New(rus.ConfigDefault)
	log.Infof("启动...")

	dig.Add(fx.Logger(log.Prefix("Fx")))
	dig.Provide(supplyConfig, rus.New, database.New, mux.New)
	dig.Run()
}
