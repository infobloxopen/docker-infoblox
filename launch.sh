#!/bin/sh

set -e

args=""
if [ ! -z "$GRID_HOST" ]; then
   args=$args" --grid-host=$GRID_HOST"
fi
if [ ! -z "$WAPI_PORT" ]; then
   args=$args" --wapi-port=$WAPI_PORT"
fi
if [ ! -z "$WAPI_USERNAME" ]; then
   args=$args" --wapi-username=$WAPI_USERNAME"
fi
if [ ! -z "$WAPI_PASSWORD" ]; then
   args=$args" --wapi-password=$WAPI_PASSWORD"
fi
if [ ! -z "$WAPI_VERSION" ]; then
   args=$args" --wapi-version=$WAPI_VERSION"
fi
if [ ! -z "$SSL_VERIFY" ]; then
   args=$args" --ssl-verify=$SSL_VERIFY"
fi
if [ ! -z "$HTTP_REQUEST_TIMEOUT" ]; then
   args=$args" --http-request-timeout=$HTTP_REQUEST_TIMEOUT"
fi
if [ ! -z "$HTTP_POOL_CONNECTIONS" ]; then
   args=$args" --http-pool-connections=$HTTP_POOL_CONNECTIONS"
fi
if [ ! -z "$GLOBAL_VIEW" ]; then
   args=$args" --global-view=$GLOBAL_VIEW"
fi
if [ ! -z "$GLOBAL_NETWORK_CONTAINER" ]; then
   args=$args" --global-network-container=$GLOBAL_NETWORK_CONTAINER"
fi
if [ ! -z "$GLOBAL_PREFIX_LENGTH" ]; then
   args=$args" --global-prefix-length=$GLOBAL_PREFIX_LENGTH"
fi
if [ ! -z "$LOCAL_VIEW" ]; then
   args=$args" --local-view=$LOCAL_VIEW"
fi
if [ ! -z "$LOCAL_NETWORK_CONTAINER" ]; then
   args=$args" --local-network-container=$LOCAL_NETWORK_CONTAINER"
fi
if [ ! -z "$LOCAL_PREFIX_LENGTH" ]; then
   args=$args" --local-prefix-length=$LOCAL_PREFIX_LENGTH"
fi
if [ ! -z "$CONF_FILE_NAME" ]; then
   args=$args" --conf-file-name=$CONF_FILE_NAME"
fi

echo $args

echo "Starting Infoblox Docker IPAM Plugin"

/ipam-driver $args
