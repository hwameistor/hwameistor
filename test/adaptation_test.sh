#! /usr/bin/env bash

set -x
set -e
# git clone https://github.com/hwameistor/hwameistor.git test/hwameistor
source /etc/profile || true
# common defines

IMAGE_TAG=v${GITHUB_RUN_ID}
export IMAGE_TAG=${IMAGE_TAG}


function prepare_install_params() {
	# FIXME: image tags should be passed by helm install params
	sed -i '/.*hwameistorImageRegistry: ghcr.io*/c\ \ hwameistorImageRegistry: '$ImageRegistry'' helm/hwameistor/values.yaml
#
#	# sed -i '/hwameistor\/local-disk-manager/{n;d}' helm/hwameistor/values.yaml
#	 sed -i "/hwameistor\/local-disk-manager/a \ \ \ \ tag: ${IMAGE_TAG}" helm/hwameistor/values.yaml
#
#	# sed -i '/local-storage/{n;d}' helm/hwameistor/values.yaml
#	 sed -i "/local-storage/a \ \ \ \ tag: ${IMAGE_TAG}" helm/hwameistor/values.yaml
#
#	# sed -i '/hwameistor\/scheduler/{n;d}' helm/hwameistor/values.yaml
#	 sed -i "/hwameistor\/scheduler/a \ \ tag: ${IMAGE_TAG}" helm/hwameistor/values.yaml
#
#	 sed -i "/hwameistor\/admission/a \ \ tag: ${IMAGE_TAG}" helm/hwameistor/values.yaml
#
#	 sed -i "/hwameistor\/evictor/a \ \ tag: ${IMAGE_TAG}" helm/hwameistor/values.yaml
#
#	 sed -i "/hwameistor\/exporter/a \ \ tag: ${IMAGE_TAG}" helm/hwameistor/values.yaml
#
#   sed -i "/hwameistor\/apiserver/a \ \ tag: ${IMAGE_TAG}" helm/hwameistor/values.yaml

   sed -i "5c version: ${IMAGE_TAG}" helm/hwameistor/Chart.yaml

	 sed -i 's/rclone\/rclone/10.6.112.210\/hwameistor\/hwameistor-migrate-rclone/' helm/hwameistor/values.yaml

	 sed -i 's/tag: 1.53.2/tag: v1.1.2/' helm/hwameistor/values.yaml
}

# Step2: prepare install params included image tag or other install options
prepare_install_params

# Step3: go e2e test
ginkgo -timeout=10h --fail-fast  --label-filter=ad_test test/e2e

