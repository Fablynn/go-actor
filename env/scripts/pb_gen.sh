#!/usr/bin/env bash

GO_BIN="$(go env GOPATH)/bin"
SYSTEM=$(go env GOOS)
PROTO_PATH=../poker_protocol
PB_GO_PATH=./common/pb

rm -rf ${PB_GO_PATH} && mkdir -p ${PB_GO_PATH}

if [ "${SYSTEM}" == "windows" ]; then
    protoc.exe --plugin=protoc-gen-xorm.exe -I${PROTO_PATH} ${PROTO_PATH}/*.proto --go_opt paths=source_relative --go_out=${PB_GO_PATH} --xorm_out=${PB_GO_PATH}
    protoc-go-inject-tag.exe -input=${PB_GO_PATH}/*.pb.go -XXX_skip="state,sizeCache,unknownFields"
else
	protoc --plugin=${GO_BIN}/protoc-gen-xorm.exe -I${PROTO_PATH} ${PROTO_PATH}/*.proto --go_opt paths=source_relative --go_out=${PB_GO_PATH} --xorm_out=${PB_GO_PATH}
    protoc-go-inject-tag -input=${PB_GO_PATH}/*.pb.go -XXX_skip="state,sizeCache,unknownFields"
fi

# 使用 sed 批量添加忽略标签
if [ "${SYSTEM}" == "darwin" ]; then
    sed -i '' -E 's/(^[[:space:]]*(state|sizeCache|unknownFields)[[:space:]]+protoimpl\.[[:alpha:]]+)/\1 `xorm:"-"`/' ${PB_GO_PATH}/*.pb.go
    sed -i '' 's/`protogen:"open.v1"`//g' ${PB_GO_PATH}/*.pb.go
else 
    sed -i -E 's/(^\s*(state|sizeCache|unknownFields)\s+protoimpl\.[A-Za-z]+)/\1 `xorm:"-"`/' ${PB_GO_PATH}/*.pb.go
    sed -i 's/`protogen:"open.v1"`//g' ${PB_GO_PATH}/*.pb.go
fi
