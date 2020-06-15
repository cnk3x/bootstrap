package bootstrap

import (
	"go.shu.run/bootstrap/config/etcd"
	"go.shu.run/bootstrap/database"
	"go.shu.run/bootstrap/dig"
	"go.shu.run/bootstrap/logger/rus"
	"go.shu.run/bootstrap/mux"

	"go.etcd.io/etcd/v3/clientv3"
	"go.uber.org/fx"
)

func Run() {
	log := rus.Provide()
	log.Infof("启动...")

	client, err := configProvide([]string{"http://127.0.0.1:2379"}, "root", "root")
	if err != nil {
		log.Fatalf("%v", client)
	}

	dig.Supply(client)
	dig.Add(fx.Logger(log.Prefix("Fx")))
	dig.Provide(rus.Provide, etcd.New, database.New, mux.New)
	dig.Run()
}

func configProvide(endpoints []string, usr, pwd string) (*clientv3.Client, error) {
	return clientv3.New(clientv3.Config{Endpoints: endpoints, Username: usr, Password: pwd})
}
