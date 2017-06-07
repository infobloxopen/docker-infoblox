# Infoblox Docker IPAM Plugin

Infoblox ipam-plugin is a Docker Engine managed plugin that interfaces with Infoblox
to provide IP Address Management services for Docker containers.

## Prerequisite

To use the driver, you need access to the Infoblox DDI product. For evaluation purposes, you can download a
virtual version of the product from the Infoblox Download Center (https://www.infoblox.com/infoblox-download-center)
Alternatively, if you are an existing Infoblox customer, you can download it from the support site.

Refer to [CONFIG.md](docs/CONFIG.md) on how to configure vNIOS.

## Installation

By default, the ipam-plugin assumes that the "Cloud Network Automation" licensed feature is activated in NIOS. Should this not be the case, refer to "Manual Configuration of Cloud Extensible Attributes" in CONFIG.md for additional
configuration required.

### 1) Create configuration file for the plugin.
create a file **`/etc/infoblox/docker-infoblox.conf`** and add the configuation parameters for the ipam-plugin. The configuration parameters are:

| Option | Type  | Description |
| ------ | ----- | ----------- |
| grid-host string   | String | IP of Infoblox Grid Host
| wapi-port  | String | Infoblox WAPI Port (default "443")
| wapi-username | String | Infoblox WAPI Username
| wapi-password | String | Infoblox WAPI Password
| wapi-version | String | Infoblox WAPI Version (default "2.0")
| ssl-verify  | String | Specifies whether (true/false) to verify server certificate. If a file path is specified, it is assumed to be a certificate file and will be used to verify server certificate.
| http-request-timeout | Integer | Infoblox WAPI request timeout in seconds (default 60)
| http-pool-connections | Integer | Infoblox WAPI request connection pool size (default 10)
| global-view  | String | Infoblox Network View for Global Address Space (default "default")
| global-network-container | String | Subnets will be allocated from this container when --subnet is not specified during network creation
| global-prefix-length | Integer | The default CIDR prefix length when allocating a global subnet (default 24)
| local-view | String | Infoblox Network View for Local Address Space (default "default")
| local-network-container | String | Subnets will be allocated from this container when --subnet is not specified during network creation
| local-prefix-length | Integer | The default CIDR prefix length when allocating a local subnet (default 24)


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


### 2) Installing plugin from the Docker Hub
```
$ docker plugin install infoblox/ipam-plugin:1.1.0 --alias infoblox

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


TODO : Need to update config.json to fix this
By default, ipam-driver uses Docker API Version 1.22 to access Docker Remote API.
The default can be overridden using the DOCKER_API_VERSION environment variable prior to running the driver. For example,

```
DOCKER_API_VERSION=1.23
export DOCKER_API_VERSION
```

## Usage

To start using the driver, a docker network needs to be created specifying the driver using the --ipam-driver option:
```
$ docker network create --ipam-driver=infoblox:latest priv-net
```
This creates a docker network called "priv-net" which uses "infoblox" as the IPAM driver and the default "bridge" driver as the network driver. A network will be automatically allocated from the list of network containers
specified during driver start up.

By default, the network will be created using the default prefix length specified during driver start up. You can override this using the --ipam-opt option. For example:

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
For example, the following command runs the "alpine" image:

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


## Build

For dependencies and build instructions, refer to ```BUILD.md```.
