---
sidebar_position: 4
sidebar_label: "Scheduler"
---

# Scheduler

The Scheduler is used to automatically schedule the Pod to the correct node which is associated with the HwameiStor volume.
With the scheduler, the Pod does not need the NodeAffinity or NodeSelector field to select the node. A scheduler will work for both LVM and Disk volumes.

The Scheduler should be deployed with the HA mode in the cluster, which is a best practice for production.
