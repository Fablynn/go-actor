package uerror

type ErrorCode int

var ErrorDesc = map[ErrorCode]string{
	ErrorCodeNone: "成功",
}

func (e *ErrorCode) String() string {
	if name, ok := ErrorDesc[*e]; ok {
		return name
	}
	return "unknown"
}

const (
	ErrorCodeNone ErrorCode = iota //start
)
