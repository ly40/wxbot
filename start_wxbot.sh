#!/bin/bash

prog="./wxbot"
port="8256"
logfile="./app.log"

start() {
    pid=$(lsof -ti:"$port")
    if [ ! -z "$pid" ]; then
        echo "wxbot is already running with PID $pid"
        return 1
    fi

    nohup "$prog" 2>&1 >"$logfile" &
    pid=$!
    echo "wxbot started with PID $pid"
}

stop() {
    pid=$(lsof -ti:"$port")
    if [ -z "$pid" ]; then
        echo "wxbot is not running"
        return 1
    fi

    kill "$pid"
    echo "wxbot stopped"
}

status() {
    pid=$(lsof -ti:"$port")
    if [ -z "$pid" ]; then
        echo "wxbot is not running"
    else
        echo "wxbot is running with PID $pid"
    fi
}

pid() {
    pid=$(lsof -ti:"$port")
    if [ -z "$pid" ]; then
        echo "wxbot is not running"
    else
        echo "wxbot is running with PID $pid"
    fi
}

case "$1" in
start)
    start
    ;;
stop)
    stop
    ;;
status)
    status
    ;;
pid)
    pid
    ;;
*)
    echo "Usage: $0 {start|stop|status|pid}"
    exit 1
    ;;
esac

exit 0
