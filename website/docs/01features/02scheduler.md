---
sidebar_position: 3
sidebar_label: "Scheduler"
---

# Scheduler

The Scheduler is one of important components of HwameiStor. It is used to automatically schedule the Pod to a correct node which has the associated HwameiStor volume. With the scheduler, the Pod doesn't have to has the NodeAffinity or NodeSelector field to select the node. A scheduler will work for both LVM and Disk volumes.

The Scheduler should be deployed with the HA mode in the cluster, which is a best practice for the production.

**Install by Helm Chart**

Scheduler must work with Local Storage and Local Disk Manager. It's suggested to install by [Helm Chart](../02installation/01helm-chart.md).

**Install by YAML (for developing)**

```bash
$ kubectl apply -f deploy/scheduler.yaml
```
