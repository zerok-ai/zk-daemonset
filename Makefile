NAME = zk-daemonset
IMAGE_NAME = zk-daemonset
IMAGE_VERSION = 1.0.2

LOCATION ?= us-west1
PROJECT_ID ?= zerok-dev
REPOSITORY ?= zk-client

export GO111MODULE=on
export GOARCH=amd64

sync:
	go get -v ./...

build: sync
	echo "GOARCH=${GOARCH}"
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o $(NAME) cmd/main.go

docker-build: sync
	CGO_ENABLED=0 GOOS=linux $(ARCH) go build -v -o $(NAME) cmd/main.go
	docker build --no-cache -t $(IMAGE_PREFIX)$(IMAGE_NAME):$(IMAGE_VERSION) .

docker-build-gke: IMAGE_PREFIX := $(LOCATION)-docker.pkg.dev/$(PROJECT_ID)/$(REPOSITORY)/
docker-build-gke: ARCH := GOARCH=amd64
docker-build-gke: docker-build

docker-push-gke: IMAGE_PREFIX := $(LOCATION)-docker.pkg.dev/$(PROJECT_ID)/$(REPOSITORY)/
docker-push-gke:
	docker push $(IMAGE_PREFIX)$(IMAGE_NAME):$(IMAGE_VERSION)

docker-build-push-gke: docker-build-gke docker-push-gke

# ------- CI-CD ------------
ci-cd-build: build

ci-cd-build-migration: