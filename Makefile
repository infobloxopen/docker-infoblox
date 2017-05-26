PLUGIN_NAME=ishant8/ipam-plugin
TOOLS_IMAGE=ishant8/ipam-tools
RELEASE=1.1.0

#all: clean build-image #build-plugin create-plugin

clean-plugin:
	rm -rf ./plugin ./bin
	docker plugin disable ${PLUGIN_NAME}:${RELEASE} || true
	docker plugin rm ${PLUGIN_NAME}:${RELEASE} || true
	docker rm -vf tmp || true
	docker rmi ipam-build-image || true
	docker rmi ${PLUGIN_NAME}:rootfs || true

build-plugin-image:
	docker build -t ipam-build-image -f Dockerfile.build .
	docker create --name build-container ipam-build-image
	docker cp build-container:/go/src/github.com/infobloxopen/docker-infoblox/bin .
	docker rm -vf build-container
	docker rmi ipam-build-image
	docker build -t ${PLUGIN_NAME}:rootfs .

build-plugin: build-plugin-image
	mkdir -p ./plugin/rootfs
	docker create --name tmp ${PLUGIN_NAME}:rootfs
	docker export tmp | tar -x -C ./plugin/rootfs
	cp config.json ./plugin/
	docker rm -vf tmp

create-plugin:
	docker plugin create ${PLUGIN_NAME}:${RELEASE} ./plugin

enable-plugin:
	docker plugin enable ${PLUGIN_NAME}:${RELEASE}

push-plugin:  clean build-plugin-image build-plugin create-plugin
	docker plugin push ${PLUGIN_NAME}:${RELEASE}


clean-tools-image:
	echo "WIP"

build-tools-image:
	echo "WIP"

push-tools-image: 
	echo "WIP"
