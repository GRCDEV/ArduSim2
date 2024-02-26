#!/bin/bash

echo "Starting installation process of ArduSim2"

if [[ $(id -u) -ne 0 ]]; then
  ./installation_scripts/docker.sh

fi

echo -e "\033[31mPlease run this script without using sudo.\033[0m"
exit 1
