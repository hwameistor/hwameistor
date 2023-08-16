helm repo add addon 	https://release.daocloud.io/chartrepo/addon
helm repo update addon
helm pull addon/drbd-adapter --untar
sed -i '30,36d' drbd-adapter/values.yaml
kubectl taint node --all node-role.kubernetes.io/master-
helm install drbd-adapter ./drbd-adapter -n hwameistor --create-namespace --set registry=10.6.112.210/hwameistor --set deployOnMasters=yes