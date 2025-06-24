package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"poker_server/common/pb"
	"poker_server/common/yaml"
	"poker_server/library/uerror"
)

var (
	fileMgr = make(map[string]func(string) error)
	watcher *Watcher
)

func Register(sheet string, f func(string) error) {
	if _, ok := fileMgr[sheet]; ok {
		panic(fmt.Sprintf("%s已经注册过了", sheet))
	}
	fileMgr[sheet] = f
}

func Init(cfg *yaml.EtcdConfig, ccfg *yaml.CommonConfig) (err error) {
	tmps := make(map[string]struct{})
	if err := InitConfig(ccfg.ConfigPath, tmps); err != nil {
		return err
	}
	if ccfg.ConfigIsRemote {
		if watcher, err = NewWatcher(cfg, ccfg); err != nil {
			return
		}
		if err := watcher.Load(tmps); err != nil {
			return err
		}
		if err := watcher.Watch(tmps); err != nil {
			return err
		}
	}
	return
}

func InitConfig(cpath string, tmps map[string]struct{}) error {
	for sheet, f := range fileMgr {
		filename := path.Join(cpath, sheet+".conf")
		if _, err := os.Stat(filename); err != nil {
			if os.IsNotExist(err) {
				continue
			}
		}
		buf, err := ioutil.ReadFile(filename)
		if err != nil {
			return uerror.New(1, pb.ErrorCode_OPEN_FILE_FAILED, sheet)
		}
		if err := f(string(buf)); err != nil {
			return uerror.New(1, pb.ErrorCode_PARSE_FAILED, "加载%s配置错误： %v", sheet, err)
		}
		if tmps != nil {
			tmps[sheet] = struct{}{}
		}
	}
	return nil
}

func Close() error {
	if watcher != nil {
		return watcher.Close()
	}
	return nil
}
