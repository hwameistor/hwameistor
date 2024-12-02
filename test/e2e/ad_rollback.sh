#! /usr/bin/env bash
# simple scripts mng machine
# link hosts

#!/bin/bash



export hosts="fupan-rocky8.10-ad-10.6.113.120"
export snapshot=$K8S_VERSION

export GOVC_INSECURE=1

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


TARGET_IP="10.6.113.120"
TIMEOUT=300
INTERVAL=5

START_TIME=$(date +%s)

while true; do
    if ping -c 1 "$TARGET_IP" &> /dev/null; then
        echo "successfully pinged $TARGET_IP"
        break
    fi

    CURRENT_TIME=$(date +%s)
    ELAPSED_TIME=$((CURRENT_TIME - START_TIME))

    if [ "$ELAPSED_TIME" -ge "$TIMEOUT" ]; then
        echo "failed to ping $TARGET_IP"
        exit 1
    fi

    sleep "$INTERVAL"
done



sudo rm -rf /home/github/.kube/config
sudo rm -rf /root/.kube/config

sshpass -vp $password scp -rp -o StrictHostKeyChecking=no root@10.6.113.120:/root/.kube/config /home/github/.kube/config
if [ $? -eq 0 ]; then
    echo "Successfully copied .kube/config"
else
    echo "Copy Failure"
fi

sed -i 's|https://127.0.0.1:6443|https://10.6.113.120:6443|g' /home/github/.kube/config


sudo cp -f /home/github/.kube/config /root/.kube/config

