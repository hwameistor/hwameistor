helm repo add hwameistor http://hwameistor.io/hwameistor
helm repo update hwameistor
helm pull hwameistor/drbd-adapter --untar
helm install drbd-adapter ./drbd-adapter -n hwameistor --create-namespace --set registry=daocloud.io/daocloud