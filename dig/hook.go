package dig

import (
	"context"
	"go.uber.org/fx"
)

type Service interface {
	OnStart(context.Context) error
	OnStop(context.Context) error
}

func StartService(services ...Service) {
	for _, service := range services {
		Populate(service)
		Invoke(func(fc fx.Lifecycle) {
			fc.Append(fx.Hook{
				OnStart: service.OnStart,
				OnStop:  service.OnStop,
			})
		})
	}
}
