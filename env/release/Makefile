
## 需要编译的服务
TARGET=gate room match 
START=$(TARGET:%=%_start)
STOP=$(TARGET:%=%_stop)

.PHONY: startall stopall docker_run docker_stop

#=================================服务启动，停止========================================
startall: ${START}

stopall: $(STOP) 

$(START): %_start: %
	@echo "running $*"
	@nohup ./$* -config=./config.yaml -id=1 >./log/$*_monitor.log 2>&1 &
	sleep 1

$(STOP): %_stop: %
	@echo "stopping $*"
	-kill -9 $$(ps -ef | grep $* | grep -v grep | awk '{print $$2}')

docker_stop:
	@echo "停止docker环境"
	docker-compose -f ./docker_compose.yaml down

docker_run:
	@echo "启动docker环境"
	docker-compose -f ./docker_compose.yaml up -d

