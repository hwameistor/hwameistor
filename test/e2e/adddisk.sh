#!/bin/bash
# simple scripts mng machine
# link hosts
export GOVC_INSECURE=1
export hosts="fupan-e2e-k8s-node1"
export snapshot="begin-405"
# for h in hosts; do govc vm.power -off -force $h; done
# for h in hosts; do govc snapshot.revert -vm $h "机器配置2"; done
# for h in hosts; do govc vm.power -on -force $h; done
# govc vm.info $hosts[0].Power state
# govc find . -type m -runtime.powerState poweredOn
# govc find . -type m -runtime.powerState poweredOn | xargs govc vm.info
# govc vm.info $hosts
for h in $hosts; do
  ##创建硬盘
  govc vm.disk.create -vm $h -name $h/newdisk -ds="esxi-d01-04-datastore" -size 10G
  ##查看硬盘序号
  #govc device.ls -vm $h
  ##删除硬盘
  #govc device.remove -vm $h -keep=false disk-1000-2
  ##修改磁盘大小
  #govc vm.disk.change -vm $h -disk.name "disk-1000-2" -size 12G
done
