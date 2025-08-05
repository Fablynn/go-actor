#!/bin/bash

echo "关闭服务：${1}"

PIDS=$(ps -ef | grep ${1} | grep -v grep | awk '{print $2}')
if [ -z "${PIDS}" ]; then
    echo "服务已经关闭 ${1}"
    exit 1
fi

sleep 4
kill -15 ${PIDS} || true
exit 0