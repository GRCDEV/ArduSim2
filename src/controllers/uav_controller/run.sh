#!/bin/sh

mkdir -p /app/logs

./arducopter -S --base-port 5760 --home ${UAV_HOME_LOCATION:-"39.482594,-0.346265,0.0,0.0"} --model + --defaults copter.parm --sim-port-in 5501 --sim-port-out 5502 --rc-in-port 5503 --irlock-port 9005 > /app/logs/arducopter.log 2>&1 &

sleep 5 # wait for arducopter to boot

java -Ddebug=${DEBUG:-false} -jar ./uavController.jar config.json > /app/logs/uav_controller.log 2>&1 &

wait
