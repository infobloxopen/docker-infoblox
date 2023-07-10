# Infoblox Docker IPAM Plugin

Infoblox docker-ipam-plugin is a Docker Engine managed plugin that interfaces with Infoblox
to provide IP Address Management services for Docker containers.

Infoblox docker-ipam-plugin is Docker Certified IPAM Plugin and is available at [Docker hub](https://hub.docker.com/r/infoblox/ipam-plugin)

## Limitations

- Currently works against the Grid Master not the CP Member.

## Prerequisite

To use the driver, you need access to the Infoblox DDI product. For evaluation purposes, you can download a
virtual version of the product from the Infoblox Download Center (https://www.infoblox.com/infoblox-download-center)
Alternatively, if you are an existing Infoblox customer, you can download it from the support site.

Refer to [CONFIG.md](docs/CONFIG.md) on how to configure vNIOS.

## Configuration

Infoblox IPAM plugin can be configured in following ways:
1. Adding parameters in a configuration file and setting `CONF_FILE_NAME` environment variable to the file name.
2. Setting all the plugin environment variables while installing the plugin.

The configuration options are:

| Environment Variable     | Configuration File Option | Type    | Description |
| ------                   |      ------               | -----   | ----------- |
| CONF_FILE_NAME           | -                         | -       | Configuration file name in /etc/infoblox directory
| DEBUG                    | -                         | -       | Sets log level to debug
| DOCKER_API_VERSION       | -                         | -       | Docker API version to use <br>(Default : 1.22)
| GRID_HOST                | grid-host                 | String  | IP of Infoblox Grid Host
| WAPI_PORT                | wapi-port                 | String  | Infoblox WAPI Port <br>(Default : "443")
| WAPI_USERNAME            | wapi-username             | String  | Infoblox WAPI Username
| WAPI_PASSWORD            | wapi-password             | String  | Infoblox WAPI Password
| WAPI_VERSION             | wapi-version              | String  | Infoblox WAPI Version <br>(Default : "2.0")
| SSL_VERIFY               | ssl-verify                | String  | Specifies whether (true/false) to verify server <br>certificate or not. If a file path is specified, it is <br>assumed to be a certificate file and will be used <br>to verify server certificate.
| HTTP_REQUEST_TIMEOUT     | http-request-timeout      | Integer | Infoblox WAPI request timeout in seconds <br>(Default : 60)
| HTTP_POOL_CONNECTIONS    | http-pool-connections     | Integer | Infoblox WAPI request connection pool size <br>(Default : 10)
| GLOBAL_VIEW              | global-view               | String  | Infoblox Network View for Global Address Space <br>(Default : "default")
| GLOBAL_NETWORK_CONTAINER | global-network-container  | String  | Subnets will be allocated from this container when <br>--subnet is not specified during network creation
| GLOBAL_PREFIX_LENGTH     | global-prefix-length      | Integer | The default CIDR prefix length when allocating a <br>global subnet <br>(Default : 24)
| LOCAL_VIEW               | local-view                | String  | Infoblox Network View for Local Address Space <br>(Default : "default")
| LOCAL_NETWORK_CONTAINER  | local-network-container   | String  | Subnets will be allocated from this container when <br>--subnet is not specified during network creation
| LOCAL_PREFIX_LENGTH      | local-prefix-length       | Integer | The default CIDR prefix length when allocating a <br>local subnet <br>(Default : 24)


**If some option is passed in both the ways, then configuration passed as plugin environment variable overrides the configuration defined in the configuration file.**

## Installation

By default, the docker-ipam-plugin assumes that the "Cloud Network Automation" licensed feature is activated in NIOS. Should this not be the case, refer to "Manual Configuration of Cloud Extensible Attributes" in CONFIG.md for additional configuration required.

Plugin is installed by pulling the infoblox/ipam-plugin:1.1.0 from Docker Hub and setting its environment variables.

In Docker swarm mode the plugin needs to be installed on all the nodes.

Plugin can be configured with either of the following ways:
1. Using the plugin configuration file
2. Using the plugin environment variables

### 1) Installing and configuring the plugin with the configuration file
Create a file `docker-infoblox.conf` (configurable via CONF_FILE_NAME parameter) in **`/etc/infoblox/`** directory and add the configuration options in the file.

A sample plugin configuration file looks like this:
```
[grid-config]
grid-host="10.120.21.150"
wapi-port="443"
wapi-username="infoblox"
wapi-password="infoblox"
wapi-version="2.0"
ssl-verify="false"
http-request-timeout=60
http-pool-connections=10

[ipam-config]
global-view="global_view"
global-network-container="172.18.0.0/16"
global-prefix-length=24
local-view="local_view"
local-network-container="192.168.0.0/20,192.169.0.0/22"
local-prefix-length=25
```

Set the CONF_FILE_NAME variable to this file name while installing the plugin.

```
$ docker plugin install --alias infoblox infoblox/ipam-plugin:1.1.0 \
CONF_FILE_NAME=docker-infoblox.conf

Plugin "infoblox/ipam-plugin:1.1.0" is requesting the following privileges:
 - network: [host]
 - mount: [/etc/infoblox]
 - mount: [/var/run]
Do you grant the above permissions? [y/N]

```

The plugin requests the following privileges:
  * access to the host network
  * mounts /etc/infoblox directory on the host as a volume to  container to read its configuration file
  * mounts /var/run directory on the host as a volume to container to access docker socket file

To avoid the privileges request prompt pass the `--grant-all-permissions` option

```
$ docker plugin install --grant-all-permissions --alias infoblox \
infoblox/ipam-plugin:1.1.0 CONF_FILE_NAME=docker-infoblox.conf
```

### 2) Installing and configuring the plugin by setting its environment variables

* Plugin can be configured without configuration file by setting all the plugin environment variables while installing the plugin.
```
$ docker plugin install --grant-all-permissions --alias infoblox \
infoblox/ipam-plugin:1.1.0 GRID_HOST=10.120.21.150 \
WAPI_USERNAME=admin WAPI_PASSWORD=infoblox GLOBAL_VIEW=global_view \
GLOBAL_NETWORK_CONTAINER=172.18.0.0/16 LOCAL_VIEW=local_view \
LOCAL_NETWORK_CONTAINER=192.168.0.0/20 LOCAL_PREFIX_LENGTH=25
```

* To override a configuration file option with the plugin environment variable
```
$ docker plugin install --grant-all-permissions --alias infoblox \
infoblox/ipam-plugin:1.1.0 CONF_FILE_NAME=docker-infoblox.conf \
LOCAL_NETWORK_CONTAINER=172.16.10.0/24
```
Here `LOCAL_NETWORK_CONTAINER` overrides the `local-network-container` option in conf file.

* Inorder to set the plugin log level as debug, set the `DEBUG` variable
```
$ docker plugin install --grant-all-permissions --alias infoblox \
infoblox/ipam-plugin:1.1.0 CONF_FILE_NAME=docker-infoblox.conf DEBUG=true
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

Inoder to use an existing NIOS network as a Docker network, add an EA "Network Name" to that network and then pass the EA value
as ipam-opt while creating the Docker network.

For example if there is a network "172.56.121.0/24" with gateway "172.56.121.1" already existing, then add an EA "Network Name"
for the network with value as "SomeNetwork" in NIOS. Then create the Docker network:
```
docker network create --ipam-driver=infoblox:latest --subnet 172.56.121.0/24 --gateway 172.56.121.1 --ipam-opt="network-name=SomeNetwork" SomeNetwork
```

### Using plugin in swarm mode with swarm scope networks

Currently the plugin supports only the MACVLAN network driver with swarm scope.
Before performing the following steps, the plugin needs to be installed on all the swarm nodes.

1) Create a config only network on all the nodes
```
sudo docker network create --config-only -o parent=eth1 --ipam-driver=infoblox:latest --ipam-opt="network-name=docker-macvlan" mv-config
```

2) Create MACVLAN network with swarm scope on the swarm manager node
```
sudo docker network create -d macvlan --scope=swarm --config-from mv-config --attachable swarm-macvlan
```

3) Run service with the swarm network
```
sudo docker service create --replicas 3 --network swarm-macvlan --name swarm-macvlan-test chrch/docker-pets:1.0
```

4) Verify the IPs allocated to the containers of the service
```
master $ sudo docker network inspect --verbose --format '{{json .Services}}' swarm-macvlan | python -m json.tool
{
    "swarm-macvlan-test": {
        "LocalLBIndex": 262,
        "Ports": [],
        "Tasks": [
            {
                "EndpointID": "18c19d799d117a8afa20fcf5bb51c8927a4921abe121913e0b080284da459080",
                "EndpointIP": "192.168.9.196",
                "Info": null,
                "Name": "swarm-macvlan-test.2.xqlym8mbvpo9qw6g9b6bwbm5j"
            },
            {
                "EndpointID": "ac6d449822bac91916ed2aa9ab6dd02001cd324ec902fc14358cdebad3f9ca5e",
                "EndpointIP": "192.168.9.198",
                "Info": null,
                "Name": "swarm-macvlan-test.3.kopddbmdmavmmd1mbmo4s9i4v"
            },
            {
                "EndpointID": "c399388cbc2ca94c5a64bb0c7fef6d398d58baf01c6dec3ede26b58f976c0eb7",
                "EndpointIP": "192.168.9.197",
                "Info": null,
                "Name": "swarm-macvlan-test.1.cfjzh0hnnroguzire8xpskjrb"
            }
        ],
        "VIP": "<nil>"
    }
}
```

4) Verify the connectivity between the containers by running these commands on the swarm manager node
```
master $ sudo docker service ls
ID                  NAME                 MODE                REPLICAS            IMAGE                   PORTS
huj54859rbgr        swarm-macvlan-test   replicated          3/3                 chrch/docker-pets:1.0   
```
```
master $ sudo docker service ps huj54859rbgr
ID                  NAME                   IMAGE                   NODE                DESIRED STATE       CURRENT STATE        ERROR               PORTS
cfjzh0hnnrog        swarm-macvlan-test.1   chrch/docker-pets:1.0   worker-1            Running             Running 7 days ago                       
xqlym8mbvpo9        swarm-macvlan-test.2   chrch/docker-pets:1.0   worker-2            Running             Running 7 days ago                       
kopddbmdmavm        swarm-macvlan-test.3   chrch/docker-pets:1.0   master              Running             Running 7 days ago                       
```
```
master $ sudo docker exec -it swarm-macvlan-test.3.kopddbmdmavmmd1mbmo4s9i4v sh
/app # ip a
1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN qlen 1
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
40: eth0@if3: <BROADCAST,MULTICAST,UP,LOWER_UP,M-DOWN> mtu 1500 qdisc noqueue state UNKNOWN
    link/ether 02:42:05:4d:d0:a1 brd ff:ff:ff:ff:ff:ff
    inet 192.168.9.198/26 scope global eth0
       valid_lft forever preferred_lft forever
/app #
/app # ping swarm-macvlan-test.1.cfjzh0hnnroguzire8xpskjrb
PING swarm-macvlan-test.1.cfjzh0hnnroguzire8xpskjrb (192.168.9.197): 56 data bytes
64 bytes from 192.168.9.197: seq=0 ttl=64 time=0.868 ms
64 bytes from 192.168.9.197: seq=1 ttl=64 time=0.728 ms
64 bytes from 192.168.9.197: seq=2 ttl=64 time=0.759 ms
64 bytes from 192.168.9.197: seq=3 ttl=64 time=0.783 ms

--- swarm-macvlan-test.1.cfjzh0hnnroguzire8xpskjrb ping statistics ---
4 packets transmitted, 4 packets received, 0% packet loss
round-trip min/avg/max = 0.728/0.784/0.868 ms
/app #
/app # ping swarm-macvlan-test.2.xqlym8mbvpo9qw6g9b6bwbm5j
PING swarm-macvlan-test.2.xqlym8mbvpo9qw6g9b6bwbm5j (192.168.9.196): 56 data bytes
64 bytes from 192.168.9.196: seq=0 ttl=64 time=1.129 ms
64 bytes from 192.168.9.196: seq=1 ttl=64 time=0.427 ms
64 bytes from 192.168.9.196: seq=2 ttl=64 time=0.636 ms
64 bytes from 192.168.9.196: seq=3 ttl=64 time=0.744 ms
64 bytes from 192.168.9.196: seq=4 ttl=64 time=0.655 ms

--- swarm-macvlan-test.2.xqlym8mbvpo9qw6g9b6bwbm5j ping statistics ---
5 packets transmitted, 5 packets received, 0% packet loss
round-trip min/avg/max = 0.427/0.718/1.129 ms
/app #
/app # curl swarm-macvlan-test.2.xqlym8mbvpo9qw6g9b6bwbm5j:5000
<html>
  <head>  
    <link rel='stylesheet' type='text/css' href="../static/style.css">
    <title>Docker PaaS</title>
  </head> ...
```

## Logging
Docker IPAM Plugin logs are logged in the docker daemon logs.

To check the logs find the plugin id
```
$ docker plugin inspect infoblox:latest -f '{{ .ID }}'

980b5a3befebc1a64d6d788b4c1e78676ed3e2632e78b9b38a370a9e71d73ee3
```

Search for this ID in the docker daemon logs (default is /var/log/syslog) to find the plugin logs.

```
$ grep 'plugin=980b5a3befebc1a64d6d788b4c1e78676ed3e2632e78b9b38a370a9e71d73ee3' /var/log/syslog

Jun  8 16:08:40 ishant dockerd[23649]: time="2017-06-08T16:08:40+05:30" level=info msg="time=\"2017-06-08T10:38:40Z\" level=info msg=\"Loading IPAM Configuration from the file\" " plugin=980b5a3befebc1a64d6d788b4c1e78676ed3e2632e78b9b38a370a9e71d73ee3
Jun  8 16:08:40 ishant dockerd[23649]: time="2017-06-08T16:08:40+05:30" level=info msg="time=\"2017-06-08T10:38:40Z\" level=info msg=\"Found Configuration file /etc/infoblox/docker-infoblox.conf" plugin=980b5a3befebc1a64d6d788b4c1e78676ed3e2632e78b9b38a370a9e71d73ee3
Jun  8 16:08:40 ishant dockerd[23649]: time="2017-06-08T16:08:40+05:30" level=info msg="\" " plugin=980b5a3befebc1a64d6d788b4c1e78676ed3e2632e78b9b38a370a9e71d73ee3
Jun  8 16:08:40 ishant dockerd[23649]: time="2017-06-08T16:08:40+05:30" level=info msg="time=\"2017-06-08T10:38:40Z\" level=info msg=\"Loading IPAM Configuration from the environment variables\" " plugin=980b5a3befebc1a64d6d788b4c1e78676ed3e2632e78b9b38a370a9e71d73ee3
Jun  8 16:08:40 ishant dockerd[23649]: time="2017-06-08T16:08:40+05:30" level=info msg="time=\"2017-06-08T10:38:40Z\" level=info msg=\"Configuration successfully loaded" plugin=980b5a3befebc1a64d6d788b4c1e78676ed3e2632e78b9b38a370a9e71d73ee3
Jun  8 16:08:40 ishant dockerd[23649]: time="2017-06-08T16:08:40+05:30" level=info msg="\" " plugin=980b5a3befebc1a64d6d788b4c1e78676ed3e2632e78b9b38a370a9e71d73ee3
Jun  8 16:08:40 ishant dockerd[23649]: time="2017-06-08T16:08:40+05:30" level=info msg="time=\"2017-06-08T10:38:40Z\" level=info msg=\"Socket File: '/run/docker/plugins/infoblox.sock'\" " plugin=980b5a3befebc1a64d6d788b4c1e78676ed3e2632e78b9b38a370a9e71d73ee3
Jun  8 16:08:40 ishant dockerd[23649]: time="2017-06-08T16:08:40+05:30" level=info msg="time=\"2017-06-08T10:38:40Z\" level=info msg=\"Docker id is 'DVZR:HNJZ:42OG:XTRO:YOHD:VDYA:EBKK:UDO7:ILEA:JF7R:KYGG:QCIO'" plugin=980b5a3befebc1a64d6d788b4c1e78676ed3e2632e78b9b38a370a9e71d73ee3

```
## Build

For dependencies and build instructions, refer to ```BUILD.md```.
