#! /usr/bin/env bash
# simple scripts mng machine
# link hosts

ssh root@172.30.40.239 << reallssh
virsh destroy   fupan-e2e
virsh undefine  fupan-e2e --nvram
rm -f  /kvm/fupan-e2e-1.img  /kvm/fupan-e2e-2.img
virt-clone   --original   fupan-e2e-moban  --name fupan-e2e --file /kvm/fupan-e2e-1.img --file /kvm/fupan-e2e-2.img
virsh start fupan-e2e
exit
reallssh
kubectl config use-context k8sarm

