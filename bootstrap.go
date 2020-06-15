package bootstrap

import (
	"go.etcd.io/etcd/v3/clientv3"
	"go.shu.run/bootstrap/config/etcd"
	"go.shu.run/bootstrap/database"
	"go.shu.run/bootstrap/dig"
	"go.shu.run/bootstrap/logger/rus"
	"go.shu.run/bootstrap/mux"
	"go.uber.org/fx"
)

func Run() {
	log := rus.Provide()
	log.Infof("启动...")
	dig.Add(fx.Logger(log.Prefix("Fx")))
	dig.Provide(rus.Provide, configProvide, etcd.New, database.New, mux.New)
	dig.Run()
}

func configProvide() (*clientv3.Client, error) {
	return clientv3.New(clientv3.Config{Endpoints: []string{"http://127.0.0.1:2379"}, Username: "root", Password: "root"})
}
