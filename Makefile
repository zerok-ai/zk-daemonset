NAME = zk-daemonset-amd64
NAME_AMD64 = zk-daemonset-amd64
NAME_ARM64 = zk-daemonset-arm64
IMAGE_NAME = zk-daemonset
IMAGE_VERSION = 1.0.2

LOCATION ?= us-west1
PROJECT_ID ?= zerok-dev
REPOSITORY ?= zk-client

BUILDER_NAME = multi-platform-builder

export GO111MODULE=on
export GOARCH=amd64
export GOPRIVATE=github.com/zerok-ai/zk-utils-go

sync:
	go get -v ./...

build: sync
	go build -v -o $(NAME) cmd/main.go



docker-build: sync
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o $(NAME_AMD64) cmd/main.go
	docker build --no-cache -t $(IMAGE_PREFIX)$(IMAGE_NAME):$(IMAGE_VERSION) .

docker-build-gke: IMAGE_PREFIX := $(LOCATION)-docker.pkg.dev/$(PROJECT_ID)/$(REPOSITORY)/
docker-build-gke: docker-build

docker-push-gke: IMAGE_PREFIX := $(LOCATION)-docker.pkg.dev/$(PROJECT_ID)/$(REPOSITORY)/
docker-push-gke:
	docker push $(IMAGE_PREFIX)$(IMAGE_NAME):$(IMAGE_VERSION)

docker-build-push-gke: docker-build-gke docker-push-gke

docker-build-push-multiarch: IMAGE_PREFIX := $(LOCATION)-docker.pkg.dev/$(PROJECT_ID)/$(REPOSITORY)/
docker-build-push-multiarch:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o bin/$(NAME_AMD64) cmd/main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -v -o bin/$(NAME_ARM64) cmd/main.go
	#Adding remove here again to account for the case when buildx was not removed in previous run.
	docker buildx rm ${BUILDER_NAME} || true
	docker buildx create --use --platform=linux/arm64,linux/amd64 --name ${BUILDER_NAME}
	docker buildx build --platform=linux/arm64,linux/amd64 --push \
	--tag $(IMAGE_PREFIX)$(IMAGE_NAME):$(IMAGE_VERSION) .
	docker buildx rm ${BUILDER_NAME}

# ------- CI-CD ------------
debug-build: sync
	dlv debug --headless --listen=:2345 --api-version=2 --accept-multiclient cmd/main.go
ci-cd-build: sync
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/$(NAME_AMD64) cmd/main.go
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o bin/$(NAME_ARM64) cmd/main.go

ci-cd-build-migration:
