package parser

import (
	"fmt"
	"poker_server/tools/cfgtool/internal/base"
	"poker_server/tools/cfgtool/internal/manager"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/spf13/cast"
)

func parseCmd(tab *base.Table) {
	cfg := manager.GetConfig(tab.Type)
	cmdEnum := manager.GetOrNewEnum("CMD")
	cmdClientEnum := manager.GetOrNewEnum("CMD_CLIENT")
	cmdEnum.FileName = "cmd"
	cmdClientEnum.FileName = "cmd/cmd"
	cmdEnum.AddValueByCmd("CMD_EMPTY", "空参数", 0)
	cmdClientEnum.AddValueByCmd("CMD_CLIENT_EMPTY", "空参数", 0)
	for i, row := range tab.Rows[3:] {
		lfields := len(cfg.Fields)
		if len(row) < lfields {
			continue
		}
		cmdVal := row[cfg.Fields["Cmd"].Position]
		if len(cmdVal) <= 0 {
			continue
		}
		if cast.ToUint32(cmdVal)%2 != 0 {
			fmt.Printf("%s第%d行配置错误，cmd值不能为奇数, %v\n", cfg.FileName, i+1, row)
			continue
		}
		desc := ""
		if lfields+1 <= len(row) {
			desc = row[lfields]
		}
		// 通知一定只有req
		if reqVal := row[cfg.Fields["Request"].Position]; len(reqVal) > 0 {
			manager.AddCmd(cast.ToUint32(cmdVal), reqVal)
			cmdClientEnum.AddValueByCmd(reqVal, desc, cast.ToInt32(cmdVal))
			reqVal = strings.ToUpper(strcase.ToSnake(reqVal))
			cmdEnum.AddValueByCmd(reqVal, desc, cast.ToInt32(cmdVal))
		}
		// 非notify一定有rsp
		if rspVal := row[cfg.Fields["Response"].Position]; len(rspVal) > 0 {
			manager.AddCmd(cast.ToUint32(cmdVal)+1, rspVal)
			cmdClientEnum.AddValueByCmd(rspVal, desc, cast.ToInt32(cmdVal)+1)
			rspVal = strings.ToUpper(strcase.ToSnake(rspVal))
			cmdEnum.AddValueByCmd(rspVal, desc, cast.ToInt32(cmdVal)+1)
		}
	}
}
