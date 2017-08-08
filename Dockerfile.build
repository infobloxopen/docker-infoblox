FROM golang:1.7-alpine

RUN apk --update add --no-cache --virtual .build-deps \
    gcc libc-dev linux-headers

COPY . /go/src/github.com/infobloxopen/docker-infoblox
WORKDIR /go/src/github.com/infobloxopen/docker-infoblox

RUN go build -o bin/ipam-plugin ./driver \
    && go build -o bin/create-ea-defs ./ea-defs
