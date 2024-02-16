#!/bin/bash

# Use a more reliable check for Docker installation
if command -v docker &>/dev/null; then
  echo -e "\033[32mDocker is already installed. Skipping installation...\033[0m"
  exit 0
else
  echo "Going to install Docker"
  # Add Docker's official GPG key:
  apt-get update
  apt-get install ca-certificates curl
  install -m 0755 -d /etc/apt/keyrings
  curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
  chmod a+r /etc/apt/keyrings/docker.asc

  # Add the repository to Apt sources:
  echo \
    "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu \
    $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
    tee /etc/apt/sources.list.d/docker.list > /dev/null
  apt-get update

  apt-get install docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

  # Verify installation
  if ! docker info &>/dev/null; then
    echo -e "\033[31mFailed to install Docker Engine. Please consult the official documentation for troubleshooting.\033[0m"
    exit 1
  fi

  echo -e "\033[32mDocker successfully installed!\033[0m"
fi
exit 0