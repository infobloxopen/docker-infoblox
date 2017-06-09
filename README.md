# Infoblox Docker IPAM Plugin

Infoblox ipam-plugin is a Docker Engine managed plugin that interfaces with Infoblox
to provide IP Address Management services for Docker containers.

## Prerequisite

To use the driver, you need access to the Infoblox DDI product. For evaluation purposes, you can download a
virtual version of the product from the Infoblox Download Center (https://www.infoblox.com/infoblox-download-center)
Alternatively, if you are an existing Infoblox customer, you can download it from the support site.

Refer to [CONFIG.md](docs/CONFIG.md) on how to configure vNIOS.

## Configuration

Infoblox IPAM plugin can be configured either by passing arguments to the plugin while installing or adding parmeters in the configuration file and passing that file name as an argument.

### 1) Plugin configuration using command line arguments
Configuration parameters can be passed to the plugin while installing the plugin.

The parameters which can be passed are:

| Parameter | Description |
| ------ | ----------- |
| CONF_FILE_NAME | Configuration file name in /etc/infoblox directory
| DEBUG | Sets log level to debug
| DOCKER_API_VERSION | Docker API version to use (Default: 1.22)
| GRID_HOST | IP of Infoblox Grid Host
| WAPI_PORT | Infoblox WAPI Port
| WAPI_USERNAME | Infoblox WAPI Username
| WAPI_PASSWORD | Infoblox WAPI Password
| WAPI_VERSION | Infoblox WAPI Version
| SSL_VERIFY  | Specifies whether to verify server certificate or not
| HTTP_REQUEST_TIMEOUT | Infoblox WAPI request timeout in seconds
| HTTP_POOL_CONNECTIONS | Infoblox WAPI request connection pool size
| GLOBAL_VIEW  | Infoblox Network View for Global Address Space
| GLOBAL_NETWORK_CONTAINER | Subnets will be allocated from this container when --subnet <br>is not specified during network creation
| GLOBAL_PREFIX_LENGTH | The default CIDR prefix length when allocating a global subnet
| LOCAL_VIEW | Infoblox Network View for Local Address Space
| LOCAL_NETWORK_CONTAINER | Subnets will be allocated from this container when --subnet <br>is not specified during network creation
| LOCAL_PREFIX_LENGTH | The default CIDR prefix length when allocating a local subnet

### 2) Plugin configuration using configuration file
Create a file `docker-infoblox.conf` (configurable via CONF_FILE_NAME parameter) in **`/etc/infoblox/`** directory and add the configuation options in the file. Pass this configuration file name as a parameter to plugin while installing (CONF_FILE_NAME=docker-infoblox.conf)

The configuration options are:

| Option | Type  | Description |
| ------ | ----- | ----------- |
| grid-host | String | IP of Infoblox Grid Host
| wapi-port | String | Infoblox WAPI Port <br>(Default : 443)
| wapi-username | String | Infoblox WAPI Username
| wapi-password | String | Infoblox WAPI Password
| wapi-version | String | Infoblox WAPI Version <br>(Default : "2.0")
| ssl-verify  | String | Specifies whether (true/false) to verify server certificate or not. <br>If a file path is specified, it is assumed to be a certificate file <br>and will be used to verify server certificate.
| http-request-timeout | Integer | Infoblox WAPI request timeout in seconds <br>(Default : 60)
| http-pool-connections | Integer | Infoblox WAPI request connection pool size <br>(Default : 10)
| global-view  | String | Infoblox Network View for Global Address Space <br>(Default : "default")
| global-network-container | String | Subnets will be allocated from this container when --subnet <br>is not specified during network creation
| global-prefix-length | Integer | The default CIDR prefix length when allocating a global subnet <br>(Default : 24)
| local-view | String | Infoblox Network View for Local Address Space <br>(Default : "default")
| local-network-container | String | Subnets will be allocated from this container when --subnet <br>is not specified during network creation
| local-prefix-length | Integer | The default CIDR prefix length when allocating a local subnet <br>(Default : 24)


A sample plugin configuration file looks like this:
```
[grid_config]
grid-host="10.120.21.150"
wapi-port="443"
wapi-username="infoblox"
wapi-password="infoblox"
wapi-version="2.0"
ssl-verify="false"
http-request-timeout=60
http-pool-connections=10

[ipam_config]
global-view="global_view"
global-network-container="172.18.0.0/16"
global-prefix-length=24
local-view="local_view"
local-network-container="192.168.0.0/20,192.169.0.0/22"
local-prefix-length=25
```

**If some option is in passed both the ways, then configuration parameters passed as arguments overrides the configuration defined in the configuration file.**

## Installation

By default, the ipam-plugin assumes that the "Cloud Network Automation" licensed feature is activated in NIOS. Should this not be the case, refer to "Manual Configuration of Cloud Extensible Attributes" in CONFIG.md for additional configuration required.


### Installing plugin from the Docker Store
```
$ docker plugin install --alias infoblox infoblox/ipam-plugin:1.1.0 \
CONF_FILE_NAME=docker-infoblox.conf

Plugin "infoblox/ipam-plugin:1.1.0" is requesting the following privileges:
 - network: [host]
 - mount: [/etc/infoblox]
 - mount: [/var/run]
Do you grant the above permissions? [y/N]

```

The plugin requests the following priviliges:
  * access to the host network
  * mounts /etc/infoblox directory on the host as a volume to  container to read its configuration file
  * mounts /var/run directory on the host as a volume to container to access docker socket file

TO avoid the privileges request prompt pass the `--grant-all-permissions` option
```
$ docker plugin install --grant-all-permissions --alias infoblox \
infoblox/ipam-plugin:1.1.0 CONF_FILE_NAME=docker-infoblox.conf
```

To override a configuration file option
```
$ docker plugin install --grant-all-permissions --alias infoblox \
infoblox/ipam-plugin:1.1.0 CONF_FILE_NAME=docker-infoblox.conf \
LOCAL_NETWORK_CONTAINER=172.16.10.0/24
```
`LOCAL_NETWORK_CONTAINER` overides the `local-network-container` option in conf file.

Inoder to run the plugin in debug mode, pass `DEBUG` parameter

```
$ docker plugin install --grant-all-permissions --alias infoblox \
infoblox/ipam-plugin:1.1.0 CONF_FILE_NAME=docker-infoblox.conf DEBUG=true
```

Passing all the parameters from command line and not using the configuration file
```
$ docker plugin install --grant-all-permissions --alias infoblox \
infoblox/ipam-plugin:1.1.0 GRID_HOST=10.120.21.150 \
WAPI_USERNAME=admin WAPI_PASSWORD=infoblox GLOBAL_VIEW=global_view \
GLOBAL_NETWORK_CONTAINER=172.18.0.0/16 LOCAL_VIEW=local_view \
LOCAL_NETWORK_CONTAINER=192.168.0.0/20 LOCAL_PREFIX_LENGTH=25
```

## Usage

To start using the plugin, a docker network needs to be created specifying the driver using the --ipam-driver option:
```
$ docker network create --ipam-driver=infoblox:latest priv-net
```
This creates a docker network called "priv-net" which uses "infoblox" as the IPAM driver and the default "bridge" driver as the network driver. A network will be automatically allocated from the list of network containers
specified during plugin installation.

By default, the network will be created using the default prefix length specified during plugin installation. You can override this using the --ipam-opt option. For example:

```
$ docker network create --ipam-driver=infoblox:latest --ipam-opt="prefix-length=24" priv-net-2
```

Additionally, if you are deploying containers in a cluster, you can specify "network-name" using the --ipam-opt option.
This will be used as an identifier so that docker networks created on different docker hosts can share the same IP address
space. For example:

```
$ docker network create --ipam-driver=infoblox:latest --ipam-opt="network-name=blue" blue-net
```
This will allocate a network, say, 192.168.10.0/24, from the default address pool. Additionally, the network will be
tagged in Infoblox with the network name "blue". Should the same command be issued on a different host, the driver will
look for a network on Infoblox tagged with the same name, "blue", and will share the same network, 192.168.10.0/24, instead
of allocating a new one.


After the network is created, Docker containers can be started attaching to the "priv-net" network created above.
For example, the following command runs the container from "alpine" image:

```
$ docker run -it --network priv-net alpine /bin/sh

/ # ip a
1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN qlen 1
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
    inet6 ::1/128 scope host
       valid_lft forever preferred_lft forever
49: eth0@if50: <BROADCAST,MULTICAST,UP,LOWER_UP,M-DOWN> mtu 1500 qdisc noqueue state UP
    link/ether 02:42:ae:aa:e4:1c brd ff:ff:ff:ff:ff:ff
    inet 192.168.3.2/25 scope global eth0
       valid_lft forever preferred_lft forever
    inet6 fe80::42:aeff:feaa:e41c/64 scope link
       valid_lft forever preferred_lft forever
/ #

```

## Logging
Docker IPAM Plugin logs are logged in the docker daemon logs.

To check the logs find the plugin id
```
$ docker plugin inspect infoblox:latest -f '{{ .ID }}'

7e42725527bf4b631894b7e7baa7501ee65eaafec50e0966426bf327288adf95
```

Search for this ID in the docker daemon logs (default is /var/log/syslog) to find the plugin logs.


## Build

For dependencies and build instructions, refer to ```BUILD.md```.
