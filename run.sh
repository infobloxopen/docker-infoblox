#!/bin/bash

DRIVER_NAME=mddi

DEFAULT_CIDR=111.0.0.0/24

SOCK_EXT=sock
PLUGIN_DIR=/run/docker/plugins

DRIVER_SOCKET=${PLUGIN_DIR}/${DRIVER_NAME}.${SOCK_EXT}

rm -f $DRIVER_SOCKET

./infoblox-ipam --grid-host=192.168.124.200 --wapi-port=443 --wapi-username=cloudadmin --wapi-password=cloudadmin --global-view=global_view --local-view=local_view --cidr=$DEFAULT_CIDR --socket=$DRIVER_SOCKET
