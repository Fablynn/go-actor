package util

import (
	"reflect"
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

func ArrayPool[T any](size int) *sync.Pool {
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

func SliceIsset[T any](dst T, src []T) bool {
	for i := range src {
		if reflect.DeepEqual(src[i], dst) {
			return true
		}
	}
	return false
}

func Must(err error) {
	if err != nil {
		panic(err)
	}
}
