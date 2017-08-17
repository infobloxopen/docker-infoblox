FROM alpine:3.5

RUN mkdir -p /run/docker/plugins

COPY bin/ipam-plugin /ipam-plugin

ENTRYPOINT ["/ipam-plugin"]
