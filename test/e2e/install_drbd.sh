#helm repo add drbd-adapter https://hwameistor.io/drbd-adapter/
#helm repo update drbd-adapter
#helm pull drbd-adapter/drbd-adapter --untar
#kubectl taint node --all node-role.kubernetes.io/master-
#helm install drbd-adapter ./drbd-adapter \
#    -n hwameistor --create-namespace \
#    --set imagePullPolicy=Always

echo "drbd is pre-installed"