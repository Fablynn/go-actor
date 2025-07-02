package uerror

import (
	"fmt"
	"path"
	"runtime"
)

type UError struct {
	filename string    // 文件名
	line     int       // 文件行号
	funcname string    // 函数名
	code     ErrorCode // 错误码
	msg      string    // 错误
}

func (e *UError) GetCode() ErrorCode {
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

func New(depth int, code ErrorCode, format string, args ...interface{}) *UError {
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
