---
sidebar_position: 7
sidebar_label: "PV and PVC"
---

# PV and PVC

The PersistentVolume subsystem provides an API for users and administrators that
abstracts details of how storage is provided from how it is consumed. To do this,
we introduce two new API resources: PersistentVolume (PV) and PersistentVolumeClaim (PVC).
A PersistentVolume (PV) is a piece of storage in the cluster that has been provisioned by
an administrator or dynamically provisioned using Storage Classes. It is a resource in
the cluster just like a node is a cluster resource. PVs are volume plugins like Volumes,
but have a lifecycle independent of any individual Pod that uses the PV. This API object
captures the details of the implementation of the storage, be that NFS, iSCSI, or a
cloud-provider-specific storage system.

A PersistentVolumeClaim (PVC) is a request for storage by a user. It is similar to a Pod.
Pods consume node resources and PVCs consume PV resources. Pods can request specific levels
of resources (CPU and Memory). Claims can request specific size and access modes
(e.g., they can be mounted ReadWriteOnce, ReadOnlyMany or ReadWriteMany, see AccessModes).

While PersistentVolumeClaims allow a user to consume abstract storage resources, it is
common that users need PersistentVolumes with varying properties, such as performance,
for different problems. Cluster administrators need to be able to offer a variety of
PersistentVolumes that differ in more ways than size and access modes, without exposing
users to the details of how those volumes are implemented. For these needs, there is
the StorageClass resource. It is used to mark storage resources and performance, and
dynamically provision appropriate PV resources based on PVC demand. After the mechanism
of StorageClass and dynamic provisioning developed for storage resources, the on-demand
creation of volumes is realized, which is an important step in the automatic management
process of shared storage.

See also the official documentation provided by Kubernetes:

- [Persistent Volume](https://kubernetes.io/docs/concepts/storage/persistent-volumes/)
- [StorageClass](https://kubernetes.io/docs/concepts/storage/storage-classes/)
- [Dynamic Volume Provisioning](https://kubernetes.io/docs/concepts/storage/dynamic-provisioning/)
