infoblox-ipam
=============

Docker (libnetwork) driver for IPAM
-----------------------------------

infoblox-ipam is a Docker libnetwork driver that interfaces with Infoblox to provide IP Address Management
services. libnetwork is the library provided by Docker that allows third-party drivers for container
networking.


Prerequisite
------------
To use the driver, you need access to the Infoblox DDI product. For evaluation purposes, you can download a
virtual version of the product from the Infoblox Download Center (https://www.infoblox.com/infoblox-download-center)
Alternatively, if you are an existing Infoblox customer, you can download it from the support site.


Build
-----
(refer to BUILD.md)

Installation
------------
Installation is fairly simple. Make sure the directory ```/run/docker/plugins/``` exists and the
UNIX socket file ```/run/docker/plugins/mddi.sock``` does not exist so that the driver could write to
```/run/docker/plugins/mddi.sock``` .


Run Executable
--------------
The "infoblox-ipam" driver accept a number of arguments which can be listed by specifying -h:

```
ubuntu$ ./infoblox-ipam --help
Usage of ./infoblox-ipam:
  -default-cidr string
    			Default Network CIDR if --subnet is not specified during docker network create (default "10.2.1.0/24")
  -driver-name string
    		   Name of Infoblox IPAM driver (default "mddi")
  -global-view string
    		   Infoblox Network View for Global Address Space (default "default")
  -grid-host string
    		 IP of Infoblox Grid Host (default "192.168.124.200")
  -local-view string
    		  Infoblox Network View for Local Address Space (default "default")
  -plugin-dir string
    		  Docker plugin directory where driver socket is created (default "/run/docker/plugins")
  -wapi-password string
    			 Infoblox WAPI Password
  -wapi-port string
    		 Infoblox WAPI Port. (default "443")
  -wapi-username string
    			 Infoblox WAPI Username
  -wapi-version string
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

A pre-built docker image can be pulled from Docker Hub using the following command:
```
docker pull infoblox/ipam-driver
```

After successfully pulling the image, you use the ```docker run``` command to run the driver. For exampe:
```
docker run -v /var/run:/var/run -v /run/docker:/run/docker infoblox/ipam-driver --grid-host=192.168.124.200 --wapi-username=cloudadmin --wapi-password=cloudadmin --global-view=global_view
```

Note that the -v options are necessary to provide the container access to the specified directories on the
host file system.

For convenience, a script called "run-container.sh" is provided.

Usage
-----
To start using the dirver, a docker network needs to be created specifying the driver using the --ipam-driver option:
```
sudo docker network create --ipam-driver=mddi mddi-net
```
This creates a docker network called "mddi-net" which uses "mddi" as the IPAM driver and the default "bridge"
driver as the network driver.

After which, Docker containers can be started attaching to the "mddi-net" network created above. For example,
the following command run the "ubuntu" image:

```
sudo docker run -i -t --net=mddi-net --name=ubuntu1 ubuntu
```

When the container comes up, verify using the "ifconfig" command that IP has been successfully provisioned
from Infoblox.
