package mysql

import (
	"fmt"
	"poker_server/common/pb"
	"poker_server/common/yaml"
	"poker_server/library/uerror"
	"sync/atomic"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
)

type OrmSql struct {
	engine  *xorm.EngineGroup // 数据库引擎组
	dsn     []string          // 数据库连接字符串
	dbname  string            // 数据库名称
	isAlive int32             // 连接是否正常
}

func NewOrmSql(cfg *yaml.DbConfig, tables ...interface{}) *OrmSql {
	dsn := []string{}
	// 主节点
	dsn = append(dsn, fmt.Sprintf("%s:%s@tcp(%s)/%s?timeout=3s&parseTime=true&loc=Local&charset=utf8",
		cfg.User, cfg.Password, cfg.Host, cfg.DbName),
	)
	// 从节点配置
	for _, scfg := range cfg.Slave {
		dsn = append(dsn, fmt.Sprintf("%s:%s@tcp(%s)/%s?timeout=3s&parseTime=true&loc=Local&charset=utf8",
			scfg.User, scfg.Password, scfg.Host, scfg.DbName))
	}
	return &OrmSql{dsn: dsn, dbname: cfg.DbName}
}

func (o *OrmSql) Connect(tables ...interface{}) error {
	// 创建数据库引擎组
	eng, err := xorm.NewEngineGroup("mysql", o.dsn)
	if err != nil {
		return uerror.New(1, pb.ErrorCode_CONNECT_FAILED, "mysql连接失败:%v", err)
	}
	eng.SetMaxIdleConns(10)
	eng.SetMaxOpenConns(200)
	if len(tables) > 0 {
		eng.Sync2(tables...)
	}

	// 查看连接是否联通
	if err := eng.Ping(); err != nil {
		eng.Close()
		return uerror.New(1, pb.ErrorCode_PING_FAILED, "mysql连接失败:%v", err)
	}
	if o.engine != nil {
		o.engine.Close()
	}
	o.engine = eng
	atomic.StoreInt32(&o.isAlive, 1)
	return nil
}

// 检测连接是否联通
func (o *OrmSql) Ping() error {
	return o.engine.Ping()
}

func (o *OrmSql) IsAlive() bool {
	return atomic.LoadInt32(&o.isAlive) > 0
}

// 创建session
func (o *OrmSql) NewSession() *xorm.Session {
	return o.engine.NewSession()
}

func (o *OrmSql) GetEngine() *xorm.Engine {
	return o.engine.Engine
}
