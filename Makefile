SOURCES=ipam-driver.go infoblox-ipam.go config.go
BINARY_NAME=ipam-driver
IMAGE_NAME=ipam-driver
LOCAL_IMAGE=$(IMAGE_NAME)
DEV_IMAGE=$(DOCKERHUB_ID)/$(IMAGE_NAME)  # Requires DOCKERHUB_ID environment variable
RELEASE_IMAGE=infoblox/$(IMAGE_NAME)

CREATE_EA_DEFS=create_ea_defs
CREATE_EA_DEFS_SOURCES=create_ea_defs.go config.go constants.go


# Build binary - this is the default target
build: $(BINARY_NAME) $(CREATE_EA_DEFS)


# Build binary and docker image
all: build image


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

$(BINARY_NAME): $(SOURCES)
	go build -o $(BINARY_NAME) ${SOURCES}

$(CREATE_EA_DEFS): $(CREATE_EA_DEFS_SOURCES)
	go build -o $(CREATE_EA_DEFS) ${CREATE_EA_DEFS_SOURCES}

# Delete binary for clean build
clean:
	rm -f $(BINARY_NAME) $(CREATE_EA_DEFS)

# Delete local docker images
clean-images:
	docker rmi -f $(LOCAL_IMAGE) $(DEV_IMAGE) $(RELEASE_IMAGE)

# Clean everything
clean-all: clean clean-images
