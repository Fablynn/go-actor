package service

import (
	"bytes"
	"go-actor/tools/cfgtool/domain"
	"go-actor/tools/cfgtool/internal/base"
	"go-actor/tools/cfgtool/internal/manager"
	"go-actor/tools/cfgtool/internal/templ"
	"go-actor/tools/library/uerror"
	"sort"
)

type ProtoInfo struct {
	RefList    []string
	EnumList   []*base.Enum
	StructList []*base.Struct
	ConfigList []*base.Config
}

func GenProto(buf *bytes.Buffer) error {
	// 根据文件分类
	tmps := map[string]*ProtoInfo{}
	for _, val := range manager.GetEnumList() {
		sort.Slice(val.ValueList, func(i, j int) bool {
			return val.ValueList[i].Value < val.ValueList[j].Value
		})
		if _, ok := tmps[val.FileName]; !ok {
			tmps[val.FileName] = &ProtoInfo{}
		}
		tmps[val.FileName].EnumList = append(tmps[val.FileName].EnumList, val)
	}

	for _, val := range manager.GetStructList() {
		if _, ok := tmps[val.FileName]; !ok {
			tmps[val.FileName] = &ProtoInfo{}
		}
		tmps[val.FileName].StructList = append(tmps[val.FileName].StructList, val)
	}

	for _, val := range manager.GetConfigList() {
		if _, ok := tmps[val.FileName]; !ok {
			tmps[val.FileName] = &ProtoInfo{}
		}
		tmps[val.FileName].ConfigList = append(tmps[val.FileName].ConfigList, val)
	}

	// 生成proto文件
	for fileName, data := range tmps {
		buf.Reset()
		data.RefList = manager.GetRefList(fileName)
		if err := templ.ProtoTpl.Execute(buf, data); err != nil {
			return uerror.New(1, -1, "gen proto file error: %s", err.Error())
		}
		manager.AddProto(fileName, buf)
	}
	return nil
}

func SaveProto() error {
	if len(domain.ProtoPath) <= 0 {
		return nil
	}

	for fileName, data := range manager.GetProtoMap() {
		if err := base.Save(domain.ProtoPath, fileName, []byte(data)); err != nil {
			return err
		}
	}
	return nil
}

// 客户端专用
func GenCmdProto() error {
	return nil
}
