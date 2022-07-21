---
sidebar_position: 2
sidebar_label: "Install by Helm Charts"
---

# Install by Helm Charts

Local Storage is a component of HwameiStor and must work with Local Disk Manager module. It's suggested to install by helm-charts.

This functionality is in alpha and is subject to change. The code is provided as-is with no warranties. Alpha features are not subject to the support SLA of official GA features.  

To install it by helm charts, perform the following procedure.

## Step 1: Install HwameiStor

[Helm](https://helm.sh/) is required to use charts for installation. Refer to Helm's [Documentation](https://helm.sh/docs/) to get started.

```bash
$ git clone https://github.com/hwameistor/helm-charts.git 
$ cd helm-charts/charts 
$ helm install hwameistor -n hwameistor --create-namespace --generate-name
```

Or

```bash
$ helm repo add hwameistor http://hwameistor.io/helm-charts 
$ helm install hwameistor/hwameistor -n hwameistor --create-namespace --generate-name
```

Run `helm search repo hwameistor` to check the charts.

## Step 2: Enable HwameiStor on a node

Once the Helm charts are installed, you can enable HwameiStor on a specific node.

```bash
$ kubectl label node <node-name> "lvm.hwameistor.io/enable=true"
```

## Step 3: Claim disks by types on a node

Then claim disks for your local-storage by applying LocalDiskClaim CR.

```yaml
cat > ./local-disk-claim.yaml <<- EOF
apiVersion: hwameistor.io/v1alpha1
kind: LocalDiskClaim
metadata:
  name: <anyname>
  namespace: hwameistor
spec:
  nodeName: <node-name>
  description:
    diskType: <HDD or SSD or NVMe>
EOF
```

Congratulations! HwameiStor is successfully deployed on your cluster now.