#! /usr/bin/env bash

export GOVC_INSECURE=1
export GOVC_RESOURCE_POOL="fupan-k8s"
export hosts="adaptation-master adaptation-node1 adaptation-node2 fupan-ad-c81 fupan-ad-u2204 fupan-ad-offline"

for h in $hosts; do
  if [[ `govc vm.info $h | grep poweredOn | wc -l` -eq 1 ]]; then
    govc vm.power -off -force $h
    echo -e "\033[35m === $h has been down === \033[0m"
  fi

done


