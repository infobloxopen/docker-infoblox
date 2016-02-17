#!/bin/bash

DOCKER_IMAGE="ipam-driver"
DRIVER_NAME="mddi"
PLUGIN_DIR="/run/docker/plugins"
GRID_HOST="192.168.124.200"
WAPI_PORT="443"
WAPI_USERNAME=""
WAPI_PASSWORD=""
WAPI_VERSION="2.0"
GLOBAL_VIEW="default"
LOCAL_VIEW="default"
DEFAULT_CIDR="10.2.1.0/24"


docker run  -v /var/run:/var/run -v /run/docker:/run/docker ${DOCKER_IMAGE} --grid-host=${GRID_HOST} --wapi-port=${WAPI_PORT} --wapi-username=${WAPI_USERNAME} --wapi-password=${WAPI_PASSWORD} --wapi-version=${WAPI_VERSION} --global-view=${GLOBAL_VIEW} --local-view=${LOCAL_VIEW} --default-cidr=${DEFAULT_CIDR} --plugin-dir=${PLUGIN_DIR} --driver-name=${DRIVER_NAME}
