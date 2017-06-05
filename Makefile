ifeq ($(DOCKERHUB_ID),)
    PLUGIN_NAME=ipam-plugin
	TOOLS_IMAGE=ipam-tools
else
    PLUGIN_NAME=${DOCKERHUB_ID}/ipam-plugin
    TOOLS_IMAGE=${DOCKERHUB_ID}/ipam-tools
endif
RELEASE=1.1.0

.PHONY: clean-plugin
clean-plugin:
	rm -rf ./plugin ./bin
	docker plugin disable ${PLUGIN_NAME}:${RELEASE} || true
	docker plugin rm ${PLUGIN_NAME}:${RELEASE} || true
	docker rm -vf tmp || true
	docker rmi ipam-build-image || true
	docker rmi ${PLUGIN_NAME}:rootfs || true

.PHONY: build-plugin-image
build-plugin-image:
	docker build -t ipam-build-image -f Dockerfile.build .
	docker create --name build-container ipam-build-image
	docker cp build-container:/go/src/github.com/infobloxopen/docker-infoblox/bin .
	docker rm -vf build-container
	docker rmi ipam-build-image
	docker build -t ${PLUGIN_NAME}:rootfs .

.PHONY: build-plugin
build-plugin: build-plugin-image
	mkdir -p ./plugin/rootfs
	docker create --name tmp ${PLUGIN_NAME}:rootfs
	docker export tmp | tar -x -C ./plugin/rootfs
	cp config.json ./plugin/
	docker rm -vf tmp

.PHONY: create-plugin
create-plugin: 
	docker plugin create ${PLUGIN_NAME}:${RELEASE} ./plugin

.PHONY: enable-plugin
enable-plugin:
	docker plugin enable ${PLUGIN_NAME}:${RELEASE}

.PHONY: push-plugin
push-plugin:  clean-plugin build-plugin-image build-plugin create-plugin
	docker plugin push ${PLUGIN_NAME}:${RELEASE}

clean-tools-image:
	echo "WIP"

build-tools-image:
	echo "WIP"

push-tools-image: 
	echo "WIP"
