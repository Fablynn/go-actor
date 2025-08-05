package encode

import (
	"bytes"
	"encoding/gob"
	"go-actor/common/pb"
	"go-actor/library/uerror"
	"reflect"
	"sync"

	"google.golang.org/protobuf/proto"
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
func Decode(data []byte, rets []reflect.Value, mfun reflect.Method, begin int, ends ...int) error {
	item := decPool.Get().(*GobDecoder)
	defer decPool.Put(item)
	item.buf.Reset()
	item.buf.Write(data)
	end := mfun.Type.NumIn()
	if len(ends) > 0 {
		end = ends[0]
	}
	for i := begin; i < end; i++ {
		val := reflect.New(mfun.Type.In(i))
		if err := item.dec.DecodeValue(val); err != nil {
			return err
		}
		rets[i] = val.Elem()
	}
	return nil
}

func Marshal(args ...interface{}) ([]byte, error) {
	if len(args) == 1 {
		switch vv := args[0].(type) {
		case []byte:
			return vv, nil
		case proto.Message:
			buf, err := proto.Marshal(vv)
			if err != nil {
				return nil, uerror.New(pb.ErrorCode_MARSHAL_FAILED, "序列化失败：%v", err)
			}
			return buf, nil
		}
	}
	return Encode(args...), nil
}
