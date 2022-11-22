helm repo add addon 	https://release.daocloud.io/chartrepo/addon
helm repo update addon
helm pull addon/drbd-adapter --untar
sed -i '30,36d' drbd-adapter/values.yaml
helm install drbd-adapter ./drbd-adapter -n hwameistor --create-namespace --set registry=daocloud.io/daocloud --set deployOnMasters=yes