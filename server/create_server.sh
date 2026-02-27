#!/bin/bash


if [ -z "$1" ]; then
  echo "Usage: $0 <server_name> <host_port> <container_port>"
  exit 1
fi

SERVER_NAME=$1
HOST_PORT=$2
CONTAINER_PORT=$3


if [ -z "$HOST_PORT" ] || [ -z "$CONTAINER_PORT" ]; then
  echo "Error: You must specify the host port and the container port."
  echo "Usage: $0 <server_name> <host_port> <container_port>"
  exit 1
fi

docker run -d --name "$SERVER_NAME" -p "$HOST_PORT:$CONTAINER_PORT" -p "$HOST_PORT:$CONTAINER_PORT/udp" server-vintagestory:latest || {
  echo "Error: Failed to start the Docker container."
  exit 1
}

echo "Server $SERVER_NAME started successfully in detached mode on host port $HOST_PORT."