FROM ubuntu

ADD bin/ipam-driver /


ENTRYPOINT ["/ipam-driver"]
