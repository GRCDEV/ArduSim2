#!/bin/bash

# Use a more reliable check for Docker installation
if command -v docker &>/dev/null; then
  echo -e "\033[32mDocker is already installed. Skipping installation...\033[0m"
  exit 0
else
  echo "Going to install Docker"
  sudo snap install docker
  sudo addgroup --system docker
  sudo adduser $USER docker
  newgrp docker
  sudo snap disable docker
  sudo snap enable docker
  echo -e "\033[32mDocker successfully installed!\033[0m"
fi
exit 0