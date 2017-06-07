#!/bin/sh

set -e

args=""
if [ ! -z "$WAPI_USERNAME" ]; then
   args=$args" --wapi-username=$WAPI_USERNAME"
fi
if [ ! -z "$WAPI_PASSWORD" ]; then
   args=$args" --wapi-password=$"
fi
if [ ! -z "$WAPI_VERSION" ]; then
   args=$args" --wapi-version=$WAPI_VERSION"
fi
if [ ! -z "$SSL_VERIFY" ]; then
   args=$args" --ssl-verify=$"
fi
if [ ! -z "$HTTP_REQUEST_TIMEOUT" ]; then
   args=$args" --http-request-timeout=$"
fi
if [ ! -z "$HTTP_POOL_CONNECTIONS" ]; then
   args=$args" --http-pool-connections=$"
fi
if [ ! -z "$GLOBAL_VIEW" ]; then
   args=$args" --global-view=$"
fi
if [ ! -z "$GLOBAL_NETWORK_CONTAINER" ]; then
   args=$args" --global-network-container=$"
fi
if [ ! -z "$GLOBAL_PREFIX_LENGTH" ]; then
   args=$args" --global-prefix-length=$"
fi
if [ ! -z "$LOCAL_VIEW" ]; then
   args=$args" --local-view=$"
fi
if [ ! -z "$" ]; then
   args=$args" --local-network-container=$"
fi
if [ ! -z "$LOCAL_PREFIX_LENGTH" ]; then
   args=$args" --local-prefix-length=$"
fi
if [ ! -z "$CONF_FILE_NAME" ]; then
   args=$args" --conf-file-name=$"
fi

echo $args

echo "Starting Infoblox Docker IPAM Plugin"

/ipam-driver $args
