FROM ubuntu

ADD ipam-driver /


ENTRYPOINT ["/ipam-driver"]
