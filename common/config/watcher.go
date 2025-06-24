package config

import (
	"context"
	"path"
	"path/filepath"
	"poker_server/common/pb"
	"poker_server/common/yaml"
	"poker_server/library/async"
	"poker_server/library/mlog"
	"poker_server/library/uerror"
	"poker_server/library/util"
	"sync"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type Watcher struct {
	sync.WaitGroup
	topic  string
	cpath  string
	client *clientv3.Client
	exit   chan struct{}
}

func NewWatcher(cfg *yaml.EtcdConfig, ccfg *yaml.CommonConfig) (*Watcher, error) {
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
	return &Watcher{
		client: cli,
		topic:  ccfg.ConfigTopic,
		cpath:  ccfg.ConfigPath,
		exit:   make(chan struct{}),
	}, nil
}

func (d *Watcher) Close() error {
	d.exit <- struct{}{}
	close(d.exit)
	d.Wait()
	return d.client.Close()
}

func (d *Watcher) Load(tmps map[string]struct{}) error {
	rsp, err := d.client.Get(context.Background(), d.topic, clientv3.WithPrefix())
	if err != nil {
		return uerror.New(1, pb.ErrorCode_REQUEST_FAIELD, "获取注册服务节点失败: %v", err)
	}

	for _, kv := range rsp.Kvs {
		sheet := path.Base(string(kv.Key))
		if f, ok := fileMgr[sheet]; ok {
			if err := f(string(kv.Value)); err != nil {
				return uerror.New(1, pb.ErrorCode_PARSE_FAILED, "加载%s配置错误： %v", sheet, err)
			}
			tmps[sheet] = struct{}{}

			if err := util.Save(d.cpath, sheet+".conf", kv.Value); err == nil {
				mlog.Infof("更新配置：%s", filepath.Join(d.cpath, sheet+".conf"))
			}
		}
	}
	return nil
}

func (d *Watcher) Watch(tmps map[string]struct{}) error {
	watchChan := d.client.Watch(context.Background(), d.topic, clientv3.WithPrefix())
	if watchChan == nil {
		return uerror.New(1, pb.ErrorCode_REQUEST_FAIELD, "Config监听服务失败: watchChan is nil")
	}

	wg := &sync.WaitGroup{}
	wg.Add(len(fileMgr) - len(tmps))

	async.SafeGo(mlog.Errorf, func() {
		for resp := range watchChan {
			if resp.Canceled {
				mlog.Errorf("Config监听被取消，尝试重新连接")
				return
			}
			if resp.Err() != nil {
				mlog.Errorf("Config监听服务失败: %v", resp.Err())
				continue
			}
			for _, event := range resp.Events {
				if event.Type != clientv3.EventTypePut {
					continue
				}

				sheet := path.Base(string(event.Kv.Key))
				f, ok := fileMgr[sheet]
				if !ok {
					continue
				}

				if err := util.Save(d.cpath, sheet+".conf", event.Kv.Value); err == nil {
					mlog.Infof("更新配置：%s", filepath.Join(d.cpath, sheet+".conf"))
				}

				if err := f(string(event.Kv.Value)); err != nil {
					mlog.Errorf("加载%s配置错误： %v", sheet, err)
				} else {
					if _, ok := tmps[sheet]; !ok {
						tmps[sheet] = struct{}{}
						wg.Done()
					}
					mlog.Infof("config_watcher 更新游戏配置：%s", sheet)
				}
			}
		}
	})
	wg.Wait()
	return nil
}
