package config
//
//import (
//	"bytes"
//	"context"
//	"io/ioutil"
//	"time"
//
//	"github.com/fsnotify/fsnotify"
//)
//
//type Loader interface {
//	Update(v []byte) error
//}
//
//type cfg struct {
//	last   []byte
//	cancel context.CancelFunc
//}
//
//func (w *cfg) Update(fn string, loader Loader) error {
//	v, err := ioutil.ReadFile(fn)
//	if err != nil {
//		return err
//	}
//	return loader.Update(v)
//}
//
//func (w *cfg) Watch(ctx context.Context, fn string, loader Loader) error {
//	n, err := fsnotify.NewWatcher()
//	if err != nil {
//		return err
//	}
//
//	if err := n.Add(fn); err != nil {
//		return err
//	}
//
//	if w.cancel != nil {
//		w.cancel()
//	}
//	w.cancel = func() {}
//
//	for {
//		select {
//		case <-ctx.Done():
//			w.cancel()
//			return ctx.Err()
//		case <-n.Events:
//			w.cancel()
//			w.cancel = w.delayRun(ctx, func() {
//				v, err := ioutil.ReadFile(fn)
//				if err != nil {
//					return
//				}
//				if !bytes.Equal(w.last, v) {
//					if err := loader.Update(v); err != nil {
//						return
//					}
//				}
//			}, time.Second*3)
//		}
//	}
//}
//
//func (w *cfg) delayRun(ctx context.Context, run func(), timeout time.Duration) context.CancelFunc {
//	ctx, cancel := context.WithTimeout(ctx, timeout)
//	go func() {
//		select {
//		case <-ctx.Done():
//			if ctx.Err() == context.DeadlineExceeded {
//				run()
//			}
//		}
//	}()
//	return cancel
//}
