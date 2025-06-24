package base

import (
	"go/format"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"poker_server/common/pb"
	"poker_server/library/uerror"
	"regexp"
)

// 遍历目录所有文件
func Glob(dir, pattern string, recursive bool) (rets []string, err error) {
	pre, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		// 不深度迭代
		if !recursive && info.IsDir() && dir != path {
			return filepath.SkipDir
		}
		// 过滤目录
		if info.IsDir() {
			return nil
		}
		// 是否配置
		if pre.MatchString(path) {
			rets = append(rets, path)
		}
		return nil
	})
	return
}

func SaveGo(path, filename string, buf []byte) error {
	result, err := format.Source(buf)
	if err != nil {
		Save("./", "gen_error.gen.go", buf)
		return uerror.New(1, pb.ErrorCode_UNKNOWN, "格式化失败: %v", err)
	}
	return Save(path, filename, result)
}

func Save(ppath, filename string, buf []byte) error {
	fileName := path.Join(ppath, filename)
	// 创建目录
	if err := os.MkdirAll(path.Dir(fileName), os.FileMode(0777)); err != nil {
		return uerror.New(1, pb.ErrorCode_UNKNOWN, "filename: %s, error: %v", fileName, err)
	}

	// 写入文件
	if err := ioutil.WriteFile(fileName, buf, os.FileMode(0666)); err != nil {
		return uerror.New(1, pb.ErrorCode_UNKNOWN, "filename: %s, error: %v", fileName, err)
	}
	return nil
}
