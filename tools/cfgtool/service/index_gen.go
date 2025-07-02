package service

import (
	"bytes"
	"go-actor/tools/cfgtool/internal/manager"
	"go-actor/tools/cfgtool/internal/templ"
	"go-actor/tools/library/uerror"

	"sort"

	"go-actor/tools/cfgtool/domain"
	"go-actor/tools/cfgtool/internal/base"
)

type IndexInfo struct {
	Pkg       string
	IndexList []int
}

func genIndex(buf *bytes.Buffer) error {
	indexs := &IndexInfo{
		Pkg:       "pb",
		IndexList: manager.GetIndexMap(),
	}

	if len(indexs.IndexList) > 0 {
		sort.Slice(indexs.IndexList, func(i, j int) bool {
			return indexs.IndexList[i] < indexs.IndexList[j]
		})

		buf.Reset()
		if err := templ.IndexTpl.Execute(buf, indexs); err != nil {
			return uerror.New(1, -1, "gen index file error: %s", err.Error())
		}
		return base.SaveGo(domain.PbPath, "index.gen.go", buf.Bytes())
	}
	return nil
}
