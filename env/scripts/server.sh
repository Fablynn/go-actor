#!/bin/bash

startall(){
    ./start.sh builder 1
    ./start.sh game 1
    ./start.sh gate 1
    ./start.sh client 1
}

stopall(){
    ./stop.sh builder
    ./stop.sh game
    ./stop.sh gate
    ./stop.sh client
    return 0
}

case $1 in
# 测试使用
startall)
    startall
    ;;
stopall)
    stopall
    ;;
restartall)
    stopall
    startall
    ;;
esac
