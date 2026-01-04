echo "install hwameistor"
helm repo add hwameistor https://hwameistor.io/hwameistor
helm repo update hwameistor
helm install hwameistor hwameistor/hwameistor    -n hwameistor --create-namespace --set global.k8sImageRegistry=m.daocloud.io/registry.k8s.io   --set global.hwameistorImageRegistry=ghcr.m.daocloud.io