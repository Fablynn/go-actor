package uerror

import (
	"fmt"
	"path"
	"poker_server/common/pb"
	"runtime"
)

type UError struct {
	filename string       // 文件名
	line     int          // 文件行号
	funcname string       // 函数名
	code     pb.ErrorCode // 错误码
	msg      string       // 错误
}

func NEW(code pb.ErrorCode, head *pb.Head, format string, args ...interface{}) *UError {
	str := fmt.Sprintf("head:%v, %s", head, format)
	return &UError{
		code: code,
		msg:  fmt.Sprintf(str, args...),
	}
}

func New(depth int, code pb.ErrorCode, format string, args ...interface{}) *UError {
	// 获取调用堆栈
	pc, file, line, _ := runtime.Caller(depth)
	funcName := runtime.FuncForPC(pc).Name()
	// 返回错误
	return &UError{
		filename: file,
		line:     line,
		funcname: path.Base(funcName),
		code:     code,
		msg:      fmt.Sprintf(format, args...),
	}
}

func ToRspHead(err error) *pb.RspHead {
	switch vv := err.(type) {
	case *UError:
		return &pb.RspHead{Code: int32(vv.GetCode()), Msg: vv.GetMsg()}
	case nil:
		return nil
	}
	return &pb.RspHead{Code: int32(pb.ErrorCode_UNKNOWN), Msg: err.Error()}
}

func ToError(head *pb.RspHead) error {
	if head == nil || head.Code == int32(pb.ErrorCode_SUCCESS) {
		return nil
	}
	return &UError{code: pb.ErrorCode(head.Code), msg: head.Msg}
}

func (e *UError) GetCode() pb.ErrorCode {
	return e.code
}

func (e *UError) GetMsg() string {
	return e.msg
}

func (e *UError) Error() string {
	if len(e.filename) > 0 {
		return fmt.Sprintf("[%s]%s:%d\t%s\t%s", e.code.String(), e.filename, e.line, e.funcname, e.msg)
	}
	return fmt.Sprintf("[%s]\t%s", e.code.String(), e.msg)
}
