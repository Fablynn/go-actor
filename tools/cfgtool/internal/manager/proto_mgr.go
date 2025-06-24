package manager

import (
	"bytes"
	"poker_server/library/uerror"
	"poker_server/tools/cfgtool/domain"
	"poker_server/tools/cfgtool/internal/base"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/jhump/protoreflect/dynamic"
)

var (
	referenceMgr = make(map[string]map[string]struct{})
	protoMgr     = make(map[string]string)
	protoList    = []string{}
	descMap      = make(map[string]*desc.FileDescriptor)
)

func Clear() {
	referenceMgr = nil
	protoMgr = nil
	protoList = nil
	descMap = nil
}

func AddRef(filename string, reference map[string]struct{}) {
	val, ok := referenceMgr[filename]
	if !ok {
		val = make(map[string]struct{})
		referenceMgr[filename] = val
	}
	for ke := range reference {
		val[ke] = struct{}{}
	}
}

func GetRefList(file string) (rets []string) {
	for key := range referenceMgr[file] {
		rets = append(rets, key)
	}
	return
}

func AddProto(file string, buf *bytes.Buffer) {
	filename := base.GetProtoName(file)
	protoMgr[filename] = buf.String()
	protoList = append(protoList, filename)
}

func GetProtoList() []string {
	return protoList
}

func GetProtoMap() map[string]string {
	return protoMgr
}

func ParseProto() error {
	paser := protoparse.Parser{Accessor: protoparse.FileContentsFromMap(protoMgr)}
	descs, err := paser.ParseFiles(protoList...)
	if err != nil {
		return uerror.New(1, -1, "parse proto file error: %v", err)
	}
	for i := range protoList {
		descMap[protoList[i]] = descs[i]
	}
	return nil
}

func NewProto(fileName, name string) *dynamic.Message {
	val, ok := descMap[base.GetProtoName(fileName)]
	if !ok {
		return nil
	}
	typeOf := val.FindMessage(domain.ProtoPkgName + "." + name)
	if typeOf == nil {
		return nil
	}
	return dynamic.NewMessage(typeOf)
}
