#!/bin/bash

DRIVER_NAME="mddi"
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



./ipam-driver --grid-host=${GRID_HOST} --wapi-port=${WAPI_PORT} --wapi-username=${WAPI_USERNAME} --wapi-password=${WAPI_PASSWORD} --wapi-version=${WAPI_VERSION} --ssl-verify=${SSL_VERIFY} --global-view=${GLOBAL_VIEW} --global-network-container=${GLOBAL_CONTAINER} --global-prefix-length=${GLOBAL_PREFIX} --local-view=${LOCAL_VIEW} --local-network-container=${LOCAL_CONTAINER} --local-prefix-length=${LOCAL_PREFIX} --plugin-dir=${PLUGIN_DIR} --driver-name=${DRIVER_NAME}
