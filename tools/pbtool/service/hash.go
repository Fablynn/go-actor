package service

import (
	"bytes"
	"fmt"
	"go-actor/tools/pbtool/domain"
	"go-actor/tools/pbtool/internal/base"
	"go-actor/tools/pbtool/internal/manager"
	"go-actor/tools/pbtool/internal/templ"
	"path"
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
