ifeq ($(DOCKERHUB_ID),)
    PLUGIN_NAME=docker-ipam-plugin
	TOOLS_IMAGE_NAME=docker-ipam-tools
else
    PLUGIN_NAME=${DOCKERHUB_ID}/docker-ipam-plugin
    TOOLS_IMAGE_NAME=${DOCKERHUB_ID}/docker-ipam-tools
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

.PHONY: build-binary
build-binary:
	docker build -t ipam-build-image -f Dockerfile.build .
	docker create --name build-container ipam-build-image
	docker cp build-container:/go/src/github.com/infobloxopen/docker-infoblox/bin .
	docker rm -vf build-container
	docker rmi ipam-build-image

.PHONY: build-plugin
build-plugin:
	docker build -t ${PLUGIN_NAME}:rootfs .
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
push-plugin:  clean-plugin build-binary build-plugin create-plugin
	docker plugin push ${PLUGIN_NAME}:${RELEASE}

.PHONY: clean-tools-image
clean-tools-image:
	docker rmi ${TOOLS_IMAGE_NAME}:${RELEASE} || true
	docker rmi ${TOOLS_IMAGE_NAME} || true

.PHONY: build-tools-image
build-tools-image:
	docker build -t ${TOOLS_IMAGE_NAME} -f Dockerfile.tools .

.PHONY: push-tools-image
push-tools-image: clean-tools-image build-binary build-tools-image
	docker tag ${TOOLS_IMAGE_NAME} ${TOOLS_IMAGE_NAME}:${RELEASE}
	docker push ${TOOLS_IMAGE_NAME}:${RELEASE}
