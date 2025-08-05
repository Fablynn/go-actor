package util

import (
	"go/format"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
)

// os.O_CREATE|os.O_APPEND|os.O_RDWR
func CreateFile(fileName string, flag int) (fb *os.File, err error) {
	// 判断路径是否存在
	pp := path.Dir(fileName)
	if err := os.MkdirAll(pp, os.FileMode(0755)); err != nil {
		return nil, err
	}
	// 创建文件
	if fb, err = os.OpenFile(fileName, flag, os.FileMode(0644)); err != nil {
		return nil, err
	}
	return
}

// 文件是否相同
func IsSameFile(fb *os.File, name string) bool {
	st2, _ := os.Stat(name)
	st1, _ := fb.Stat()
	return os.SameFile(st1, st2)
}

func Save(ppath, filename string, buf []byte) error {
	fileName := path.Join(ppath, filename)
	// 创建目录
	if err := os.MkdirAll(path.Dir(fileName), os.FileMode(0777)); err != nil {
		return err
	}
	// 写入文件
	if err := ioutil.WriteFile(fileName, buf, os.FileMode(0666)); err != nil {
		return err
	}
	return nil
}

func SaveGo(path, filename string, buf []byte) error {
	result, err := format.Source(buf)
	if err != nil {
		Save("./", "gen_error.gen.go", buf)
		return err
	}
	return Save(path, filename, result)
}

func Glob(dir, pattern string, recursive bool) (rets []string, err error) {
	pre, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if !recursive && info.IsDir() && dir != path {
			return filepath.SkipDir
		}
		if info.IsDir() {
			return nil
		}
		if pre.MatchString(path) {
			rets = append(rets, path)
		}
		return nil
	})
	return
}
