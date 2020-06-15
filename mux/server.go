package mux

import (
	"context"
	"errors"
	"net/http"
	"time"
)

func (mux *Mux) Start(listenAt string) {
	mux.server = &http.Server{Handler: mux, Addr: listenAt}
	mux.Log().Infof("启动路由: %s", listenAt)
	go mux.doStart()
}

func (mux *Mux) doStart() {
	mux.wait.Add(1)
	defer mux.wait.Done()
	for {
		if err := mux.server.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				mux.Log().Errorf("启动路由 (出错): %v", err)
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				mux.cancel = cancel
				<-ctx.Done()
				if errors.Is(ctx.Err(), context.DeadlineExceeded) {
					mux.Log().Infof("启动路由 (重试)")
					continue
				}
			}
		}
		mux.Log().Infof("路由服务: 已关闭")
		return
	}
}

func (mux *Mux) Shutdown() {
	mux.updateCloseStatus(true)

	if mux.cancel != nil {
		mux.cancel()
	}

	mux.wait.Add(1)
	go func() {
		defer mux.wait.Done()
		defer mux.updateCloseStatus(false)
		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()
		if err := mux.server.Shutdown(shutdownCtx); err != nil {
			mux.Log().Errorf("关闭路由 (出错): %v", err)
		}
	}()
}

func (mux *Mux) Wait() {
	mux.wait.Wait()
	mux.Log().Errorf("路由服务: 已结束")
}

func (mux *Mux) updateCloseStatus(closing bool) {
	mux.locker.Lock()
	defer mux.locker.Unlock()
	mux.closing = closing
}
