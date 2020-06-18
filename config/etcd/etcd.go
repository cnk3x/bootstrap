package etcd

//import (
//	"bytes"
//	"context"
//	"fmt"
//
//	"go.etcd.io/etcd/v3/clientv3"
//	"go.shu.run/bootstrap/config"
//	"go.shu.run/bootstrap/utils/strs"
//)
//
//func New(client *clientv3.Client) config.Loader {
//	return &loader{client: client}
//}
//
//type loader struct {
//	client *clientv3.Client
//}
//
//var _ config.Loader = (*loader)(nil)
//
////监听一个对象
//func (c *loader) Watch(ctx context.Context, cfg config.cfg) error {
//	fmt.Printf("监听配置：%s\n", cfg.Key())
//	s, err := c.client.Get(ctx, cfg.Key(), clientv3.WithPrefix())
//	if err == nil {
//		for _, kv := range s.Kvs {
//			cfg.OnUpdate(string(kv.Key), string(kv.Value))
//		}
//	}
//
//	w := c.client.Watch(ctx, cfg.Key(), clientv3.WithPrefix())
//
//	go func() {
//		for {
//			select {
//			case <-ctx.Done():
//				return
//			case n := <-w:
//				for _, e := range n.Events {
//					fmt.Println(e.Type.String(), strs.B(e.Kv.Key), strs.B(e.Kv.Value))
//					c.notifyUpdate(cfg, e)
//				}
//			}
//		}
//	}()
//
//	return nil
//}
//
//func (c *loader) notifyUpdate(cfg config.cfg, e *clientv3.Event) {
//	if e.IsModify() {
//		if bytes.Equal(e.Kv.Value, e.PrevKv.Value) {
//			cfg.OnUpdate(string(e.Kv.Key), string(e.Kv.Value))
//		}
//		return
//	}
//	if e.IsCreate() {
//		cfg.OnUpdate(string(e.Kv.Key), string(e.Kv.Value))
//		return
//	}
//}
