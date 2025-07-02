GCFLAGS=-gcflags "all=-N -l"
SYSTEM=$(shell go env GOOS)

CONFIG_PATH=./gameconf
XLSX_PATH=${CONFIG_PATH}/xlsx
PROTO_PATH=./protocol
CFG_CODE_PATH=./common/config/repository
PB_GO_PATH=./common/pb
REDIS_CODE_PATH=./common/dao/repository/redis
CLIENT_GO_PATH=./server/client/internal/manager
SERVER_PATH=./server
OUTPUT=./output

## 需要编译的服务
TARGET_ENV=local develop release pre
TARGET=gate db builder match room game gm #client
STOP_TARGET=gate gm builder match room game db
LINUX=$(TARGET:%=%_linux)
BUILD=$(TARGET:%=%_build)
START=$(TARGET:%=%_start)
STOP=$(STOP_TARGET:%=%_stop)

.PHONY: all build linux clean ${TARGET} config pb pbtool docker_run docker_stop ${TARGET_ENV}

all: clean
	make ${BUILD}

linux: clean
	make ${LINUX}

build: clean ${BUILD}

clean:
	-rm -rf ${OUTPUT}
	@mkdir -p ${OUTPUT}/log

$(LINUX): %_linux: %
	@echo "Building $*"
	CGO_ENABLE=0 GOOS=linux GOARCH=amd64 go build ${GCFLAGS} -o ${OUTPUT}/ ${SERVER_PATH}/$*

$(BUILD): %_build: %
	@echo "Building $*"
	go build ${GCFLAGS} -o ${OUTPUT}/$* ${SERVER_PATH}/$*


#=================================proto转go代码========================================
config:
	@echo "gen enum config code..."
	@go run ./tools/cfgtool/main.go -xlsx=${XLSX_PATH}/router.xlsx -proto=${PROTO_PATH} -code=${CFG_CODE_PATH} -text=${CONFIG_PATH}/data -pb=${PB_GO_PATH} -client=${CLIENT_GO_PATH}
	@echo "gen config code..."
	@rm -rf ${CFG_CODE_PATH}
	@make pb
	@go run ./tools/cfgtool/main.go -xlsx=${XLSX_PATH} -proto=${PROTO_PATH} -code=${CFG_CODE_PATH} -text=${CONFIG_PATH}/data -pb=${PB_GO_PATH} -client=${CLIENT_GO_PATH}
	@make pbtool

pb:
	@echo "Building pb"
	@./env/scripts/pb_gen.sh

pbtool:
	@echo "gen redis code..."
	@go run ./tools/pbtool/main.go -pb=${PB_GO_PATH} -redis=${REDIS_CODE_PATH}

#=================================打包选项========================================
${TARGET_ENV}: %: %
	@echo "Building $*"
	@cp -rf ${CONFIG_PATH}/data ./env/$*/*  ${OUTPUT}

#=================================服务启动，停止========================================
startall: ${START}

stopall: $(STOP)

$(START): %_start: %
	@echo "running $*"
	sleep 3
	-cd ./output && nohup ./$* -config=./config.yaml -id=1 >./log/$*_monitor.log 2>&1 &

$(STOP): %_stop: %
	@echo "stopping $*"
ifeq (${SYSTEM}, darwin)
	-kill -9 $$(ps -ef | grep $* | grep -v grep | grep "config.yaml" | awk '{print $$2}')
else # linux darwin(mac)
	-kill -9 $$(ps -ef | grep $* | grep -v grep | awk '{print $$2}')
endif

docker_stop:
	@echo "停止docker环境"
	docker-compose -f ./env/local/docker_compose.yaml down

docker_run:
	@echo "启动docker环境"
	docker-compose -f ./env/local/docker_compose.yaml up -d

