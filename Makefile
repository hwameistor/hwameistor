include Makefile.variables
# local-disk-manager build definitions
LDM_MODULE_NAME = local-disk-manager
LDM_BUILD_INPUT = ${CMDS_DIR}/${LDM_MODULE_NAME}/diskmanager.go

# local-storage build definitions
LS_MODULE_NAME = local-storage
LS_BUILD_INPUT = ${CMDS_DIR}/${LS_MODULE_NAME}/storage.go

# scheduler build definitions
SCHEDULER_MODULE_NAME = scheduler
SCHEDULER_BUILD_INPUT = ${CMDS_DIR}/${SCHEDULER_MODULE_NAME}/scheduler.go

.PHONY: compile
compile: compile_ldm compile_ls compile_scheduler

.PHONY: image
image: build_ldm_image build_ls_image build_scheduler_image

.PHONY: release
release: _enable_buildx release_ldm release_ls release_scheduler

.PHONY: unit-test
unit-test:
	go test -race -coverprofile=coverage.txt -covermode=atomic ./pkg/...
	curl -s https://codecov.io/bash | bash

.PHONY: vendor
vendor:
	go mod tidy -compat=1.17
	go mod vendor

.PHONY: release_ldm
release_ldm:
	# build for amd64 version
	${DOCKER_MAKE_CMD} make compile_ldm
	${DOCKER_BUILDX_CMD_AMD64} -t ${LDM_IMAGE_NAME}:${RELEASE_TAG}-amd64 -f ${LDM_IMAGE_DOCKERFILE} ${PROJECT_SOURCE_CODE_DIR}
	# build for arm64 version
	${DOCKER_MAKE_CMD} make compile_ldm_arm64
	${DOCKER_BUILDX_CMD_ARM64} -t ${LDM_IMAGE_NAME}:${RELEASE_TAG}-arm64 -f ${LDM_IMAGE_DOCKERFILE} ${PROJECT_SOURCE_CODE_DIR}
	# push to a public registry
	${MUILT_ARCH_PUSH_CMD} ${LDM_IMAGE_NAME}:${RELEASE_TAG}

.PHONY: release_ls
release_ls:
	# build for amd64 version
	${DOCKER_MAKE_CMD} make compile_ls
	${DOCKER_BUILDX_CMD_AMD64} -t ${LS_IMAGE_NAME}:${RELEASE_TAG}-amd64 -f ${LS_IMAGE_DOCKERFILE} ${PROJECT_SOURCE_CODE_DIR}
	# build for arm64 version
	${DOCKER_MAKE_CMD} make compile_ls_arm64
	${DOCKER_BUILDX_CMD_ARM64} -t ${LS_IMAGE_NAME}:${RELEASE_TAG}-arm64 -f ${LS_IMAGE_DOCKERFILE} ${PROJECT_SOURCE_CODE_DIR}
	# push to a public registry
	${MUILT_ARCH_PUSH_CMD} ${LS_IMAGE_NAME}:${RELEASE_TAG}

.PHONY: release_scheduler
release_scheduler:
	# build for amd64 version
	${DOCKER_MAKE_CMD} make compile_scheduler
	${DOCKER_BUILDX_CMD_AMD64} -t ${SCHEDULER_IMAGE_NAME}:${RELEASE_TAG}-amd64 -f ${SCHEDULER_IMAGE_DOCKERFILE} ${PROJECT_SOURCE_CODE_DIR}
	# build for arm64 version
	${DOCKER_MAKE_CMD} make compile_scheduler_arm64
	${DOCKER_BUILDX_CMD_ARM64} -t ${SCHEDULER_IMAGE_NAME}:${RELEASE_TAG}-arm64 -f ${SCHEDULER_IMAGE_DOCKERFILE} ${PROJECT_SOURCE_CODE_DIR}
	# push to a public registry
	${MUILT_ARCH_PUSH_CMD} ${SCHEDULER_IMAGE_NAME}:${RELEASE_TAG}

.PHONY: build_ldm_image
build_ldm_image:
	@echo "Build local-disk-manager image ${LDM_IMAGE_NAME}:${IMAGE_TAG}"
	${DOCKER_MAKE_CMD} make compile_ldm
	docker build -t ${LDM_IMAGE_NAME}:${IMAGE_TAG} -f ${LDM_IMAGE_DOCKERFILE} ${PROJECT_SOURCE_CODE_DIR}

.PHONY: build_ls_image
build_ls_image:
	@echo "Build local-storage image ${LS_IMAGE_NAME}:${IMAGE_TAG}"
	${DOCKER_MAKE_CMD} make compile_ls
	docker build -t ${LS_IMAGE_NAME}:${IMAGE_TAG} -f ${LS_IMAGE_DOCKERFILE} ${PROJECT_SOURCE_CODE_DIR}

.PHONY: build_scheduler_image
build_scheduler_image:
	@echo "Build scheduler image ${SCHEDULER_IMAGE_NAME}:${IMAGE_TAG}"
	${DOCKER_MAKE_CMD} make compile_scheduler
	docker build -t ${SCHEDULER_IMAGE_NAME}:${IMAGE_TAG} -f ${SCHEDULER_IMAGE_DOCKERFILE} ${PROJECT_SOURCE_CODE_DIR}

.PHONY: apis
apis:
	${DOCKER_MAKE_CMD} make _gen-apis

.PHONY: builder
builder:
	docker build -t ${BUILDER_NAME}:${BUILDER_TAG} -f ${BUILDER_DOCKERFILE} ${PROJECT_SOURCE_CODE_DIR}

.PHONY: _gen-apis
_gen-apis:
	${OPERATOR_CMD} generate k8s
	${OPERATOR_CMD} generate crds
	bash hack/update-codegen.sh

.PHONY: compile_ldm
compile_ldm:
	GOARCH=amd64 ${BUILD_ENVS} ${BUILD_CMD} ${BUILD_OPTIONS} -o ${LDM_BUILD_OUTPUT} ${LDM_BUILD_INPUT}

.PHONY: compile_ldm_arm64
compile_ldm_arm64:
	GOARCH=arm64 ${BUILD_ENVS} ${BUILD_CMD} ${BUILD_OPTIONS} -o ${LDM_BUILD_OUTPUT} ${LDM_BUILD_INPUT}

.PHONY: compile_ls
compile_ls:
	GOARCH=amd64 ${BUILD_ENVS} ${BUILD_CMD} ${BUILD_OPTIONS} -o ${LS_BUILD_OUTPUT} ${LS_BUILD_INPUT}

.PHONY: compile_ls_arm64
compile_ls_arm64:
	GOARCH=arm64 ${BUILD_ENVS} ${BUILD_CMD} ${BUILD_OPTIONS} -o ${LS_BUILD_OUTPUT} ${LS_BUILD_INPUT}

.PHONY: compile_scheduler
compile_scheduler:
	GOARCH=amd64 ${BUILD_ENVS} ${BUILD_CMD} ${BUILD_OPTIONS} -o ${SCHEDULER_BUILD_OUTPUT} ${SCHEDULER_BUILD_INPUT}

.PHONY: compile_scheduler_arm64
compile_scheduler_arm64:
	GOARCH=arm64 ${BUILD_ENVS} ${BUILD_CMD} ${BUILD_OPTIONS} -o ${SCHEDULER_BUILD_OUTPUT} ${SCHEDULER_BUILD_INPUT}

.PHONY: _enable_buildx
_enable_buildx:
	@echo "Checking if buildx enabled"
	@if [[ "$(shell docker version -f '{{.Server.Experimental}}')" == "true" ]]; \
	then \
		docker buildx inspect mutil-platform-builder &>/dev/null; \
	        [ $$? == 0 ] && echo "ok" && exit 0; \
		docker buildx create --name mutil-platform-builder &>/dev/null&& echo "ok" && exit 0; \
	else \
		echo "experimental config of docker is false"; \
		exit 1; \
	fi


.PHONY: e2e-test
e2e-test:
	bash test/e2e-test.sh