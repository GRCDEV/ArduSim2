#!/bin/bash

if [ $(getent group "docker") ]; then
  echo -e "\033[32mDocker can already run without root-level privileges. Skipping...\033[0m"
  exit 0
fi

echo "Configuring Docker so it can run without root-level privileges"
sudo groupadd docker
sudo usermod -aG docker $USER

echo -e "\033[32mDocker group added\033[0m"
exit 0