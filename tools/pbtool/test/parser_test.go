package test

import (
	"bytes"
	"go-actor/tools/pbtool/domain"
	"go-actor/tools/pbtool/internal/base"
	"go-actor/tools/pbtool/internal/parse"
	"go-actor/tools/pbtool/service"
	"testing"
)

func Test_Parser(t *testing.T) {
	domain.PbPath = "../../../common/pb/"
	domain.RedisPath = "../../../common/dao/repository/redis/"

	if len(domain.PbPath) <= 0 {
		panic("proto文件目录不能为空")
	}

	// 加载所有文件
	files, err := base.Glob(domain.PbPath, ".*\\.pb\\.go", true)
	if err != nil {
		panic(err)
	}

	// 解析所有文件
	if err := parse.ParseFiles(&parse.Parser{}, files...); err != nil {
		panic(err)
	}

	// 生成代码
	buf := bytes.NewBuffer(nil)
	service.GenString(buf)
	service.GenHash(buf)
}
