package service

import (
	"bytes"
	"fmt"
	"path"
	"poker_server/tools/pbtool/domain"
	"poker_server/tools/pbtool/internal/base"
	"poker_server/tools/pbtool/internal/manager"
	"poker_server/tools/pbtool/internal/templ"
)

func GenHash(buf *bytes.Buffer) error {
	if len(domain.RedisPath) <= 0 {
		return nil
	}
	manager.WalkHash(func(st *base.Hash) bool {
		buf.Reset()
		if err := templ.HashTpl.Execute(buf, st); err != nil {
			fmt.Printf("生成失败：%v, error: %v\n", st, err)
			return true
		}

		// 保存代码
		dst := path.Join(domain.RedisPath, st.Pkg)
		if err := base.SaveGo(dst, st.Name+".gen.go", buf.Bytes()); err != nil {
			fmt.Printf("生成失败：%v\n", st)
		}
		return true
	})
	return nil
}
