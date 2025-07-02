package test

import (
	"bytes"
	"go-actor/tools/cfgtool/domain"
	"go-actor/tools/cfgtool/internal/base"
	"go-actor/tools/cfgtool/internal/manager"
	"go-actor/tools/cfgtool/internal/parser"
	"go-actor/tools/cfgtool/service"
	"path"
	"testing"
)

func TestConfig(t *testing.T) {
	domain.XlsxPath = "../../../../poker_gameconf"
	domain.TextPath = "../gen/data"
	domain.ProtoPath = "../gen/proto"
	domain.CodePath = "../gen/code"
	domain.PbPath = "../../../common/pb"
	domain.PkgName = path.Base(domain.PbPath)

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
}
