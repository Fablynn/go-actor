package main

/*
该工具从github个人仓库拷贝过来，仅共使用
*/

import (
	"bytes"
	"flag"
	"fmt"
	"path"
	"poker_server/tools/cfgtool/domain"
	"poker_server/tools/cfgtool/internal/base"
	"poker_server/tools/cfgtool/internal/manager"
	"poker_server/tools/cfgtool/internal/parser"

	"poker_server/tools/cfgtool/service"
)

func main() {
	flag.StringVar(&domain.XlsxPath, "xlsx", "./xlsx", "cfg文件目录")
	flag.StringVar(&domain.TextPath, "text", "./data", "数据文件目录")
	flag.StringVar(&domain.ProtoPath, "proto", "", "proto文件目录")
	flag.StringVar(&domain.JsonPath, "json", "", "数据文件目录")
	flag.StringVar(&domain.BytesPath, "bytes", "", "数据文件目录")
	flag.StringVar(&domain.CodePath, "code", "", "go代码文件目录")
	flag.StringVar(&domain.PbPath, "pb", "", "proto生成路径")
	flag.StringVar(&domain.ClientPath, "client", "", "客户端代码生成路径")
	flag.Parse()

	if len(domain.XlsxPath) <= 0 {
		panic("配置文件目录不能为空")
	}
	if len(domain.PbPath) > 0 {
		domain.PkgName = path.Base(domain.PbPath)
	}
	// 加载所有配置
	files, err := base.Glob(domain.XlsxPath, ".*\\.xlsx", true)
	if err != nil {
		panic(err)
	}
	// 解析所有文件
	if err := parser.ParseFiles(files...); err != nil {
		panic(err)
	}
	// 生成proto文件数据
	buf := bytes.NewBuffer(nil)
	if err := service.GenProto(buf); err != nil {
		panic(err)
	}
	if err := service.SaveProto(); err != nil {
		panic(err)
	}
	// 解析proto文件
	if err := manager.ParseProto(); err != nil {
		panic(err)
	}
	if err := service.GenData(); err != nil {
		panic(err)
	}
	if err := service.GenCode(buf); err != nil {
		panic(err)
	}
	if err := service.GenHttpKit(buf); err != nil {
		panic(err)
	}
}

func init() {
	flag.Usage = func() {
		flag.PrintDefaults()
		fmt.Fprint(flag.CommandLine.Output(), fmt.Sprintf(`
枚举类型说明：
E|道具类型-金币|PropertType|Coin|1	

配置规则说明：
@config|sheet@结构名|map:字段名[,字段名]:别名|group:字段名[,字段名]:别名
@struct|sheet@结构名
@enum|sheet
		`))
	}
}
