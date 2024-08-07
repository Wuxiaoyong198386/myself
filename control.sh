#!/bin/bash

workspace=$(cd $(dirname $0); pwd)
cd $workspace

mkdir -p logs

app=myselfgo 
pid_file=logs/app.pid
log_file=logs/app.log
sleep_time=2

conf=configs/cfg.production.yaml 

echo "$conf will be used"

function check_pid() {
    if [ -f $pid_file ]; then
        pid=`cat $pid_file`
        if [ -n $pid ]; then
            running=`ps -p $pid | grep -v "PID TTY" | wc -l`
            return $running
        fi
    fi
    return 0
}

function start() {
    check_pid
    running=$?
    if [ $running -gt 0 ]; then
        echo -n "$app is running already, pid="
        cat $pid_file
        return 1
    fi
    echo "use the config file: $conf"
    nohup $workspace/$app -c $conf > $log_file 2>&1 &
    echo $! > $pid_file

    echo "wait $sleep_time seconds to check process..."
    sleep $sleep_time
    check_pid
    running=$?
    if [ $running -le 0 ]; then
        echo -e "failed to start $app, see $log_file or info.log"
    else
        echo "succeed to start $app, pid=$!"
    fi
}

function stop() {
    check_pid
    running=$?
    if [ $running -le 0 ]; then
        echo "$app is already stopped"
        return
    fi

    pid=`cat $pid_file`
    if [ $pid -gt 0 ]; then
        kill -9 $pid
        echo "$app stopped..."
        return
    fi

    pid=`ps -ef | grep myselfgo | grep -v grep | awk '{print $2}'`
    if [ $pid -gt 0 ]; then
        kill -9 $pid
    fi

    echo "$app stopped..."
}

function restart() {
    stop
    sleep 1
    echo "try to restart $app..."
    start
}

function status() {
    check_pid
    running=$?
    if [ $running -gt 0 ]; then
        echo -n "$app is running, pid="
        cat $pid_file
    else
        echo "$app is stopped"
    fi
}


function help() {
    echo "$0 start|stop|restart|status"
}

if [ "$1" == "" ]; then
    help
elif [ "$1" == "stop" ]; then
    stop
elif [ "$1" == "start" ]; then
    start
elif [ "$1" == "stop" ]; then
    stop
elif [ "$1" == "restart" ]; then
    restart
elif [ "$1" == "status" ]; then
    status
else
    help
fi
