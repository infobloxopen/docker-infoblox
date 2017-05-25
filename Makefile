BINARY_NAME=ipam-driver-v2
IMAGE_NAME=ipam-driver-v2
LOCAL_IMAGE=$(IMAGE_NAME)
DEV_IMAGE=$(DOCKERHUB_ID)/$(IMAGE_NAME)  # Requires DOCKERHUB_ID environment variable
RELEASE_IMAGE=infoblox/$(IMAGE_NAME)

# Clean everything
clean-all: clean clean-images
CREATE_EA_DEFS=create_ea_defs

PLUGIN_NAME=ishant8/infoblox
# Build binary - this is the default target
build: $(BINARY_NAME) $(CREATE_EA_DEFS)

# Build binary and docker image
all: build image

build-image:
	docker build -t buildimage -f Dockerfile.build .
	docker create --name build-container buildimage
	docker cp build-container:/go/src/github.com/infobloxopen/docker-infoblox/bin .
	docker rm -vf build-container
	docker rmi buildimage
	docker build -t $(IMAGE_NAME):rootfs .

build-plugin:
	mkdir -p ./plugin/rootfs
	docker create --name build-plugin-container $(IMAGE_NAME):rootfs
	docker export build-plugin-container | tar -x -C ./plugin/rootfs
	cp config.json ./plugin/
	docker rm -vf build-plugin-container

# Build local docker image
image: build
	docker build -t $(LOCAL_IMAGE) .

# Push image to user's docker hub. NOTE: requires DOCKERHUB_ID environment variable
push: image
	docker tag $(LOCAL_IMAGE) $(DEV_IMAGE)
	docker push $(DEV_IMAGE)

# Push image to infoblox docker hub
push-release: image
	docker tag $(LOCAL_IMAGE) $(RELEASE_IMAGE)
	docker push $(RELEASE_IMAGE)

$(BINARY_NAME):
	mkdir -p bin
	go build -o bin/$(BINARY_NAME) ./driver/

$(CREATE_EA_DEFS):
	mkdir -p bin
	go build -o bin/$(CREATE_EA_DEFS) ./ea-defs/

# Delete binary for clean build
clean:
	rm -rf $(BINARY_NAME) $(CREATE_EA_DEFS) bin/
	rm -rf ./plugin
	docker rm build-plugin-container

create-plugin:
	docker plugin create $(PLUGIN_NAME) ./plugin
	docker plugin enable $(PLUGIN_NAME)

delete-plugin:
	docker plugin disable $(PLUGIN_NAME)
	docker plugin rm $(PLUGIN_NAME)

push-plugin:
	docker plugin push ${PLUGIN_NAME}

# Delete local docker images
clean-images:
	docker rmi -f $(LOCAL_IMAGE) $(DEV_IMAGE) $(RELEASE_IMAGE)

# Clean everything
clean-all: clean clean-images
