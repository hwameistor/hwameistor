MEMBER_NAME = localstorage-local-storage
MEMBER_IMAGE_DIR = ${PROJECT_SOURCE_CODE_DIR}/images/member
MEMBER_BUILD_BIN = ${BINS_DIR}/${MEMBER_NAME}-run
MEMBER_BUILD_MAIN = ${CMDS_DIR}/manager/main.go

MEMBER_IMAGE_NAME = ${DOCKER_REGISTRY}/${MEMBER_NAME}
RELEASE_MEMBER_IMAGE_NAME=${RELEASE_DOCKER_REGISTRY}/${MEMBER_NAME}

.PHONY: member
member:
	GOARCH=amd64 ${BUILD_ENVS} ${BUILD_CMD} ${BUILD_OPTIONS} -o ${MEMBER_BUILD_BIN} ${MEMBER_BUILD_MAIN}

.PHONY: member_arm64
member_arm64:
	GOARCH=arm64 ${BUILD_ENVS} ${BUILD_CMD} ${BUILD_OPTIONS} -o ${MEMBER_BUILD_BIN} ${MEMBER_BUILD_MAIN}

.PHONY: member_image
member_image:
	${DOCKER_MAKE_CMD} make member
	docker build -t ${MEMBER_IMAGE_NAME}:${IMAGE_TAG} -f ${MEMBER_IMAGE_DIR}/Dockerfile ${PROJECT_SOURCE_CODE_DIR}
	docker push ${MEMBER_IMAGE_NAME}:${IMAGE_TAG}

.PHONY: linkcodes
linkcodes:
	mkdir -p /Users/lijie$(CURDIR)
	rsync -rupE $(CURDIR)/* /Users/lijie$(CURDIR)
	#ln -s ..$(CURDIR)/ /Users/lijie$(CURDIR)

.PHONY: member_release
member_release:
	#make linkcodes
	# build for amd64 version
	${DOCKER_MAKE_CMD} make member
	${DOCKER_BUILDX_CMD_AMD64} -t ${RELEASE_MEMBER_IMAGE_NAME}:${RELEASE_TAG}-amd64 -f ${MEMBER_IMAGE_DIR}/Dockerfile ${PROJECT_SOURCE_CODE_DIR}
	# build for arm64 version
	${DOCKER_MAKE_CMD} make member_arm64
	${DOCKER_BUILDX_CMD_ARM64} -t ${RELEASE_MEMBER_IMAGE_NAME}:${RELEASE_TAG}-arm64 -f ${MEMBER_IMAGE_DIR}/Dockerfile ${PROJECT_SOURCE_CODE_DIR}
	# push to a public registry
	${MUILT_ARCH_PUSH_CMD} ${RELEASE_MEMBER_IMAGE_NAME}:${RELEASE_TAG}
