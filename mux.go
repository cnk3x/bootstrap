package bootstrap

import (
	"context"
	"net"
	"net/http"

	"go.shu.run/bootstrap/dig"
	"go.shu.run/bootstrap/logger"
	"go.shu.run/bootstrap/mux"
	"go.uber.org/fx"
)

//MuxConfig .
type MuxConfig struct {
	ListenAt string `json:"listen_at"`
}

//AddHTTPServer StartMux
func AddHTTPServer() {
	dig.Invoke(startHTTPServer)
}

func startHTTPServer(mux *mux.Mux, log logger.Logger, cfg MuxConfig, fc fx.Lifecycle) {
	ms := &muxServer{
		log:     log.Prefix("Mux"),
		cfg:     cfg,
		handler: mux,
	}
	fc.Append(fx.Hook{
		OnStart: ms.OnStart,
		OnStop:  ms.OnStop,
	})
}

//muxServer muxServer
type muxServer struct {
	log     logger.Logger
	cfg     MuxConfig
	handler *mux.Mux
	server  *http.Server
}

//OnStart OnStart
func (m *muxServer) OnStart(ctx context.Context) error {
	m.server = &http.Server{
		Addr:    m.cfg.ListenAt,
		Handler: m.handler,
		BaseContext: func(ln net.Listener) context.Context {
			return ctx
		},
	}
	m.handler.SetLogger(m.log)
	m.log.Infof("http server starting...")
	go m.server.ListenAndServe()
	m.log.Infof("listen at: %s", m.cfg.ListenAt)
	return nil
}

//OnStop OnStop
func (m *muxServer) OnStop(ctx context.Context) error {
	return m.server.Shutdown(ctx)
}
