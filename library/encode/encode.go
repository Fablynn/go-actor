package encode

import (
	"bytes"
	"encoding/gob"
	"reflect"
	"sync"
)

type GobEncoder struct {
	buf *bytes.Buffer
	enc *gob.Encoder
}

type GobDecoder struct {
	buf *bytes.Buffer
	dec *gob.Decoder
}

var (
	encPool = sync.Pool{
		New: func() interface{} {
			buf := bytes.NewBuffer(make([]byte, 0, 1024))
			enc := gob.NewEncoder(buf)
			return &GobEncoder{buf: buf, enc: enc}
		},
	}
	decPool = sync.Pool{
		New: func() interface{} {
			buf := bytes.NewBuffer(make([]byte, 0, 1024))
			dec := gob.NewDecoder(buf)
			return &GobDecoder{buf: buf, dec: dec}
		},
	}
)

// 编码
func Encode(args ...interface{}) []byte {
	item := encPool.Get().(*GobEncoder)
	defer encPool.Put(item)
	item.buf.Reset()

	for _, arg := range args {
		item.enc.Encode(arg)
	}
	return item.buf.Bytes()
}

// 解码
func Decode(data []byte, mfun reflect.Method, pos int) (rets []reflect.Value, err error) {
	item := decPool.Get().(*GobDecoder)
	defer decPool.Put(item)
	item.buf.Reset()
	item.buf.Write(data)

	rets = make([]reflect.Value, mfun.Type.NumIn())
	for i := pos; i < mfun.Type.NumIn(); i++ {
		val := reflect.New(mfun.Type.In(i))
		if err = item.dec.DecodeValue(val); err != nil {
			return nil, err
		}
		rets[i] = val.Elem()
	}
	return
}
