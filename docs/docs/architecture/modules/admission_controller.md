---
sidebar_position: 5
sidebar_label: "Admission Controller"
---

# admission-controller

`admission-controller` is a webhook that can automatically verify which pod uses the HwameiStor volume and help to modify the schedulerName to hwameistor-scheduler. For the specific principle, refer to [K8S Dynamic Admission Control](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/).

## How to identify a HwameiStor volume?

`admission-controller` gets all the PVCs used by a pod, and checks the [provisioner](https://kubernetes.io/docs/concepts/storage/storage-classes/) of each PVC in turn. If the suffix of the provisioner name is `*.hwameistor.io`, it is believed that the pod is using the volume provided by HwameiStor.

## Which resources will be verified?

Only `POD` resources will be verified, and the verification process occurs at the time of creation.

:::info
In order to ensure that the pods of HwameiStor can be started smoothly, the pods in the namespace where HwameiStor is deployed will not be verified.
:::
