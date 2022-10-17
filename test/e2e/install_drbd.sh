helm repo add hwameistor2 https://release.daocloud.io/chartrepo/system
helm repo update hwameistor2
helm pull hwameistor2/drbd-adapter --untar
helm install drbd-adapter ./drbd-adapter -n hwameistor --create-namespace --set registry=daocloud.io/daocloud