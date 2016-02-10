infoblox-ipam
=============

Docker (libnetwork) plugin for IPAM
-----------------------------------


This program is a Docker libnetwork plugin that provides IPAM services. The libnetwork is the library provided by Docker 1.9 that allows third-party plugins for container networking. This plugin is an IPAM driver that interface with Infoblox to IP Address Management functions


Build
-----
(refer to BUILD.md)

Installation
------------
Installation is fairly simple. Make sure the directory ```/run/docker/plugins/``` exists and the UNIX socket file ```/run/docker/plugins/mddi.sock``` does not exist so that the plugin could write to ```/run/docker/plugins/mddi.sock```. 


Run Executable
--------------
The "infoblox-ipam" driver accept a number of arguments which can be listed by specifying -h:

```
ubuntu$ ./infoblox-ipam --help
Usage of ./infoblox-ipam:
  -cidr string
        Default Network CIDR if --subnet is not specified during docker network create (default "10.2.1.0/24")
  -global-view string
        Infoblox Network View for Global Address Space (default "default")
  -grid-host string
        IP of Infoblox Grid Host (default "192.168.124.200")
  -local-view string
        Infoblox Network View for Local Address Space (default "default")
  -socket string
        Unix socket for mDDI Docker (libnetwork) plugin in bridge/ipam driver (default "/run/docker/plugins/mddi.sock")
  -wapi-password string
        Infoblox WAPI Password
  -wapi-port string
        Infoblox WAPI Port. (default "443")
  -wapi-username string
        Infoblox WAPI Username
  -wapi-ver string
        Infoblox WAPI Version. (default "2.2")
```

For example,

```
./infoblox-ipam --grid-host=192.168.124.200 --wapi-username=cloudadmin --wapi-password=cloudadmin --global-view=global_view
```
The command need to be executed with root permission.

For convenience, a script called "run.sh" is provided which can be editted to specify the desired options.


Run Container
------------
Alternatively, the driver can also be run as a docker container.
For example, assuming that the name of the docker image for the driver is "infoblox-ipam", the
following command with start the driver in a container:

```
docker run -v /var/run:/var/run -v /run/docker:/run/docker ./infoblox-ipam --grid-host=192.168.124.200 --wapi-username=cloudadmin --wapi-password=cloudadmin --global-view=global_view
```

Note that the -v options are necessary to provide the container access to the specified directories on the
host file system.

For coneniences, a script called "run-container.sh" is proivded.

Usage
-----
To start using the plugin, a docker network needs to be created specifying the driver using the --ipam-driver option:
```
sudo docker network create --ipam-driver=mddi mddi-net
```
This creates a docker network called "mddi-net" which uses "mddi" as the IPAM driver and the default "bridge" driver as the network driver.

After which, Docker containers can be started attaching to the "mddi-net" network created above. For example, the following command run the "ubuntu" image:

```
sudo docker run -i -t --net=mddi-net --name=ubuntu1 ubuntu
```

When the container comes up, verify using the "ifconfig" command that IP has been successfully provisioned from Infoblox. 
