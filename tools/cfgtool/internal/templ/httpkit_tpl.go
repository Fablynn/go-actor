package templ

const httpKitTpl = `
/*
* 本代码由cfgtool工具生成，请勿手动修改
 */

package {{.Pkg}}

import (
	"poker_server/common/pb"

	"github.com/golang/protobuf/proto"
)

var (
	cmds = make(map[uint32]func() proto.Message)
)

func init() {
	{{range $cmd, $pb := .Data -}}
	cmds[{{$cmd}}] = func() proto.Message { return &pb.{{$pb}}{} }
	{{end -}}
}

`
