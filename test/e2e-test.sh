#! /usr/bin/env bash
git clone https://github.com/hwameistor/helm-charts.git test/helm-charts
git clone https://github.com/hwameistor/local-disk-manager.git test/local-disk-manager
cp -r -f test/local-disk-manager/deploy/crds/hwameistor.io_l* test/helm-charts/charts/hwameistor/crds/
cat test/helm-charts/charts/hwameistor/values.yaml | while read line
##
do
    result=$(echo $line | grep "imageRepository")
    if [[ "$result" != "" ]]
    then
        img=${line:17:50}
    fi
    result=$(echo $line | grep "tag")
    if [[ "$result" != "" ]]
    then
        hwamei=$(echo $img | grep "hwameistor")
        if [[ "$hwamei" != "" ]]
        then
            image=$img:${line:5:50}
            echo "docker pull ghcr.io/$image"
            docker pull ghcr.io/$image
            echo "docker tag ghcr.io/$image $ImageRegistry/$image"
            docker tag ghcr.io/$image 10.6.170.180/$image
            echo "docker push $ImageRegistry/$image"
            docker push 10.6.170.180/$image
        fi
    fi
done
##
date=$(date +%Y%m%d%H%M)
docker tag $ImageRegistry/hwameistor/local-disk-manager:99.9-dev $ImageRegistry/hwameistor/local-disk-manager:$date
docker push $ImageRegistry/hwameistor/local-disk-manager:$date

sed -i '/.*ghcr.io*/c\hwameistorImageRegistry: '$ImageRegistry'' test/helm-charts/charts/hwameistor/values.yaml
sed -i '/hwameistor\/local-disk-manager/{n;d}' test/helm-charts/charts/hwameistor/values.yaml
sed -i '/hwameistor\/local-disk-manager/a \ \ \ \ tag: 99.9-dev' test/helm-charts/charts/hwameistor/values.yaml
ginkgo --fail-fast --label-filter=${E2E_TESTING_LEVEL} test/e2e