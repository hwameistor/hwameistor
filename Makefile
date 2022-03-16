IMAGE_REGISTRY ?= ghcr.io/hwameistor

GO_VERSION = $(shell go version)
BUILD_TIME = ${shell date +%Y-%m-%dT%H:%M:%SZ}
BUILD_VERSION = ${shell git rev-parse --short "HEAD^{commit}" 2>/dev/null}
BUILD_ENVS = CGO_ENABLED=0 GOOS=linux
BUILD_FLAGS = -X 'main.BUILDVERSION=${BUILD_VERSION}' -X 'main.BUILDTIME=${BUILD_TIME}' -X 'main.GOVERSION=${GO_VERSION}'
BUILD_OPTIONS = -a -mod vendor -installsuffix cgo -ldflags "${BUILD_FLAGS}"
BUILD_CMD = go build
OPERATOR_CMD = operator-sdk

DOCKER_SOCK_PATH=/var/run/docker.sock
DOCKER_MAKE_CMD = docker run --rm -v ${PROJECT_SOURCE_CODE_DIR}:${BUILDER_MOUNT_DST_DIR} -v ${DOCKER_SOCK_PATH}:${DOCKER_SOCK_PATH} -w ${BUILDER_MOUNT_DST_DIR} -i ${BUILDER_NAME}:${BUILDER_TAG}
DOCKER_DEBUG_CMD = docker run --rm -v ${PROJECT_SOURCE_CODE_DIR}:${BUILDER_MOUNT_DST_DIR} -v ${DOCKER_SOCK_PATH}:${DOCKER_SOCK_PATH} -w ${BUILDER_MOUNT_DST_DIR} -it ${BUILDER_NAME}:${BUILDER_TAG}
DOCKER_BUILDX_CMD_AMD64 = DOCKER_CLI_EXPERIMENTAL=enabled docker buildx build --platform=linux/amd64 -o type=docker
DOCKER_BUILDX_CMD_ARM64 = DOCKER_CLI_EXPERIMENTAL=enabled docker buildx build --platform=linux/arm64 -o type=docker
MUILT_ARCH_PUSH_CMD = ${PROJECT_SOURCE_CODE_DIR}/build/utils/docker-push-with-multi-arch.sh

PROJECT_SOURCE_CODE_DIR ?= $(CURDIR)
BINS_DIR = ${PROJECT_SOURCE_CODE_DIR}/_build
CMDS_DIR = ${PROJECT_SOURCE_CODE_DIR}/cmd

# image_tag/release_tag will be applied to all the images
IMAGE_TAG ?= 99.9-dev
RELEASE_TAG ?= $(shell tagged="$$(git describe --tags --match='v*' --abbrev=0 2> /dev/null)"; if [ "$$tagged" ] && [ "$$(git rev-list -n1 HEAD)" = "$$(git rev-list -n1 $$tagged)" ]; then echo $$tagged; fi)

MODULE_NAME = local-storage

BUILDER_NAME = ${IMAGE_REGISTRY}/${MODULE_NAME}-builder
BUILDER_TAG = latest
BUILDER_DOCKERFILE = ${PROJECT_SOURCE_CODE_DIR}/build/builder/Dockerfile
BUILDER_MOUNT_DST_DIR = /go/src/github.com/hwameistor/${MODULE_NAME}
BUILD_BIN = ${BINS_DIR}/${MODULE_NAME}
BUILD_MAIN = ${CMDS_DIR}/manager/main.go
IMAGE_NAME = ${IMAGE_REGISTRY}/${MODULE_NAME}
IMAGE_DOCKERFILE = ${PROJECT_SOURCE_CODE_DIR}/build/Dockerfile

.PHONY: debug
debug:
	${DOCKER_DEBUG_CMD} ash

.PHONY: builder
builder:
	docker build -t ${BUILDER_NAME}:${BUILDER_TAG} -f ${BUILDER_DOCKERFILE} ${PROJECT_SOURCE_CODE_DIR}

.PHONY: compile
compile:
	GOARCH=amd64 ${BUILD_ENVS} ${BUILD_CMD} ${BUILD_OPTIONS} -o ${BUILD_BIN} ${BUILD_MAIN}

.PHONY: compile_arm64
compile_arm64:
	GOARCH=arm64 ${BUILD_ENVS} ${BUILD_CMD} ${BUILD_OPTIONS} -o ${BUILD_BIN} ${BUILD_MAIN}

.PHONY: image
image:
	${DOCKER_MAKE_CMD} make compile
	docker build -t ${IMAGE_NAME}:${IMAGE_TAG} -f ${IMAGE_DOCKERFILE} ${PROJECT_SOURCE_CODE_DIR}

.PHONY: release
release:
	# build for amd64 version
	${DOCKER_MAKE_CMD} make compile
	${DOCKER_BUILDX_CMD_AMD64} -t ${IMAGE_NAME}:${RELEASE_TAG}-amd64 -f ${IMAGE_DOCKERFILE} ${PROJECT_SOURCE_CODE_DIR}
	# build for arm64 version
	${DOCKER_MAKE_CMD} make compile_arm64
	${DOCKER_BUILDX_CMD_ARM64} -t ${IMAGE_NAME}:${RELEASE_TAG}-arm64 -f ${IMAGE_DOCKERFILE} ${PROJECT_SOURCE_CODE_DIR}
	# push to a public registry
	${MUILT_ARCH_PUSH_CMD} ${IMAGE_NAME}:${RELEASE_TAG}

.PHONY: _gen-apis
_gen-apis:
	${OPERATOR_CMD} generate k8s
	${OPERATOR_CMD} generate crds
	/code-generator/generate-groups.sh all github.com/hwameistor/local-storage/pkg/apis/client github.com/hwameistor/local-storage/pkg/apis "v1alpha1" --go-header-file /code-generator/boilerplate.go.txt

.PHONY: apis
apis:
	${DOCKER_MAKE_CMD} make _gen-apis

.PHONY: vendor
vendor:
	go mod tidy -compat=1.17
	go mod vendor 

.PHONY: clean
clean:
	go clean -r -x
	rm -rf ${BINS_DIR}
	docker container prune -f
	docker rmi -f $(shell docker images -f dangling=true -qa)

unit-test:
	bash test/unit-test.sh

e2e-test:
	bash test/e2e-test.sh
