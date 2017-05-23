#!/bin/bash

DOCKER_IMAGE="ipam-driver"
DRIVER_NAME="infoblox"
PLUGIN_DIR="/run/docker/plugins"
GRID_HOST="192.168.124.200"
WAPI_PORT="443"
WAPI_USERNAME=""
WAPI_PASSWORD=""
WAPI_VERSION="2.0"
SSL_VERIFY="false"
GLOBAL_VIEW="default"
GLOBAL_CONTAINER="172.18.0.0/16"
GLOBAL_PREFIX=24
LOCAL_VIEW="default"
LOCAL_CONTAINER="192.168.0.0/16"
LOCAL_PREFIX=24

DOCKER_API_VERSION="1.22"

docker run -e DOCKER_API_VERSION=${DOCKER_API_VERSION} -v /var/run:/var/run -v /run/docker:/run/docker ${DOCKER_IMAGE} --grid-host=${GRID_HOST} --wapi-port=${WAPI_PORT} --wapi-username=${WAPI_USERNAME} --wapi-password=${WAPI_PASSWORD} --wapi-version=${WAPI_VERSION} --ssl-verify=${SSL_VERIFY} --global-view=${GLOBAL_VIEW} --global-network-container=${GLOBAL_CONTAINER} --global-prefix-length=${GLOBAL_PREFIX} --local-view=${LOCAL_VIEW} --local-network-container=${LOCAL_CONTAINER} --local-prefix-length=${LOCAL_PREFIX} --plugin-dir=${PLUGIN_DIR} --driver-name=${DRIVER_NAME}
