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
go get github.com/docker/libnetwork  # libnetwork library
go get github.com/docker/engine-api  # used to obtain docker engine id
go get github.com/infobloxopen/infoblox-go-client  # Infoblox client
```
This would install "libnetwork" as well as its dependencies.

"engine-api" is used by the infoblox-ipam driver to obtain the docker 
engine id, which is used to populate the "Tenant ID" EA.

"infoblox-go-client" is used by the ipam-driver to interact
with Infoblox.

By default, the "master" branch of "libnetwork" and "docker" will be used. To build a
release version, the corresponding branches need to be checked out:

```
libnetwork  release-0.5
docker      release/v1.9
```
or
```
libnetwork  release/v0.6
docker      release/v1.10
```

Obviously, the driver need to be rebuilt after a different version of the above
is checked out.

Build Executable
----------------
A Makefile is provided for automate the build process. To build the ipam-driver, just type
```make``` in the "docker-infoblox" source directory. This creates an executable called "ipam-driver".

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
