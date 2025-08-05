
## 需要编译的服务
#TARGET=gate room match client
TARGET=gate room match client
START=$(TARGET:%=%_start)
STOP=$(TARGET:%=%_stop)

.PHONY: startall stopall docker_run docker_stop

#=================================服务启动，停止========================================
startall:
	./server.sh hch

stopall:
	./server.sh stopall

$(START): %_start: %
	./start.sh $* 1

$(STOP): %_stop: %
	./stop.sh $*

docker_stop:
	@echo "停止docker环境"
	docker-compose -f ./docker_compose.yaml down

docker_run:
	@echo "启动docker环境"
	docker-compose -f ./docker_compose.yaml up -d

