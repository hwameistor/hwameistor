MEMBER_NAME = local-storage
MEMBER_IMAGE_DIR = ${PROJECT_SOURCE_CODE_DIR}/images/member
MEMBER_BUILD_BIN = ${BINS_DIR}/${MEMBER_NAME}-run
MEMBER_BUILD_MAIN = ${CMDS_DIR}/manager/main.go

.PHONY: member
member:
	GOARCH=amd64 ${BUILD_ENVS} ${BUILD_CMD} ${BUILD_OPTIONS} -o ${MEMBER_BUILD_BIN} ${MEMBER_BUILD_MAIN}

.PHONY: member_arm64
member_arm64:
	GOARCH=arm64 ${BUILD_ENVS} ${BUILD_CMD} ${BUILD_OPTIONS} -o ${MEMBER_BUILD_BIN} ${MEMBER_BUILD_MAIN}

.PHONY: member_image
member_image:
	${DOCKER_MAKE_CMD} make member
	docker build -t ${DOCKER_REGISTRY}:${IMAGE_TAG} -f ${MEMBER_IMAGE_DIR}/Dockerfile ${PROJECT_SOURCE_CODE_DIR}
	docker push ${DOCKER_REGISTRY}:${IMAGE_TAG}

.PHONY: member_release
member_release:
	# build for amd64 version
	${DOCKER_MAKE_CMD} make member
	${DOCKER_BUILDX_CMD_AMD64} -t ${RELEASE_DOCKER_REGISTRY}:${RELEASE_TAG}-amd64 -f ${MEMBER_IMAGE_DIR}/Dockerfile ${PROJECT_SOURCE_CODE_DIR}
	# build for arm64 version
	${DOCKER_MAKE_CMD} make member_arm64
	${DOCKER_BUILDX_CMD_ARM64} -t ${RELEASE_DOCKER_REGISTRY}:${RELEASE_TAG}-arm64 -f ${MEMBER_IMAGE_DIR}/Dockerfile ${PROJECT_SOURCE_CODE_DIR}
	# push to a public registry
	${MUILT_ARCH_PUSH_CMD} ${RELEASE_DOCKER_REGISTRY}:${RELEASE_TAG}
