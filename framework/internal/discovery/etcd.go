package discovery

import (
	"context"
	"fmt"
	"go-actor/common/pb"
	"go-actor/common/yaml"
	"go-actor/framework/define"
	"go-actor/library/mlog"
	"go-actor/library/safe"
	"go-actor/library/uerror"
	"go-actor/library/util"
	"path"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/spf13/cast"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type Etcd struct {
	sync.WaitGroup
	client *clientv3.Client
	exit   chan struct{}
	topic  string
}

func NewEtcd(cfg *yaml.EtcdConfig) (ret *Etcd, err error) {
	err = util.Retry(3, time.Second, func() error {
		client, err := clientv3.New(clientv3.Config{
			Endpoints:            cfg.Endpoints,
			DialTimeout:          5 * time.Second,
			DialKeepAliveTime:    30 * time.Second,
			DialKeepAliveTimeout: 3 * time.Second,
			MaxCallSendMsgSize:   10 * 1024 * 1024,
		})
		if err == nil {
			ret = &Etcd{
				client: client,
				topic:  cfg.Topic,
				exit:   make(chan struct{}),
			}
		}
		return err
	})
	return
}

func (d *Etcd) Watch(cls define.INode) (err error) {
	ctx := context.Background()
	tctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	rsp, err := d.client.Get(tctx, d.topic, clientv3.WithPrefix())
	if err != nil {
		return err
	}

	for _, kv := range rsp.Kvs {
		node := &pb.Node{}
		if err := proto.Unmarshal(kv.Value, node); err != nil {
			return err
		}
		if cls.Add(node) {
			mlog.Infof("添加服务节点：%s", node.String())
		}
	}

	safe.Go(func() {
		for {
			wchan := d.client.Watch(ctx, d.topic, clientv3.WithPrefix())
			if wchan == nil {
				mlog.Infof("Etcd监听%s失败", d.topic)
				time.Sleep(time.Second)
				continue
			}
			for rsp := range wchan {
				if rsp.Canceled {
					mlog.Errorf("Etcd 监听被取消，尝试重新连接")
					break
				}
				if rsp.Err() != nil {
					mlog.Errorf("Etcd 监听出现错误: %v", rsp.Err().Error())
					continue
				}
				d.handleEvent(cls, rsp.Events...)
			}
		}
	})
	return nil
}

func (d *Etcd) handleEvent(cls define.INode, events ...*clientv3.Event) {
	for _, event := range events {
		switch event.Type {
		case clientv3.EventTypePut:
			node := &pb.Node{}
			if err := proto.Unmarshal(event.Kv.Value, node); err != nil {
				mlog.Errorf("解析服务节点失败: %v", err)
			} else {
				if cls.Add(node) {
					mlog.Infof("添加服务节点: %v", node)
				}
			}
		case clientv3.EventTypeDelete:
			key := string(event.Kv.Key)
			id := cast.ToInt32(path.Base(key))
			typ := pb.NodeType(cast.ToInt32(path.Base(path.Dir(key))))
			if cls.Del(typ, id) {
				mlog.Infof("删除服务节点: %s", key)
			}
		}
	}
}

func (d *Etcd) Register(cls define.INode, ttl int64) error {
	ctx := context.Background()
	tctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// 创建租约
	var lease clientv3.LeaseID
	if err := util.Retry(3, time.Second, func() error {
		rsp, err := d.client.Grant(tctx, ttl)
		if err == nil {
			lease = rsp.ID
		}
		return err
	}); err != nil {
		return err
	}

	// 注册服务
	nn := cls.GetSelf()
	buf, err := proto.Marshal(nn)
	if err != nil {
		return uerror.New(pb.ErrorCode_MARSHAL_FAILED, "Etcd序列化服务节点失败, node:%v, error:%v", nn, err)
	}
	topic := fmt.Sprintf("%s/%d/%d", d.topic, nn.Type, nn.Id)
	if _, err := d.client.Put(ctx, topic, string(buf), clientv3.WithLease(lease)); err != nil {
		return uerror.New(pb.ErrorCode_REQUEST_FAIELD, "Etcd注册服务节点失败, node:%v, error:%v", nn, err)
	}

	// 保活
	keep, err := d.client.KeepAlive(ctx, lease)
	if err != nil {
		return uerror.New(pb.ErrorCode_REQUEST_FAIELD, "保活初始化失败: %v", err)
	}

	// 定时保活
	d.Add(1)
	safe.Go(func() {
		defer d.Done()
		d.keepAlive(cls, ttl, keep, lease)
	})
	return nil
}

func (d *Etcd) keepAlive(cls define.INode, ttl int64, keep <-chan *clientv3.LeaseKeepAliveResponse, lease clientv3.LeaseID) {
	tt := time.NewTicker(5 * time.Second)
	defer func() {
		tt.Stop()
		d.client.Revoke(context.Background(), lease)
	}()
	for {
		select {
		case _, ok := <-keep:
			if !ok {
				mlog.Errorf("保活失败, 尝试重新注册保活")
				if err := d.Register(cls, ttl); err != nil {
					mlog.Errorf("尝试重新注册服务失败 %v", err)
				} else {
					return
				}
			}
		case <-tt.C:
			if _, err := d.client.TimeToLive(context.Background(), lease); err != nil {
				mlog.Errorf("保活异常, 尝试重新注册 %v", err)
				if err := d.Register(cls, ttl); err != nil {
					mlog.Errorf("尝试重新注册服务失败 %v", err)
				} else {
					return
				}
			}
		case <-d.exit:
			return
		}
	}
}

// 关闭服务
func (d *Etcd) Close() error {
	close(d.exit)
	d.Wait()
	return d.client.Close()
}
