package templ

const hashTpl = `
{{/* 定义全局变量  */}}
{{$pkg := .Pkg}}
{{$pb := .Name}} 
{{$dbname := .DbName}}
{{$kformat := .KFormat}}
{{$fformat := .FFormat}}

/*
* 本代码由pbtool工具生成，请勿手动修改
*/

package {{$pkg}}

import (
	"fmt"
	"poker_server/common/dao/internal/manager"
	"poker_server/common/pb"
	"poker_server/library/mlog"
	"poker_server/library/uerror"

	"github.com/golang/protobuf/proto"
)

const (
	DBNAME = "{{$dbname}}"
)

func GetKey({{.Kargs}}) string {
	return fmt.Sprintf("{{$kformat}}", {{range $i, $item := .Keys}} {{if $i}}, {{end}} {{$item.Name}} {{end}})
}

func GetField({{.Fargs}}) string {
	return fmt.Sprintf("{{$fformat}}", {{range $i, $item := .Fields}} {{if $i}}, {{end}} {{$item.Name}} {{end}})
}

func HGetAll({{.Kargs}}) (ret map[string]*pb.{{$pb}}, err error) {
	// 获取redis连接
	client := manager.GetRedis(DBNAME)
	if client == nil {
		err = uerror.New(1, pb.ErrorCode_CLIENT_NOT_FOUND, "redis数据库不存在: %s", DBNAME)
		return
	}
	key := GetKey({{.Kvalues}})

	// 加载数据
	kvs, err := client.HGetAll(key)
	if err != nil {
		return
	}

	// 解析数据
	ret = make(map[string]*pb.{{$pb}})
	for k, item := range kvs {
		if len(item) <= 0 {
			continue
		}

		data := &pb.{{$pb}}{}
		if err := proto.Unmarshal([]byte(item), data); err != nil {
			return nil, err
		}
		ret[k] = data
	}
	return
}

func HMGet({{.Kargs}} {{if .Keys}},{{end}} fields ...string) (map[string]*pb.{{$pb}}, error) {
	if len(fields) <= 0 {
		return nil, nil	
	}
	// 获取redis连接
	client := manager.GetRedis(DBNAME)
	if client == nil {
		return nil, uerror.New(1, pb.ErrorCode_CLIENT_NOT_FOUND, "redis数据库不存在: %s", DBNAME)
	}
	key := GetKey({{.Kvalues}})

	results, err := client.HMGet(key, fields...)
	if err != nil {
		return nil, err
	}

	ret := make(map[string]*pb.{{$pb}})
	for i, field := range fields {
		if results[i] == nil {
			continue
		}	
		buf, ok := results[i].([]byte)
		if !ok {
			mlog.Errorf("数据类型不支持: %s", field)	
			continue
		}
		item := &pb.{{$pb}}{}
		if err := proto.Unmarshal(buf, item); err != nil {
			mlog.Errorf("数据序列化失败：%s: %v", field, err)	
		}
		ret[field] = item
	}
	return ret, nil
}

func HMSet({{.Kargs}} {{if .Keys}},{{end}} data map[string]*pb.{{$pb}}) error {
	// 获取redis连接
	client := manager.GetRedis(DBNAME)
	if client == nil {
		return uerror.New(1, pb.ErrorCode_CLIENT_NOT_FOUND, "redis数据库不存在: %s", DBNAME)
	}
	key := GetKey({{.Kvalues}})

	// 设置数据
	vals := []interface{}{}
	for k, v := range data {
		buf, err := proto.Marshal(v)
		if err != nil {
			return err
		}
		vals = append(vals, k, buf)
	}
	return client.HMSet(key, vals...)
}

func HGet({{.Args}}) (*pb.{{$pb}}, bool, error) {
	// 获取redis连接
	client := manager.GetRedis(DBNAME)
	if client == nil {
		return nil, false, uerror.New(1, pb.ErrorCode_CLIENT_NOT_FOUND, "redis数据库不存在: %s", DBNAME)
	}
	key := GetKey({{.Kvalues}})
	field := GetField({{.Fvalues}})

	// 加载数据
	str, err := client.HGet(key, field)
	if err != nil {
		return nil, false, err
	}

	// 解析数据
	data := &pb.{{$pb}}{}
	if err := proto.Unmarshal([]byte(str), data); err != nil {
		return nil, len(str)>0, err
	}
	return data, len(str)>0, nil
}

func HSet({{.Args}}, data *pb.{{$pb}}) error {
	// 获取redis连接
	client := manager.GetRedis(DBNAME)
	if client == nil {
		return uerror.New(1, pb.ErrorCode_CLIENT_NOT_FOUND, "redis数据库不存在: %s", DBNAME)
	}
	key := GetKey({{.Kvalues}})
	field := GetField({{.Fvalues}})

	// 序列化数据
	buf, err := proto.Marshal(data)
	if err != nil {
		return err
	}

	// 设置数据
	return client.HSet(key, field, buf)
}

func HDel({{.Kargs}} {{if .Keys}},{{end}} fields ...string) error {
	// 获取redis连接
	client := manager.GetRedis(DBNAME)
	if client == nil {
		return uerror.New(1, pb.ErrorCode_CLIENT_NOT_FOUND, "redis数据库不存在: %s", DBNAME)
	}
	key := GetKey({{.Kvalues}})
	return client.HDel(key, fields...)
}

`
