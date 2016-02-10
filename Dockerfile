FROM ubuntu

ADD infoblox-ipam /


ENTRYPOINT ["/infoblox-ipam"]
