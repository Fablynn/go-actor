package main

import (
	"context"
	"flag"
	"fmt"
	"hash/crc32"
	"os"
	"path"
	"path/filepath"
	"poker_server/common/yaml"
	"poker_server/library/util"
	"strings"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func main() {
	var config, dataPath string
	flag.StringVar(&dataPath, "data", "./data", "配置文件目录")
	flag.StringVar(&config, "config", "config.yaml", "配置文件目录")
	flag.Parse()

	cfg, err := yaml.NewConfig(config)
	if err != nil {
		panic(err)
	}

	client, err := clientv3.New(clientv3.Config{
		Endpoints:            cfg.Etcd.Endpoints,
		DialTimeout:          5 * time.Second,
		DialKeepAliveTime:    30 * time.Second,
		DialKeepAliveTimeout: 3 * time.Second,
		MaxCallSendMsgSize:   1024 * 1024 * 1024,
	})
	if err != nil {
		panic(err)
	}
	defer client.Close()

	rsp, err := client.Get(context.Background(), cfg.Common.ConfigTopic, clientv3.WithPrefix())
	if err != nil {
		panic(err)
	}
	tmps := map[string]uint32{}
	for _, kv := range rsp.Kvs {
		sheet := filepath.Base(string(kv.Key))
		fmt.Println("已经上传文件: ", sheet)
		tmps[sheet] = crc32.ChecksumIEEE(kv.Value)
	}

	files, err := util.Glob(dataPath, ".*\\.conf", true)
	if err != nil {
		panic(err)
	}

	for _, filename := range files {
		buf, err := os.ReadFile(filename)
		if err != nil {
			panic(err)
		}
		sheet := filepath.Base(strings.TrimSuffix(filename, filepath.Ext(filename)))
		val := crc32.ChecksumIEEE(buf)
		old, ok := tmps[sheet]
		if ok && old == val {
			continue
		}

		_, err = client.Put(context.Background(), path.Join(cfg.Common.ConfigTopic, sheet), string(buf))
		if err != nil {
			panic(err)
		}
		fmt.Println("上传文件：", filepath.Base(filename))
	}
}
