package dao

import (
	"poker_server/common/dao/internal/manager"
	"poker_server/common/dao/internal/mysql"
	"poker_server/common/yaml"
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
