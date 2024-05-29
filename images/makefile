IMAGE_NAME=ghcr.io/fortytwoapps/kubelitedb
IMAGE_TAG=latest

.PHONY: docker-build docker-push

all: docker-build docker-push

docker-build:
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .

docker-push:
	docker push $(IMAGE_NAME):$(IMAGE_TAG)