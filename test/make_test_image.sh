#! /usr/bin/env bash
set -e
set -x

IMAGE_TAG=v${GITHUB_RUN_ID}
export IMAGE_TAG=${IMAGE_TAG}
MODULES=(local-storage local-disk-manager scheduler admission evictor exporter apiserver failover-assistant auditor pvc-autoresizer local-disk-action-controller)

function build_image(){
	echo "Build hwameistor image"
	export IMAGE_TAG=${IMAGE_TAG} && make image

	for module in ${MODULES[@]}
	do
		docker push ${IMAGE_REGISTRY}/${module}:${IMAGE_TAG}
	done
}

timer_start=`date "+%Y-%m-%d %H:%M:%S"`
build_image
timer_end=`date "+%Y-%m-%d %H:%M:%S"`



