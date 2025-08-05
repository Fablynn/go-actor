GCFLAGS=-gcflags "all=-N -l"
SYSTEM=$(shell go env GOOS)

CONFIG_PATH=./gameconf
XLSX_PATH=${CONFIG_PATH}/rpg
PROTO_PATH=./protocol
CFG_CODE_PATH=./common/config/repository
PB_GO_PATH=./common/pb
REDIS_CODE_PATH=./common/redis/repository/
SERVER_PATH=./server
OUTPUT=./output

## 需要编译的服务
TARGET_ENV=local
TARGET=gate client game builder
LINUX=$(TARGET:%=%_linux)
BUILD=$(TARGET:%=%_build)
START=$(TARGET:%=%_start)
STOP=$(TARGET:%=%_stop)

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
	@echo "gen config code..."
	@rm -rf ${CFG_CODE_PATH}
	@make pb
	@go run ./tools/cfgtool/main.go -xlsx=${XLSX_PATH} -proto=${PROTO_PATH} -code=${CFG_CODE_PATH} -text=${CONFIG_PATH}/data -pb=${PB_GO_PATH}
	@make pbtool

pb:
	@echo "Building pb"
	@./env/scripts/pb_gen.sh

pbtool:
	@echo "gen redis code..."
	@go run ./tools/pbtool/main.go -pb=${PB_GO_PATH} -redis=${REDIS_CODE_PATH}

#=================================打包选项========================================
${TARGET_ENV}: %:
	@echo "Building $*"
	@cp -rf ${CONFIG_PATH}/data ./env/$*/* ./env/scripts/*.sh  ${OUTPUT}

#=================================服务启动，停止========================================
startall:
	-cd ./output && ./server.sh startall

stopall:
	-cd ./output && ./server.sh stopall

starthch:
	-cd ./output && ./server.sh starthch

stophch:
	-cd ./output && ./server.sh stophch



$(START): %_start: %
	-cd ./output && ./start.sh $* 1

$(STOP): %_stop: %
	-cd ./output && ./stop.sh $*

docker_stop:
	@echo "停止docker环境"
	-cd ./output && docker-compose -f docker_compose.yaml down

docker_run:
	@echo "启动docker环境"
	-cd ./output && docker-compose -f docker_compose.yaml up -d

