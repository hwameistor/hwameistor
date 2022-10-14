#! /usr/bin/env bash
set -x
set -e
# git clone https://github.com/hwameistor/hwameistor.git test/hwameistor

# common defines
date=$(date +%Y%m%d%H%M)
IMAGE_TAG=v${date}
export IMAGE_TAG=${IMAGE_TAG}
MODULES=(local-storage local-disk-manager scheduler admission evictor)

function build_image(){
	echo "Build hwameistor image"
	export IMAGE_TAG=${IMAGE_TAG} && make image
	
	for module in ${MODULES[@]}
	do
		docker push ${IMAGE_REGISTRY}/${module}:${IMAGE_TAG}
	done
}

function prepare_install_params() {
	# FIXME: image tags should be passed by helm install params
	sed -i '/.*ghcr.io*/c\ \ hwameistorImageRegistry: '$ImageRegistry'' helm/hwameistor/values.yaml
	
	# sed -i '/hwameistor\/local-disk-manager/{n;d}' helm/hwameistor/values.yaml
	 sed -i "/hwameistor\/local-disk-manager/a \ \ \ \ tag: ${IMAGE_TAG}" helm/hwameistor/values.yaml
	
	# sed -i '/local-storage/{n;d}' helm/hwameistor/values.yaml
	 sed -i "/local-storage/a \ \ \ \ tag: ${IMAGE_TAG}" helm/hwameistor/values.yaml

	# sed -i '/hwameistor\/scheduler/{n;d}' helm/hwameistor/values.yaml
	 sed -i "/hwameistor\/scheduler/a \ \ tag: ${IMAGE_TAG}" helm/hwameistor/values.yaml

	 sed -i "/hwameistor\/admission/a \ \ tag: ${IMAGE_TAG}" helm/hwameistor/values.yaml

	 sed -i "/hwameistor\/evictor/a \ \ tag: ${IMAGE_TAG}" helm/hwameistor/values.yaml
}

# Step1: build all images tagged with <image_registry>/<module>:<date>
build_image

# Step2: prepare install params included image tag or other install options
prepare_install_params

# Step3: go e2e test
ginkgo -timeout=3h --fail-fast  --label-filter=${E2E_TESTING_LEVEL} test/e2e
