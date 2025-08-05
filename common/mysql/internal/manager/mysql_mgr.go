package manager

import (
	"go-actor/common/mysql/internal/client"
	"go-actor/common/pb"
	"go-actor/common/yaml"
	"go-actor/library/mlog"
	"go-actor/library/safe"
	"go-actor/library/uerror"
	"sync"
	"time"
)

var (
	mysqlPool = &MysqlPool{
		pool:   make(map[string]*client.OrmSql),
		tables: make(map[string][]interface{}),
	}
)

type MysqlPool struct {
	mutex  sync.RWMutex
	pool   map[string]*client.OrmSql
	tables map[string][]interface{}
}

// 注册数据库表
func RegisterTable(dbname string, tables ...interface{}) {
	if len(tables) <= 0 {
		return
	}
	mysqlPool.tables[dbname] = append(mysqlPool.tables[dbname], tables...)
}

func InitMysql(cfgs map[int32]*yaml.DbConfig) error {
	if len(cfgs) <= 0 {
		return uerror.New(pb.ErrorCode_CONFIG_NOT_FOUND, "mysql配置为空")
	}
	// 初始化
	for _, cfg := range cfgs {
		client := client.NewOrmSql(cfg)
		if err := client.Connect(mysqlPool.tables[cfg.DbName]...); err != nil {
			return uerror.New(pb.ErrorCode_CONNECT_FAILED, "mysql连接失败, cfg:%v, error:%v", cfg, err)
		}
		mysqlPool.pool[cfg.DbName] = client
	}
	safe.Go(checkMysql)
	return nil
}

func GetMysql(dbname string) *client.OrmSql {
	mysqlPool.mutex.RLock()
	client, ok := mysqlPool.pool[dbname]
	mysqlPool.mutex.RUnlock()
	if ok && client.IsAlive() {
		return client
	}
	return nil
}

func checkMysql() {
	tt := time.NewTicker(30 * time.Second)
	defer tt.Stop()

	for {
		<-tt.C
		// 获取所有连接
		tmps := []*client.OrmSql{}
		mysqlPool.mutex.RLock()
		for _, client := range mysqlPool.pool {
			tmps = append(tmps, client)
		}
		mysqlPool.mutex.RUnlock()

		// 检测连通信
		for _, client := range tmps {
			if err := client.Ping(); err == nil {
				continue
			} else {
				mlog.Errorf("mysql连接异常断开: %v", err)
			}
			// 重新连接
			if err := client.Connect(); err != nil {
				mlog.Errorf("mysql重新连接失败, error:%v", err)
			}
		}
	}
}
