include Makefile.variables

.PHONY: debug
debug:
	${DOCKER_DEBUG_CMD} ash

.PHONY: apiserver_swag
apiserver_swag:
	${DOCKER_MAKE_CMD} swag init -d ./cmd/${APISERVER_MODULE_NAME} -o ./pkg/${APISERVER_MODULE_NAME}/docs --parseVendor --parseDependency --parseInternal --propertyStrategy pascalcase --parseDepth 5

.PHONY: apiserver_run
apiserver_run: apiserver_swag
	echo "Please browse at http://127.0.0.1/swagger/index.html"
	GIN_MODE=debug go run ${BUILD_OPTIONS} ${APISERVER_BUILD_INPUT}

.PHONY: compile
compile: compile_ldm compile_ls compile_scheduler compile_admission compile_evictor compile_exporter compile_apiserver compile_failover compile_auditor compile_pvc-autoresizer compile_lda

.PHONY: image
image: build_ldm_image build_ls_image build_scheduler_image build_admission_image build_evictor_image build_exporter_image build_apiserver_image build_failover_image build_auditor_image build_pvc-autoresizer_image build_lda_image


.PHONY: arm-image
arm-image: build_ldm_image_arm64 build_ls_image_arm64 build_scheduler_image_arm64 build_admission_image_arm64 build_evictor_image_arm64 build_exporter_image_arm64 build_apiserver_image_arm64 build_failover_image_arm64 build_auditor_image_arm64 build_pvc-autoresizer_image_arm64 build_lda_image_arm64


.PHONY: release
release: release_ldm release_ls release_scheduler release_admission release_evictor release_exporter release_apiserver release_failover release_auditor release_pvc-autoresizer release_lda

.PHONY: unit-test
unit-test:
	go test -race -coverprofile=coverage.txt -covermode=atomic `go list ./pkg/... | grep -v -E './pkg/local-storage/member|./pkg/scheduler|./pkg/evictor|./pkg/apiserver'`
	curl -s https://codecov.io/bash | bash

.PHONY: vendor
vendor:
	go mod tidy -compat=1.18
	go mod vendor

#### for LDM #########
LDM_MODULE_NAME = local-disk-manager
LDM_BUILD_INPUT = ${CMDS_DIR}/${LDM_MODULE_NAME}/diskmanager.go

.PHONY: compile_ldm
compile_ldm:
	GOARCH=amd64 ${BUILD_ENVS} ${BUILD_CMD} ${BUILD_OPTIONS} -o ${LDM_BUILD_OUTPUT} ${LDM_BUILD_INPUT}

.PHONY: compile_ldm_arm64
compile_ldm_arm64:
	GOARCH=arm64 ${BUILD_ENVS} ${BUILD_CMD} ${BUILD_OPTIONS} -o ${LDM_BUILD_OUTPUT} ${LDM_BUILD_INPUT}

.PHONY: build_ldm_image
build_ldm_image:
	@echo "Build local-disk-manager image ${LDM_IMAGE_NAME}:${IMAGE_TAG}"
	${DOCKER_MAKE_CMD} make compile_ldm
	docker build -t ${LDM_IMAGE_NAME}:${IMAGE_TAG} -f ${LDM_IMAGE_DOCKERFILE} ${PROJECT_SOURCE_CODE_DIR}

.PHONY: build_ldm_image_arm64
build_ldm_image_arm64:
	@echo "Build local-disk-manager image ${LDM_IMAGE_NAME}:${IMAGE_TAG}"
	${DOCKER_MAKE_CMD} make compile_ldm_arm64
	${DOCKER_BUILDX_CMD_ARM64} -t ${LDM_IMAGE_NAME}:${IMAGE_TAG} -f ${LDM_IMAGE_DOCKERFILE} ${PROJECT_SOURCE_CODE_DIR}

.PHONY: release_ldm
release_ldm:
	# build for amd64 version
	${DOCKER_MAKE_CMD} make compile_ldm
	${DOCKER_BUILDX_CMD_AMD64} -t ${LDM_IMAGE_NAME}:${RELEASE_TAG}-amd64 -f ${LDM_IMAGE_DOCKERFILE} ${PROJECT_SOURCE_CODE_DIR}
	# build for arm64 version
	${DOCKER_MAKE_CMD} make compile_ldm_arm64
	${DOCKER_BUILDX_CMD_ARM64} -t ${LDM_IMAGE_NAME}:${RELEASE_TAG}-arm64 -f ${LDM_IMAGE_DOCKERFILE} ${PROJECT_SOURCE_CODE_DIR}
	# push to a public registry
	${MUILT_ARCH_PUSH_CMD} -i ${LDM_IMAGE_NAME}:${RELEASE_TAG}


#### for LS ##########
LS_MODULE_NAME = local-storage
LS_BUILD_INPUT = ${CMDS_DIR}/${LS_MODULE_NAME}/storage.go

.PHONY: compile_ls
compile_ls:
	GOARCH=amd64 ${BUILD_ENVS} ${BUILD_CMD} ${BUILD_OPTIONS} -o ${LS_BUILD_OUTPUT} ${LS_BUILD_INPUT}

.PHONY: compile_ls_arm64
compile_ls_arm64:
	GOARCH=arm64 ${BUILD_ENVS} ${BUILD_CMD} ${BUILD_OPTIONS} -o ${LS_BUILD_OUTPUT} ${LS_BUILD_INPUT}

.PHONY: build_ls_image
build_ls_image:
	@echo "Build local-storage image ${LS_IMAGE_NAME}:${IMAGE_TAG}"
	${DOCKER_MAKE_CMD} make compile_ls
	${DOCKER_BUILDX_CMD_AMD64} -t ${LS_IMAGE_NAME}:${IMAGE_TAG} -f ${LS_IMAGE_DOCKERFILE} ${PROJECT_SOURCE_CODE_DIR}

.PHONY: build_ls_image_arm64
build_ls_image_arm64:
	@echo "Build local-storage image ${LS_IMAGE_NAME}:${IMAGE_TAG}"
	${DOCKER_MAKE_CMD} make compile_ls_arm64
	${DOCKER_BUILDX_CMD_ARM64} -t ${LS_IMAGE_NAME}:${IMAGE_TAG} -f ${LS_IMAGE_DOCKERFILE} ${PROJECT_SOURCE_CODE_DIR}

.PHONY: release_ls
release_ls:
	# build for amd64 version
	${DOCKER_MAKE_CMD} make compile_ls
	${DOCKER_BUILDX_CMD_AMD64} -t ${LS_IMAGE_NAME}:${RELEASE_TAG}-amd64 -f ${LS_IMAGE_DOCKERFILE} ${PROJECT_SOURCE_CODE_DIR}
	# build for arm64 version
	${DOCKER_MAKE_CMD} make compile_ls_arm64
	${DOCKER_BUILDX_CMD_ARM64} -t ${LS_IMAGE_NAME}:${RELEASE_TAG}-arm64 -f ${LS_IMAGE_DOCKERFILE} ${PROJECT_SOURCE_CODE_DIR}
	# push to a public registry
	${MUILT_ARCH_PUSH_CMD} -i ${LS_IMAGE_NAME}:${RELEASE_TAG}


#### for Scheduler ##########
SCHEDULER_MODULE_NAME = scheduler
SCHEDULER_BUILD_INPUT = ${CMDS_DIR}/${SCHEDULER_MODULE_NAME}/scheduler.go

.PHONY: compile_scheduler
compile_scheduler:
	GOARCH=amd64 ${BUILD_ENVS} ${BUILD_CMD} ${BUILD_OPTIONS} -o ${SCHEDULER_BUILD_OUTPUT} ${SCHEDULER_BUILD_INPUT}

.PHONY: compile_scheduler_arm64
compile_scheduler_arm64:
	GOARCH=arm64 ${BUILD_ENVS} ${BUILD_CMD} ${BUILD_OPTIONS} -o ${SCHEDULER_BUILD_OUTPUT} ${SCHEDULER_BUILD_INPUT}

.PHONY: build_scheduler_image
build_scheduler_image:
	@echo "Build scheduler image ${SCHEDULER_IMAGE_NAME}:${IMAGE_TAG}"
	${DOCKER_MAKE_CMD} make compile_scheduler
	docker build -t ${SCHEDULER_IMAGE_NAME}:${IMAGE_TAG} -f ${SCHEDULER_IMAGE_DOCKERFILE} ${PROJECT_SOURCE_CODE_DIR}

.PHONY: build_scheduler_image_arm64
build_scheduler_image_arm64:
	@echo "Build scheduler image ${SCHEDULER_IMAGE_NAME}:${IMAGE_TAG}"
	${DOCKER_MAKE_CMD} make compile_scheduler_arm64
	${DOCKER_BUILDX_CMD_ARM64} -t ${SCHEDULER_IMAGE_NAME}:${IMAGE_TAG} -f ${SCHEDULER_IMAGE_DOCKERFILE} ${PROJECT_SOURCE_CODE_DIR}

.PHONY: release_scheduler
release_scheduler:
	# build for amd64 version
	${DOCKER_MAKE_CMD} make compile_scheduler
	${DOCKER_BUILDX_CMD_AMD64} -t ${SCHEDULER_IMAGE_NAME}:${RELEASE_TAG}-amd64 -f ${SCHEDULER_IMAGE_DOCKERFILE} ${PROJECT_SOURCE_CODE_DIR}
	# build for arm64 version
	${DOCKER_MAKE_CMD} make compile_scheduler_arm64
	${DOCKER_BUILDX_CMD_ARM64} -t ${SCHEDULER_IMAGE_NAME}:${RELEASE_TAG}-arm64 -f ${SCHEDULER_IMAGE_DOCKERFILE} ${PROJECT_SOURCE_CODE_DIR}
	# push to a public registry
	${MUILT_ARCH_PUSH_CMD} -i ${SCHEDULER_IMAGE_NAME}:${RELEASE_TAG}


#### for Admission Controller ##########
ADMISSION_MODULE_NAME = admission
ADMISSION_BUILD_INPUT = ${CMDS_DIR}/${ADMISSION_MODULE_NAME}/admission.go

.PHONY: compile_admission
compile_admission:
	GOARCH=amd64 ${BUILD_ENVS} ${BUILD_CMD} ${BUILD_OPTIONS} -o ${ADMISSION_BUILD_OUTPUT} ${ADMISSION_BUILD_INPUT}

.PHONY: compile_admission_arm64
compile_admission_arm64:
	GOARCH=arm64 ${BUILD_ENVS} ${BUILD_CMD} ${BUILD_OPTIONS} -o ${ADMISSION_BUILD_OUTPUT} ${ADMISSION_BUILD_INPUT}

.PHONY: build_admission_image
build_admission_image:
	@echo "Build admission image ${ADMISSION_IMAGE_NAME}:${IMAGE_TAG}"
	${DOCKER_MAKE_CMD} make compile_admission
	docker build -t ${ADMISSION_IMAGE_NAME}:${IMAGE_TAG} -f ${ADMISSION_IMAGE_DOCKERFILE} ${PROJECT_SOURCE_CODE_DIR}

.PHONY: build_admission_image_arm64
build_admission_image_arm64:
	@echo "Build admission image ${ADMISSION_IMAGE_NAME}:${IMAGE_TAG}"
	${DOCKER_MAKE_CMD} make compile_admission_arm64
	${DOCKER_BUILDX_CMD_ARM64} -t ${ADMISSION_IMAGE_NAME}:${IMAGE_TAG} -f ${ADMISSION_IMAGE_DOCKERFILE} ${PROJECT_SOURCE_CODE_DIR}

.PHONY: release_admission
release_admission:
	# build for amd64 version
	${DOCKER_MAKE_CMD} make compile_admission
	${DOCKER_BUILDX_CMD_AMD64} -t ${ADMISSION_IMAGE_NAME}:${RELEASE_TAG}-amd64 -f ${ADMISSION_IMAGE_DOCKERFILE} ${PROJECT_SOURCE_CODE_DIR}
	# build for arm64 version
	${DOCKER_MAKE_CMD} make compile_admission_arm64
	${DOCKER_BUILDX_CMD_ARM64} -t ${ADMISSION_IMAGE_NAME}:${RELEASE_TAG}-arm64 -f ${ADMISSION_IMAGE_DOCKERFILE} ${PROJECT_SOURCE_CODE_DIR}
	# push to a public registry
	${MUILT_ARCH_PUSH_CMD} -i ${ADMISSION_IMAGE_NAME}:${RELEASE_TAG}


#### for Evitor ##########
EVICTOR_MODULE_NAME = evictor
EVICTOR_BUILD_INPUT = ${CMDS_DIR}/${EVICTOR_MODULE_NAME}/main.go

.PHONY: compile_evictor
compile_evictor:
	GOARCH=amd64 ${BUILD_ENVS} ${BUILD_CMD} ${BUILD_OPTIONS} -o ${EVICTOR_BUILD_OUTPUT} ${EVICTOR_BUILD_INPUT}

.PHONY: compile_evictor_arm64
compile_evictor_arm64:
	GOARCH=arm64 ${BUILD_ENVS} ${BUILD_CMD} ${BUILD_OPTIONS} -o ${EVICTOR_BUILD_OUTPUT} ${EVICTOR_BUILD_INPUT}

.PHONY: build_evictor_image
build_evictor_image:
	@echo "Build evictor image ${EVICTOR_IMAGE_NAME}:${IMAGE_TAG}"
	${DOCKER_MAKE_CMD} make compile_evictor
	docker build -t ${EVICTOR_IMAGE_NAME}:${IMAGE_TAG} -f ${EVICTOR_IMAGE_DOCKERFILE} ${PROJECT_SOURCE_CODE_DIR}

.PHONY: build_evictor_image_arm64
build_evictor_image_arm64:
	@echo "Build evictor image ${EVICTOR_IMAGE_NAME}:${IMAGE_TAG}"
	${DOCKER_MAKE_CMD} make compile_evictor_arm64
	${DOCKER_BUILDX_CMD_ARM64} -t ${EVICTOR_IMAGE_NAME}:${IMAGE_TAG} -f ${EVICTOR_IMAGE_DOCKERFILE} ${PROJECT_SOURCE_CODE_DIR}

.PHONY: release_evictor
release_evictor:
	# build for amd64 version
	${DOCKER_MAKE_CMD} make compile_evictor
	${DOCKER_BUILDX_CMD_AMD64} -t ${EVICTOR_IMAGE_NAME}:${RELEASE_TAG}-amd64 -f ${EVICTOR_IMAGE_DOCKERFILE} ${PROJECT_SOURCE_CODE_DIR}
	# build for arm64 version
	${DOCKER_MAKE_CMD} make compile_evictor_arm64
	${DOCKER_BUILDX_CMD_ARM64} -t ${EVICTOR_IMAGE_NAME}:${RELEASE_TAG}-arm64 -f ${EVICTOR_IMAGE_DOCKERFILE} ${PROJECT_SOURCE_CODE_DIR}
	# push to a public registry
	${MUILT_ARCH_PUSH_CMD} -i ${EVICTOR_IMAGE_NAME}:${RELEASE_TAG}


#### for Exporter ##########
EXPORTER_MODULE_NAME = exporter
EXPORTER_BUILD_INPUT = ${CMDS_DIR}/${EXPORTER_MODULE_NAME}/main.go

.PHONY: compile_exporter
compile_exporter:
	GOARCH=amd64 ${BUILD_ENVS} ${BUILD_CMD} ${BUILD_OPTIONS} -o ${EXPORTER_BUILD_OUTPUT} ${EXPORTER_BUILD_INPUT}

.PHONY: compile_exporter_arm64
compile_exporter_arm64:
	GOARCH=arm64 ${BUILD_ENVS} ${BUILD_CMD} ${BUILD_OPTIONS} -o ${EXPORTER_BUILD_OUTPUT} ${EXPORTER_BUILD_INPUT}

.PHONY: build_exporter_image
build_exporter_image:
	@echo "Build exporter image ${EXPORTER_IMAGE_NAME}:${IMAGE_TAG}"
	${DOCKER_MAKE_CMD} make compile_exporter
	docker build -t ${EXPORTER_IMAGE_NAME}:${IMAGE_TAG} -f ${EXPORTER_IMAGE_DOCKERFILE} ${PROJECT_SOURCE_CODE_DIR}

.PHONY: build_exporter_image_arm64
build_exporter_image_arm64:
	@echo "Build exporter image ${EXPORTER_IMAGE_NAME}:${IMAGE_TAG}"
	${DOCKER_MAKE_CMD} make compile_exporter_arm64
	${DOCKER_BUILDX_CMD_ARM64} -t ${EXPORTER_IMAGE_NAME}:${IMAGE_TAG} -f ${EXPORTER_IMAGE_DOCKERFILE} ${PROJECT_SOURCE_CODE_DIR}

.PHONY: release_exporter
release_exporter:
	# build for amd64 version
	${DOCKER_MAKE_CMD} make compile_exporter
	${DOCKER_BUILDX_CMD_AMD64} -t ${EXPORTER_IMAGE_NAME}:${RELEASE_TAG}-amd64 -f ${EXPORTER_IMAGE_DOCKERFILE} ${PROJECT_SOURCE_CODE_DIR}
	# build for arm64 version
	${DOCKER_MAKE_CMD} make compile_exporter_arm64
	${DOCKER_BUILDX_CMD_ARM64} -t ${EXPORTER_IMAGE_NAME}:${RELEASE_TAG}-arm64 -f ${EXPORTER_IMAGE_DOCKERFILE} ${PROJECT_SOURCE_CODE_DIR}
	# push to a public registry
	${MUILT_ARCH_PUSH_CMD} -i ${EXPORTER_IMAGE_NAME}:${RELEASE_TAG}


#### for APIServer ##########
APISERVER_MODULE_NAME = apiserver
APISERVER_BUILD_INPUT = ${CMDS_DIR}/${APISERVER_MODULE_NAME}/main.go

.PHONY: compile_apiserver
compile_apiserver:
	GOARCH=amd64 ${BUILD_ENVS} ${BUILD_CMD} ${BUILD_OPTIONS} -o ${APISERVER_BUILD_OUTPUT} ${APISERVER_BUILD_INPUT}

.PHONY: compile_apiserver_arm64
compile_apiserver_arm64:
	GOARCH=arm64 ${BUILD_ENVS} ${BUILD_CMD} ${BUILD_OPTIONS} -o ${APISERVER_BUILD_OUTPUT} ${APISERVER_BUILD_INPUT}

.PHONY: build_apiserver_image
build_apiserver_image:
	@echo "Build hwameistor apiserver image ${APISERVER_IMAGE_NAME}:${IMAGE_TAG}"
	${DOCKER_MAKE_CMD} make compile_apiserver
	docker build -t ${APISERVER_IMAGE_NAME}:${IMAGE_TAG} -f ${APISERVER_IMAGE_DOCKERFILE} ${PROJECT_SOURCE_CODE_DIR}

.PHONY: build_apiserver_image_arm64
build_apiserver_image_arm64:
	@echo "Build apiserver image ${APISERVER_IMAGE_NAME}:${IMAGE_TAG}"
	${DOCKER_MAKE_CMD} make compile_apiserver_arm64
	${DOCKER_BUILDX_CMD_ARM64} -t ${APISERVER_IMAGE_NAME}:${IMAGE_TAG} -f ${APISERVER_IMAGE_DOCKERFILE} ${PROJECT_SOURCE_CODE_DIR}

.PHONY: release_apiserver
release_apiserver:
	# build for amd64 version
	${DOCKER_MAKE_CMD} make compile_apiserver
	${DOCKER_BUILDX_CMD_AMD64} -t ${APISERVER_IMAGE_NAME}:${RELEASE_TAG}-amd64 -f ${APISERVER_IMAGE_DOCKERFILE} ${PROJECT_SOURCE_CODE_DIR}
	# build for arm64 version
	${DOCKER_MAKE_CMD} make compile_apiserver_arm64
	${DOCKER_BUILDX_CMD_ARM64} -t ${APISERVER_IMAGE_NAME}:${RELEASE_TAG}-arm64 -f ${APISERVER_IMAGE_DOCKERFILE} ${PROJECT_SOURCE_CODE_DIR}
	# push to a public registry
	${MUILT_ARCH_PUSH_CMD} -i ${APISERVER_IMAGE_NAME}:${RELEASE_TAG}


#### for Failover ##########
FAILOVER_MODULE_NAME = failover-assistant
FAILOVER_BUILD_INPUT = ${CMDS_DIR}/${FAILOVER_MODULE_NAME}/main.go

.PHONY: compile_failover
compile_failover:
	GOARCH=amd64 ${BUILD_ENVS} ${BUILD_CMD} ${BUILD_OPTIONS} -o ${FAILOVER_BUILD_OUTPUT} ${FAILOVER_BUILD_INPUT}

.PHONY: compile_failover_arm64
compile_failover_arm64:
	GOARCH=arm64 ${BUILD_ENVS} ${BUILD_CMD} ${BUILD_OPTIONS} -o ${FAILOVER_BUILD_OUTPUT} ${FAILOVER_BUILD_INPUT}

.PHONY: build_failover_image
build_failover_image:
	@echo "Build hwameistor failover image ${FAILOVER_IMAGE_NAME}:${IMAGE_TAG}"
	${DOCKER_MAKE_CMD} make compile_failover
	docker build -t ${FAILOVER_IMAGE_NAME}:${IMAGE_TAG} -f ${FAILOVER_IMAGE_DOCKERFILE} ${PROJECT_SOURCE_CODE_DIR}

.PHONY: build_failover_image_arm64
build_failover_image_arm64:
	@echo "Build failover image ${FAILOVER_IMAGE_NAME}:${IMAGE_TAG}"
	${DOCKER_MAKE_CMD} make compile_failover_arm64
	${DOCKER_BUILDX_CMD_ARM64} -t ${FAILOVER_IMAGE_NAME}:${IMAGE_TAG} -f ${FAILOVER_IMAGE_DOCKERFILE} ${PROJECT_SOURCE_CODE_DIR}

.PHONY: release_failover
release_failover:
	# build for amd64 version
	${DOCKER_MAKE_CMD} make compile_failover
	${DOCKER_BUILDX_CMD_AMD64} -t ${FAILOVER_IMAGE_NAME}:${RELEASE_TAG}-amd64 -f ${FAILOVER_IMAGE_DOCKERFILE} ${PROJECT_SOURCE_CODE_DIR}
	# build for arm64 version
	${DOCKER_MAKE_CMD} make compile_failover_arm64
	${DOCKER_BUILDX_CMD_ARM64} -t ${FAILOVER_IMAGE_NAME}:${RELEASE_TAG}-arm64 -f ${FAILOVER_IMAGE_DOCKERFILE} ${PROJECT_SOURCE_CODE_DIR}
	# push to a public registry
	${MUILT_ARCH_PUSH_CMD} -i ${FAILOVER_IMAGE_NAME}:${RELEASE_TAG}


#### for Auditor ##########
AUDITOR_MODULE_NAME = auditor
AUDITOR_BUILD_INPUT = ${CMDS_DIR}/${AUDITOR_MODULE_NAME}/main.go

.PHONY: compile_auditor
compile_auditor:
	GOARCH=amd64 ${BUILD_ENVS} ${BUILD_CMD} ${BUILD_OPTIONS} -o ${AUDITOR_BUILD_OUTPUT} ${AUDITOR_BUILD_INPUT}

.PHONY: compile_auditor_arm64
compile_auditor_arm64:
	GOARCH=arm64 ${BUILD_ENVS} ${BUILD_CMD} ${BUILD_OPTIONS} -o ${AUDITOR_BUILD_OUTPUT} ${AUDITOR_BUILD_INPUT}

.PHONY: build_auditor_image
build_auditor_image:
	@echo "Build hwameistor auditor image ${AUDITOR_IMAGE_NAME}:${IMAGE_TAG}"
	${DOCKER_MAKE_CMD} make compile_auditor
	docker build -t ${AUDITOR_IMAGE_NAME}:${IMAGE_TAG} -f ${AUDITOR_IMAGE_DOCKERFILE} ${PROJECT_SOURCE_CODE_DIR}

.PHONY: build_auditor_image_arm64
build_auditor_image_arm64:
	@echo "Build auditor image ${AUDITOR_IMAGE_NAME}:${IMAGE_TAG}"
	${DOCKER_MAKE_CMD} make compile_auditor_arm64
	${DOCKER_BUILDX_CMD_ARM64} -t ${AUDITOR_IMAGE_NAME}:${IMAGE_TAG} -f ${AUDITOR_IMAGE_DOCKERFILE} ${PROJECT_SOURCE_CODE_DIR}

.PHONY: release_auditor
release_auditor:
	# build for amd64 version
	${DOCKER_MAKE_CMD} make compile_auditor
	${DOCKER_BUILDX_CMD_AMD64} -t ${AUDITOR_IMAGE_NAME}:${RELEASE_TAG}-amd64 -f ${AUDITOR_IMAGE_DOCKERFILE} ${PROJECT_SOURCE_CODE_DIR}
	# build for arm64 version
	${DOCKER_MAKE_CMD} make compile_auditor_arm64
	${DOCKER_BUILDX_CMD_ARM64} -t ${AUDITOR_IMAGE_NAME}:${RELEASE_TAG}-arm64 -f ${AUDITOR_IMAGE_DOCKERFILE} ${PROJECT_SOURCE_CODE_DIR}
	# push to a public registry
	${MUILT_ARCH_PUSH_CMD} -i ${AUDITOR_IMAGE_NAME}:${RELEASE_TAG}


#### for PVC AutoResizer ##########
PVC-AUTORESIZER_MODULE_NAME = pvc-autoresizer
PVC-AUTORESIZER_BUILD_INPUT = ${CMDS_DIR}/${PVC-AUTORESIZER_MODULE_NAME}/main.go

.PHONY: compile_pvc-autoresizer
compile_pvc-autoresizer:
	GOARCH=amd64 ${BUILD_ENVS} ${BUILD_CMD} ${BUILD_OPTIONS} -o ${PVC-AUTORESIZER_BUILD_OUTPUT} ${PVC-AUTORESIZER_BUILD_INPUT}

.PHONY: compile_pvc-autoresizer_arm64
compile_pvc-autoresizer_arm64:
	GOARCH=arm64 ${BUILD_ENVS} ${BUILD_CMD} ${BUILD_OPTIONS} -o ${PVC-AUTORESIZER_BUILD_OUTPUT} ${PVC-AUTORESIZER_BUILD_INPUT}

.PHONY: build_pvc-autoresizer_image
build_pvc-autoresizer_image:
	@echo "Build hwameistor pvc-autoresizer image ${PVC-AUTORESIZER_IMAGE_NAME}:${IMAGE_TAG}"
	${DOCKER_MAKE_CMD} make compile_pvc-autoresizer
	docker build -t ${PVC-AUTORESIZER_IMAGE_NAME}:${IMAGE_TAG} -f ${PVC-AUTORESIZER_IMAGE_DOCKERFILE} ${PROJECT_SOURCE_CODE_DIR}

.PHONY: build_pvc-autoresizer_image_arm64
build_pvc-autoresizer_image_arm64:
	@echo "Build hwameistor pvc-autoresizer image ${PVC-AUTORESIZER_IMAGE_NAME}:${IMAGE_TAG}"
	${DOCKER_MAKE_CMD} make compile_pvc-autoresizer_arm64
	${DOCKER_BUILDX_CMD_ARM64} -t ${PVC-AUTORESIZER_IMAGE_NAME}:${IMAGE_TAG} -f ${PVC-AUTORESIZER_IMAGE_DOCKERFILE} ${PROJECT_SOURCE_CODE_DIR}

.PHONY: release_pvc-autoresizer
release_pvc-autoresizer:
    # build for amd64 version
	${DOCKER_MAKE_CMD} make compile_pvc-autoresizer
	${DOCKER_BUILDX_CMD_AMD64} -t ${PVC-AUTORESIZER_IMAGE_NAME}:${RELEASE_TAG}-amd64 -f ${PVC-AUTORESIZER_IMAGE_DOCKERFILE} ${PROJECT_SOURCE_CODE_DIR}
	# build for arm64 version
	${DOCKER_MAKE_CMD} make compile_pvc-autoresizer_arm64
	${DOCKER_BUILDX_CMD_ARM64} -t ${PVC-AUTORESIZER_IMAGE_NAME}:${RELEASE_TAG}-arm64 -f ${PVC-AUTORESIZER_IMAGE_DOCKERFILE} ${PROJECT_SOURCE_CODE_DIR}
	# push to a public registry
	${MUILT_ARCH_PUSH_CMD} -i ${PVC-AUTORESIZER_IMAGE_NAME}:${RELEASE_TAG}


#### for LocalDiskAction controller ##########
LDA_CONTROLLER_MODULE_NAME = local-disk-action-controller
LDA_CONTROLLER_BUILD_INPUT = ${CMDS_DIR}/${LDA_CONTROLLER_MODULE_NAME}/main.go

.PHONY: compile_lda
compile_lda:
	GOARCH=amd64 ${BUILD_ENVS} ${BUILD_CMD} ${BUILD_OPTIONS} -o ${LDA_CONTROLLER_BUILD_OUTPUT} ${LDA_CONTROLLER_BUILD_INPUT}

.PHONY: compile_lda_arm64
compile_lda_arm64:
	GOARCH=arm64 ${BUILD_ENVS} ${BUILD_CMD} ${BUILD_OPTIONS} -o ${LDA_CONTROLLER_BUILD_OUTPUT} ${LDA_CONTROLLER_BUILD_INPUT}

.PHONY: build_lda_image
build_lda_image:
	@echo "Build local-disk-action-controller image ${LDA_CONTROLLER_IMAGE_NAME}:${IMAGE_TAG}"
	${DOCKER_MAKE_CMD} make compile_lda
	docker build -t ${LDA_CONTROLLER_IMAGE_NAME}:${IMAGE_TAG} -f ${LDA_CONTROLLER_IMAGE_DOCKERFILE} ${PROJECT_SOURCE_CODE_DIR}

.PHONY: build_lda_image_arm64
build_lda_image_arm64:
	@echo "Build local-disk-action-controller image ${LDA_CONTROLLER_IMAGE_NAME}:${IMAGE_TAG}"
	${DOCKER_MAKE_CMD} make compile_lda_arm64
	${DOCKER_BUILDX_CMD_ARM64} -t ${LDA_CONTROLLER_IMAGE_NAME}:${IMAGE_TAG} -f ${LDA_CONTROLLER_IMAGE_DOCKERFILE} ${PROJECT_SOURCE_CODE_DIR}

.PHONY: release_lda
release_lda:
	# build for amd64 version
	${DOCKER_MAKE_CMD} make compile_lda
	${DOCKER_BUILDX_CMD_AMD64} -t ${LDA_CONTROLLER_IMAGE_NAME}:${RELEASE_TAG}-amd64 -f ${LDA_CONTROLLER_IMAGE_DOCKERFILE} ${PROJECT_SOURCE_CODE_DIR}
	# build for arm64 version
	${DOCKER_MAKE_CMD} make compile_lda_arm64
	${DOCKER_BUILDX_CMD_ARM64} -t ${LDA_CONTROLLER_IMAGE_NAME}:${RELEASE_TAG}-arm64 -f ${LDA_CONTROLLER_IMAGE_DOCKERFILE} ${PROJECT_SOURCE_CODE_DIR}
	# push to a public registry
	${MUILT_ARCH_PUSH_CMD} -i ${LDA_CONTROLLER_IMAGE_NAME}:${RELEASE_TAG}

### for hwameictl ###
HWAMEICTL_BUILD_INPUT = ${CMDS_DIR}/hwameictl/hwameictl.go
# Example:
# make build_hwameictl OS=linux ARCH=amd64
# make build_hwameictl OS=darwin ARCH=arm64
.PHONY: build_hwameictl
build_hwameictl:
	@echo "Building for OS: ${OS}, ARCH: ${ARCH}"
ifeq ("$(OS)", "windows")
	CGO_ENABLED=0 GOOS=${OS} GOARCH=${ARCH} ${BUILD_CMD} -o "_build/hwameictl/hwameictl-${OS}-${ARCH}.exe" ${BUILD_OPTIONS} ${HWAMEICTL_BUILD_INPUT}
else
	CGO_ENABLED=0 GOOS=${OS} GOARCH=${ARCH} ${BUILD_CMD} -o "_build/hwameictl/hwameictl-${OS}-${ARCH}" ${BUILD_OPTIONS} ${HWAMEICTL_BUILD_INPUT}
endif

.PHONY: apis
apis:
	${DOCKER_MAKE_CMD} make _gen-apis

.PHONY: builder
builder:
	docker build -t ${BUILDER_NAME}:${BUILDER_TAG} -f ${BUILDER_DOCKERFILE} ${PROJECT_SOURCE_CODE_DIR}
	docker push ${BUILDER_NAME}:${BUILDER_TAG}

.PHONY: tools
tools: juicesync

.PHONY: juicesync
juicesync:
	${DOCKER_BUILDX_CMD_AMD64} -t ${JUICESYNC_NAME}:${JUICESYNC_TAG}-amd64 -f ${JUICESYNC_DOCKERFILE}.amd64 ${PROJECT_SOURCE_CODE_DIR}
	# build for arm64 version
	${DOCKER_BUILDX_CMD_ARM64} -t ${JUICESYNC_NAME}:${JUICESYNC_TAG}-arm64 -f ${JUICESYNC_DOCKERFILE}.arm64 ${PROJECT_SOURCE_CODE_DIR}
	# push to a public registry
	${MUILT_ARCH_PUSH_CMD} -i ${JUICESYNC_NAME}:${JUICESYNC_TAG}


.PHONY: _gen-apis
_gen-apis:
	${OPERATOR_CMD} generate k8s
	${OPERATOR_CMD} generate crds
	GOPROXY=https://goproxy.cn,direct /code-generator/generate-groups.sh all github.com/hwameistor/hwameistor/pkg/apis/client github.com/hwameistor/hwameistor/pkg/apis "hwameistor:v1alpha1" --go-header-file /go/src/github.com/hwameistor/hwameistor/build/boilerplate.go.txt

.PHONY: e2e-test
e2e-test:
	bash test/e2e-test.sh

.PHONY: pr-test
pr-test:
	bash test/pr-test.sh


.PHONY: relok8s-test
relok8s-test:
	bash test/relok8s-test.sh

.PHONY: shutdownvm
shutdownvm:
	bash test/shutdownvm.sh

.PHONY: render-chart-values
render-chart-values:
	${RENDER_CHART_VALUES}

.PHONY: lint
lint: golangci-lint
	$(GOLANGLINT_BIN) run

.PHONY: lint-fix
lint-fix: golangci-lint
	$(GOLANGLINT_BIN) run --fix

.PHONY: golangci-lint
golangci-lint:
ifeq (, $(shell command -v golangci-lint))
	@echo "Installing golangci-lint"
	GO111MODULE=on go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.50
GOLANGLINT_BIN=$(shell go env GOPATH)/bin/golangci-lint
else
GOLANGLINT_BIN=$(shell which golangci-lint)
endif
