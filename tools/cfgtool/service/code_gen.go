package service

import (
	"bytes"
	"go-actor/tools/cfgtool/domain"
	"go-actor/tools/cfgtool/internal/base"
	"go-actor/tools/cfgtool/internal/manager"
	"go-actor/tools/cfgtool/internal/templ"
	"path"

	"strings"

	"github.com/iancoleman/strcase"
)

type ConfigInfo struct {
	PbPkg string
	Pkg   string
	*base.Config
}

func GenCode(buf *bytes.Buffer) error {
	if len(domain.PbPath) <= 0 || len(domain.CodePath) <= 0 {
		return nil
	}

	// 生成索引
	if err := genIndex(buf); err != nil {
		return err
	}
	// 对文件分类
	for _, st := range manager.GetConfigMap() {
		buf.Reset()
		dataName := strings.TrimSuffix(st.Name, "Config")
		name := strcase.ToSnake(st.Name)
		item := &ConfigInfo{
			PbPkg:  domain.PkgName,
			Pkg:    name,
			Config: st,
		}
		if err := templ.CodeTpl.Execute(buf, item); err != nil {
			return err
		}
		// 保存代码
		if err := base.SaveGo(path.Join(domain.CodePath, name), dataName+"Data.gen.go", buf.Bytes()); err != nil {
			return err
		}
	}
	return nil
}
