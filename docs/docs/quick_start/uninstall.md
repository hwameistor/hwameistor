---
sidebar_position: 4
sidebar_label: "Uninstall"
---

# Uninstallation

This page describes two schemes to uninstall your HwameiStor.

## To retain data volumes

If you want to retain your data volumes while uninstalling HwameiStor, perform the following steps:

1. Clean up helm instances

   1. Delete HwameiStor helm instance

      ```bash
      helm delete -n hwameistor hwameistor
      ```

   1. Delete DRBD helm instance

      ```bash
      helm delete -n hwameistor drbd-adapter
      ```

1. Clean up HwameiStor components

   1. Delete hwameistor namespace

      ```bash
      kubectl delete ns hwameistor
      ```

   1. Delete LocalVolumeGroup instances

      ```bash
      kubectl delete localvolumegroups.hwameistor.io --all
      ```

   1. Delete CRD, Hook, and RBAC

      ```bash
      kubectl get crd,mutatingwebhookconfiguration,clusterrolebinding,clusterrole -o name \
        | grep hwameistor \
        | xargs -t kubectl delete
      ```

## To delete data volumes

:::danger
Before you start to perform actions, make sure you reallly want to delete all your data.
:::

If you confirm to delete your data volumes and uninstall HwameiStor, perform the following steps:

1. Clean up stateful applications

   1. Delete stateful applications

   1. Delete PVCs

      The relevant PVs, LVs, LVRs, LVGs will also been deleted.

1. Clean up helm instances

   1. Delete HwameiStor helm instance

      ```bash
      helm delete -n hwameistor hwameistor
      ```

   1. Delete DRBD helm instance

      ```bash
      helm delete -n hwameistor drbd-adapter
      ```

1. Clean up HwameiStor components

   1. Delete hwameistor namespace

      ```bash
      kubectl delete ns hwameistor
      ```

   1. Delete CRD, Hook, and RBAC

      ```bash
      kubectl get crd,mutatingwebhookconfiguration,clusterrolebinding,clusterrole -o name \
        | grep hwameistor \
        | xargs -t kubectl delete
      ```

   1. Delete StorageClass

      ```bash
      kubectl get sc -o name \
        | grep hwameistor-storage-lvm- \
        | xargs -t kubectl delete
      ```
