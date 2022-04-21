#! /usr/bin/env bash
# simple scripts mng machine
# link hosts
export GOVC_INSECURE=1
export GOVC_RESOURCE_POOL="e2e"
export hosts="fupan-e2e-k8s-master fupan-e2e-k8s-node1 fupan-e2e-k8s-node2"
export snapshot="e2etest"
# for h in hosts; do govc vm.power -off -force $h; done
# for h in hosts; do govc snapshot.revert -vm $h "机器配置2"; done
# for h in hosts; do govc vm.power -on -force $h; done

# govc vm.info $hosts[0].Power state
# govc find . -type m -runtime.powerState poweredOn
# govc find . -type m -runtime.powerState poweredOn | xargs govc vm.info
# govc vm.info $hosts

for h in $hosts; do
  if [[ `govc vm.info $h | grep poweredOn | wc -l` -eq 1 ]]; then
    govc vm.power -off -force $h
    echo -e "\033[35m === $h has been down === \033[0m"
  fi

  govc snapshot.revert -vm $h $snapshot
  echo -e "\033[35m === $h reverted to snapshot: `govc snapshot.tree -vm $h -C -D -i -d` === \033[0m"

  govc vm.power -on $h
  echo -e "\033[35m === $h: power turned on === \033[0m"
done

echo -e "\033[35m === task will end in 1m 30s === \033[0m"
for i in `seq 1 15`; do
  echo -e "\033[35m === `date  '+%Y-%m-%d %H:%M:%S'` === \033[0m"
  sleep 6s
done
git clone https://github.com/hwameistor/helm-charts.git test/helm-charts
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
            echo "docker tag ghcr.io/$image 10.6.170.180/$image"
            docker tag ghcr.io/$image 10.6.170.180/$image
            echo "docker push 10.6.170.180/$image"
            docker push 10.6.170.180/$image
        fi
    fi
done
##
sed -i '/.*ghcr.io*/c\hwameistorImageRegistry: 10.6.170.180' test/helm-charts/charts/hwameistor/values.yaml
sed -i '/local-storage/{n;d}' test/helm-charts/charts/hwameistor/values.yaml
sed -i '/local-storage/a \ \ \ \ tag: 99.9-dev' test/helm-charts/charts/hwameistor/values.yaml
ginkgo --fail-fast --label-filter=${E2E_TESTING_LEVEL} test/e2e