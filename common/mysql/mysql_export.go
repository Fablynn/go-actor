package mysql

import (
	"go-actor/common/mysql/internal/client"
	"go-actor/common/mysql/internal/manager"
	"go-actor/common/yaml"
)

func Register(dbname string, tables ...interface{}) {
	manager.RegisterTable(dbname, tables...)
}

func Init(cfgs map[int32]*yaml.DbConfig) error {
	return manager.InitMysql(cfgs)
}

func GetClient(dbname string) *client.OrmSql {
	return manager.GetMysql(dbname)
}
