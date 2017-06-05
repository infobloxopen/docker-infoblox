FROM alpine

RUN apk update

RUN mkdir -p /run/docker/plugins

COPY bin/ipam-driver /ipam-driver
COPY launch.sh /launch.sh

ENTRYPOINT ["/launch.sh"]
