Build infoblox-ipam
===================

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

"infoblox-go-client" is used by the infoblox-ipam driver to interact
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
To build the infoblox-ipam driver, use the following command in the "docker-infoblox" source directory:

```
go build infoblox-ipam.go
```
This creates an executable called "infoblox-ipam".

Build Container Image
---------------------
You have to first create the binary executable.
To build container image using the Dockerfile in the "docker-infoblox" directory:

```
docker build -t ipam-driver .
```

This build a docker image called "ipam-driver" according to the Dockerfile in the current directory.
