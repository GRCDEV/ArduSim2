#!/bin/bash

echo "Starting installation process of ArduSim2"

if [[ $(id -u) -ne 0 ]]; then
  echo -e "\033[31mThis script requires root privileges. Please run it using sudo.\033[0m"
  exit 1
fi

./installation_scripts/docker.sh
./installation_scripts/docker_without_root.sh
