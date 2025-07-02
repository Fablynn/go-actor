package dao

import (
	"go-actor/common/dao/internal/manager"
	"go-actor/common/dao/internal/mysql"
	"go-actor/common/yaml"
)

func RegisterMysqlTable(dbname string, tables ...interface{}) {
	manager.RegisterTable(dbname, tables...)
}

func InitMysql(cfgs map[int32]*yaml.DbConfig) error {
	return manager.InitMysql(cfgs)
}

func GetMysql(dbname string) *mysql.OrmSql {
	return manager.GetMysql(dbname)
}
