package discovery

import (
	"context"
	"fmt"
	"path"
	"poker_server/common/pb"
	"poker_server/common/yaml"
	"poker_server/framework/domain"
	"poker_server/library/async"
	"poker_server/library/mlog"
	"poker_server/library/uerror"
	"poker_server/library/util"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/spf13/cast"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type Etcd struct {
	sync.WaitGroup
	topic  string
	client *clientv3.Client
	lease  clientv3.LeaseID
	exit   chan struct{}
}

func NewEtcd(cfg *yaml.EtcdConfig) (*Etcd, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:            cfg.Endpoints,
		DialTimeout:          5 * time.Second,
		DialKeepAliveTime:    30 * time.Second,
		DialKeepAliveTimeout: 3 * time.Second,
		MaxCallSendMsgSize:   10 * 1024 * 1024,
	})
	if err != nil {
		return nil, uerror.New(1, pb.ErrorCode_CONNECT_FAILED, "Etcd连接失败: %v", err)
	}
	return &Etcd{
		client: cli,
		topic:  cfg.Topic,
		exit:   make(chan struct{}),
	}, nil
}

// 监听服务
func (d *Etcd) Watch(cls domain.ICluster) error {
	// 获取当前所有服务节点
	rsp, err := d.client.Get(context.Background(), d.topic, clientv3.WithPrefix())
	if err != nil {
		return uerror.New(1, pb.ErrorCode_REQUEST_FAIELD, "获取注册服务节点失败: %v", err)
	}

	for _, kv := range rsp.Kvs {
		node := &pb.Node{}
		if err := proto.Unmarshal(kv.Value, node); err != nil {
			return uerror.New(1, pb.ErrorCode_UNMARSHAL_FAILED, "解析服务节点失败: %v", err)
		} else {
			if cls.Add(node) {
				mlog.Infof("添加服务节点: %s", node.String())
			}
		}
	}

	// 监听服务
	async.SafeGo(mlog.Errorf, func() {
		d.watch(cls)
	})
	return nil
}

func (d *Etcd) watch(cls domain.ICluster) {
	for {
		watchChan := d.client.Watch(context.Background(), d.topic, clientv3.WithPrefix())
		if watchChan == nil {
			mlog.Errorf("Etcd监听服务失败: watchChan is nil")
			time.Sleep(1 * time.Second)
			continue
		}
		for resp := range watchChan {
			if resp.Canceled {
				mlog.Errorf("Etcd监听被取消，尝试重新连接")
				break
			}
			if resp.Err() != nil {
				mlog.Errorf("Etcd监听服务失败: %v", resp.Err())
				continue
			}
			for _, event := range resp.Events {
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
	}
}

// 注册服务
func (d *Etcd) Register(nn *pb.Node, ttl int64) error {
	// 创建租约
	ctx := context.Background()
	if err := util.Retry(3, time.Second, func() error {
		rsp, err := d.client.Grant(ctx, ttl)
		if err != nil {
			return uerror.New(1, pb.ErrorCode_REQUEST_FAIELD, "租约失败：%v", err)
		}
		d.lease = clientv3.LeaseID(rsp.ID)
		return nil
	}); err != nil {
		return uerror.New(1, pb.ErrorCode_REQUEST_FAIELD, "Etcd获取租约失败, node:%v, error:%v", nn, err)
	}

	// 注册服务
	buf, err := proto.Marshal(nn)
	if err != nil {
		return uerror.New(1, pb.ErrorCode_MARSHAL_FAILED, "Etcd序列化服务节点失败, node:%v, error:%v", nn, err)
	}
	topic := fmt.Sprintf("%s/%d/%d", d.topic, nn.Type, nn.Id)
	if _, err := d.client.Put(ctx, topic, string(buf), clientv3.WithLease(d.lease)); err != nil {
		return uerror.New(1, pb.ErrorCode_REQUEST_FAIELD, "Etcd注册服务节点失败, node:%v, error:%v", nn, err)
	}

	// 定时保活
	d.Add(1)
	go func() {
		defer d.Done()
		d.keepAlive(ctx, nn, ttl)
	}()
	return nil
}

func (d *Etcd) keepAlive(ctx context.Context, nn *pb.Node, ttl int64) {
	keep, err := d.client.KeepAlive(ctx, d.lease)
	if err != nil {
		mlog.Errorf("保活初始化失败: %v", err)
		return
	}

	tt := time.NewTicker(time.Duration(ttl/2) * time.Second)
	defer tt.Stop()
	for {
		select {
		case _, ok := <-keep:
			if !ok {
				mlog.Warnf("保活异常,尝试重新注册")
				if err := d.Register(nn, ttl); err != nil {
					mlog.Errorf("重新注册失败: %v", err)
				}
				return
			}
		case <-tt.C:
			if _, err := d.client.TimeToLive(ctx, d.lease); err != nil {
				mlog.Errorf("租约检查失败: %v", err)
			}
		case <-d.exit:
			revokeCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			if _, err := d.client.Revoke(revokeCtx, d.lease); err != nil {
				mlog.Errorf("租约释放失败: %v", err)
			}
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
