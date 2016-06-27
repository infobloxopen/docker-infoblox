Build ipam-driver
=================

Prerequisite
------------
1. golang development environment is installed (https://golang.org/doc/install)


Install Dependency
------------------
The driver primarily depends on libnetwork and the infoblox-go-client . They can be installed using the
following commands:

```
go get github.com/docker/libnetwork  # libnetwork library (This also pulls down docker engine)
go get github.com/docker/engine-api  # engine-api library (NOTE: run "make deps" to get dependencies)
go get github.com/infobloxopen/infoblox-go-client  # Infoblox client
```
```engine-api``` is used by the infoblox-ipam driver to obtain the docker
engine id, which is used to populate the "Tenant ID" EA.

```infoblox-go-client``` is used by the ipam-driver to interact
with Infoblox.

By default, the ```master``` branch of ```libnetwork``` and ```docker``` will be used. To build
with release versions, the corresponding release tags need to be checked out. Minimum requirement
for the ipam-driver is:

```
libnetwork  0.5
Docker      1.9.0
```
It has also been tested using master. For libnetwork release information, refer to:
https://github.com/docker/libnetwork/wiki

Obviously, the driver need to be rebuilt after a different version of the above
is checked out.

Build Executable
----------------
A Makefile is provided for automate the build process. To build the ipam-driver, just type
```make``` in the ```docker-infoblox``` source directory. This creates an executable called ```ipam-driver```.

Build Container Image
---------------------
To build container image using the Dockerfile in the "docker-infoblox" directory:

```
make image
```

Push Container Image to Docker Hub
----------------------------------
The Makefile also includes a build target to push the "ipam-driver" container image to your Docker Hub. To do that, you need
to first setup the following environment variable:

```
export DOCKERHUB_ID="your-docker-hub-id"

```
You can then use the following command to push the "ipam-driver" image to your Docker Hub:

```
make push
```
