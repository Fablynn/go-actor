#!/bin/bash

if [ $# -lt 2 ]; then
    echo "eg: start.sh gate 2"
    exit 1
fi

if [ ! -f "./$1" ]; then
    echo "${1}不是可执行文件"
    exit 1
fi

sleep 4
mkdir -p ./log
nohup ./${1} -config=./config.yaml -id=${2} 2>./log/${1}${2}_monitor.log 1>/dev/null &
echo "启动服务成功：${1} -id=${2}"
exit 0