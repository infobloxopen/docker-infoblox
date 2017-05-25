FROM alpine

RUN apk update

RUN mkdir -p /run/docker/plugins

COPY bin/ipam-driver ipam-driver

ENTRYPOINT ["/ipam-driver", "--conf-file", "/etc/infoblox/docker-infoblox.conf"]
