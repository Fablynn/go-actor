package util

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"poker_server/common/pb"
	"poker_server/library/uerror"
	"reflect"
	"regexp"
	"sync"
	"time"
)

func Ifelse[T any](flag bool, a, b T) T {
	if flag {
		return a
	}
	return b
}

func Index[T any](arr []T, pos int, def T) T {
	ll := len(arr)
	if ll <= 0 || pos < 0 || pos >= ll {
		return def
	}
	return arr[pos]
}

func Pool[T any]() *sync.Pool {
	return &sync.Pool{
		New: func() interface{} {
			return new(T)
		},
	}
}

func PoolSlice[T any](size int) *sync.Pool {
	return &sync.Pool{
		New: func() interface{} {
			return make([]T, size)
		},
	}
}

func Retry(attempts int, sleep time.Duration, f func() error) (err error) {
	for i := 0; i < attempts; i++ {
		if err = f(); err == nil {
			return
		}

		time.Sleep(sleep)
		sleep *= 2
	}
	return
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

func Save(ppath, filename string, buf []byte) error {
	fileName := path.Join(ppath, filename)
	if err := os.MkdirAll(path.Dir(fileName), os.FileMode(0777)); err != nil {
		return uerror.New(1, pb.ErrorCode_SYSTEM_CALL_FAILED, "filename: %s, error: %v", fileName, err)
	}

	if err := ioutil.WriteFile(fileName, buf, os.FileMode(0666)); err != nil {
		return uerror.New(1, pb.ErrorCode_WRITE_FAIELD, "filename: %s, error: %v", fileName, err)
	}
	return nil
}

func SliceIsset[T any](dst T, src []T) bool {
	for i := range src {
		if reflect.DeepEqual(src[i], dst) {
			return true
		}
	}
	return false
}
