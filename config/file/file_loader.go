package file

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"go.shu.run/bootstrap/config"

	"github.com/fsnotify/fsnotify"
	"go.uber.org/fx"
)

func FileLoader(lc fx.Lifecycle) (config.Loader, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	l := &loader{w: w, notify: make(map[string][]func(key string, val string))}

	lc.Append(fx.Hook{
		OnStart: l.start,
		OnStop:  l.stop,
	})

	return l, nil
}

var _ config.Loader = (*loader)(nil)

type loader struct {
	w      *fsnotify.Watcher
	notify map[string][]func(key string, val string)
	cancel context.CancelFunc
}

func (l *loader) Watch(_ context.Context, cfg config.Config) error {
	fn := filepath.Join("conf", cfg.Key()+".toml")
	_ = l.w.Remove(fn)
	if err := l.w.Add(fn); err != nil {
		return err
	}
	l.notify[fn] = append(l.notify[fn], cfg.OnUpdate)
	return nil
}

func (l *loader) start(ctx context.Context) error {
	ctx, l.cancel = context.WithCancel(ctx)
	go func() {
		select {
		case <-ctx.Done():
			return
		case err := <-l.w.Errors:
			fmt.Println(err.Error())
		case e := <-l.w.Events:
			l.load(e.Name)
		}
	}()
	return nil
}

func (l *loader) stop(_ context.Context) error {
	l.cancel()
	return nil
}

func (l *loader) load(key string) {
	fmt.Printf("config: %s..changed...\n", key)
	if len(l.notify) > 0 {
		if n, find := l.notify[key]; find {
			v, err := ioutil.ReadFile(key)
			if err == nil {
				for i := range n {
					n[i](key, string(v))
				}
			}
		}
	}
}
