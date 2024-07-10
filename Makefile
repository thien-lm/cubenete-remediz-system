REGISTRY                    := registry.fke.fptcloud.com
IMAGE_PREFIX                := $(REGISTRY)/xplat-fke
NAME                        := auto-repair-controller
REPO_ROOT                   := $(shell dirname $(realpath $(lastword ${MAKEFILE_LIST})))
VERSION                     := $(shell cat "${REPO_ROOT}/VERSION")

.PHONY: docker-images
docker-images:
	@docker build -t $(IMAGE_PREFIX)/$(NAME):$(VERSION) -t $(IMAGE_PREFIX)/$(NAME):latest -f Dockerfile -m 6g .

.PHONY: docker-push
docker-push:
	@docker push $(IMAGE_PREFIX)/$(NAME):$(VERSION)
