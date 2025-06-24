package base

import (
	"fmt"
	"strings"
)

type Field struct {
	Name string // 字段名字
	Type string // 分类
}

type Hash struct {
	Pkg     string   // 包名
	Name    string   // pb名字
	DbName  string   // 数据库名字
	KFormat string   // redis的key格式
	Keys    []*Field // 做成key格式的字段
	FFormat string   // redis的field格式
	Fields  []*Field // redis的field字段
	Desc    string   // 描述
}

func (d *Hash) Kargs() string {
	return args(d.Keys)
}

func (d *Hash) Kvalues() string {
	return values(d.Keys)
}

func (d *Hash) Fargs() string {
	return args(d.Fields)
}

func (d *Hash) Fvalues() string {
	return values(d.Fields)
}

func (d *Hash) Args() string {
	tmps := []*Field{}
	tmps = append(tmps, d.Keys...)
	tmps = append(tmps, d.Fields...)
	return args(tmps)
}

func (d *Hash) Values() string {
	tmps := []*Field{}
	tmps = append(tmps, d.Keys...)
	tmps = append(tmps, d.Fields...)
	return values(tmps)
}

type String struct {
	Pkg    string   // 包名
	Name   string   // pb名字
	DbName string   // 数据库名字
	Format string   // redis的key格式
	Keys   []*Field // 做成key格式的字段
	Desc   string   // 描述
}

func (d *String) Args() string {
	return args(d.Keys)
}

func (d *String) Values() string {
	return values(d.Keys)
}

func args(keys []*Field) string {
	if len(keys) <= 0 {
		return ""
	}
	args := make([]string, 0)
	for _, item := range keys {
		args = append(args, fmt.Sprintf("%s %s", item.Name, item.Type))
	}
	return strings.Join(args, ",")
}

func values(keys []*Field) string {
	if len(keys) <= 0 {
		return ""
	}
	values := make([]string, 0)
	for _, item := range keys {
		values = append(values, item.Name)
	}
	return strings.Join(values, ",")
}

type XORM struct {
	Pkg       string // 包名
	TableName string // xorm对象名字
}
