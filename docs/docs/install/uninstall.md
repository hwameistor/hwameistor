---
sidebar_position: 5
sidebar_label: "Uninstall"
---

# Uninstall (For test purposes only, not for production use)

To ensure data security, it is strongly recommended not to uninstall the HwameiStor system in a production environment.
This section introduces two uninstallation scenarios for testing environments.

## Uninstall but retain data volumes

If you want to uninstall the HwameiStor components, but still keep the existing data volumes working with the applications, perform the following steps:

```console
$ kubectl get cluster.hwameistor.io
NAME             AGE
cluster-sample   21m

$ kubectl delete clusters.hwameistor.io hwameistor-cluster
```

Finally, all the HwameiStor's components (i.e. Pods) will be deleted. Check by:

```bash
kubectl -n hwameistor get pod
```

## Uninstall and delete data volumes

:::danger
Before you start to perform actions, make sure you reallly want to delete all your data.
:::

If you confirm to delete your data volumes and uninstall HwameiStor, perform the following steps:

1. Clean up stateful applications.

   1. Delete stateful applications.

   1. Delete PVCs.

      The relevant PVs, LVs, LVRs, LVGs will also been deleted.

1. Clean up HwameiStor components.

   1. Delete HwameiStor components.

      ```bash
      kubectl delete clusters.hwameistor.io hwameistor-cluster
      ```

   2. Delete hwameistor namespace.

      ```bash
      kubectl delete ns hwameistor
      ```

   3. Delete CRD, Hook, and RBAC.

      ```bash
      kubectl get crd,mutatingwebhookconfiguration,clusterrolebinding,clusterrole -o name \
        | grep hwameistor \
        | xargs -t kubectl delete
      ```

   4. Delete StorageClass.

      ```bash
      kubectl get sc -o name \
        | grep hwameistor-storage-lvm- \
        | xargs -t kubectl delete
      ```

   5. Delete hwameistor-operator.

      ```bash
      helm uninstall hwameistor-operator -n hwameistor
      ```

3. Finally, you still need to clean up the LVM configuration on each node,
   and also data on the disks by tools like [wipefs](https://man7.org/linux/man-pages/man8/wipefs.8.html).

   ```bash
   wipefs -a /dev/sdx
   blkid /dev/sdx
   ```
