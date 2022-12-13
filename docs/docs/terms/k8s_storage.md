---
sidebar_position: 1
sidebar_label: "K8s Storage"
---

# K8s Storage

Kubernetes has made several enhancements to support running Stateful Workloads
by providing the required abstractions for Platform (or Cluster Administrators)
and Application developers. The abstractions ensure that different types of file
and block storage (whether ephemeral or persistent, local or remote) are available
wherever a container is scheduled (including provisioning/creating, attaching,
mounting, unmounting, detaching, and deleting of volumes), storage capacity management
(container ephemeral storage usage, volume resizing, etc.), influencing scheduling of
containers based on storage (data gravity, availability, etc.), and generic operations
on storage (snapshotting, etc.).

The most important Kubernetes Storage abstractions to be aware of for running Stateful
workloads using HwameiStor are:

- [Kubernetes Storage](#kubernetes-storage)
  - [Container Storage Interface](#container-storage-interface)
  - [Storage Classes and Dynamic Provisioning](#storage-classes-and-dynamic-provisioning)
  - [Persistent Volume Claims](#persistent-volume-claims)
  - [Persistent Volume](#persistent-volume)
  - [StatefulSets and Deployments](#statefulsets-and-deployments)

## Container Storage Interface

The Container Storage Interface (CSI) is a standard for exposing arbitrary block and
file storage systems to containerized workloads on Container Orchestration Systems (COs)
like Kubernetes. Using CSI third-party storage providers like HwameiStor can write and
deploy plugins exposing new storage volumes like HwameiStor Local and Replicated Volumes
in Kubernetes without ever having to touch the core Kubernetes code.

When cluster administrators install HwameiStor, the required HwameiStor CSI driver
components are installed into the Kubernetes cluster.

```csharp
Prior to CSI, Kubernetes supported adding storage providers using out-of-tree provisioners referred to as external provisioners. And Kubernetes in-tree volumes pre-date the external provisioners. There is an ongoing effort in the Kubernetes community to deprecate in-tree volumes with CSI based volumes.
```

## Storage Classes and Dynamic Provisioning

A StorageClass provides a way for administrators to describe the "classes" of storage
they offer. Different classes might map to quality-of-service levels, or to backup
policies, or to arbitrary policies determined by the cluster administrators. This
concept is sometimes called "profiles" in other storage systems.

The dynamic provisioning feature eliminates the need for cluster administrators
to pre-provision storage. Instead, it automatically provisions storage when it
is requested by users. The implementation of dynamic volume provisioning is based
on the StorageClass abstraction. A cluster administrator can define as many
StorageClass objects as needed, each specifying a volume plugin (aka provisioner)
that provisions a volume and the set of parameters to pass to that provisioner when provisioning.

A cluster administrator can define and expose multiple flavors of storage
(from the same or different storage systems) within a cluster, each with a custom
set of parameters. This design also ensures that end users don't have to worry about
the complexity and nuances of how storage is provisioned, but still have the ability
to select from multiple storage options.

When HwameiStor is installed, it ships with a couple of default storage classes that
allow users to create either local (HwameiStor LocalVolume) or replicated
(HwameiStor LocalVolumeReplica) volumes. The cluster administrator can enable the
required storage engines and then create Storage Classes for the required Data Engines.

## Persistent Volume Claims

PersistentVolumeClaim (PVC) is a user’s storage request that is served by a StorageClass
offered by the cluster administrator. An application running on a container can request
a certain type of storage. For example, a container can specify the size of storage it
needs or the way it needs to access the data (read only, read/write, read-write many, etc.,).

Beyond storage size and access mode, administrators create Storage Classes to provided PVs
with custom properties, such as the type of disk (HDD vs. SSD), the level of performance,
or the storage tier (regular or cold storage).

## Persistent Volume

The PersistentVolume(PV) is dynamically provisioned by the storage providers when users
request for a PVC. PV contains the details on how the storage can be consumed by the
container. Kubernetes and the Volume Drivers use the details in the PV to attach/detach
the storage to the node where the container is running and mount/unmount storage to a container.

HwameiStor Control Plane dynamically provisions HwameiStor Local and Replicated volumes
and helps in creating the PV objects in the cluster.

## StatefulSets and Deployments

Kubernetes provides several built-in workload resources such as StatefulSets and Deployments
that let application developers define an application running on Kubernetes. You can run a
stateful application by creating a Kubernetes Deployment/Statefulset and connecting it to
a PersistentVolume using a PersistentVolumeClaim.

For example, you can create a MySQL Deployment YAML that references a PersistentVolumeClaim.
The MySQL PersistentVolumeClaim referenced by the Deployment should be created with the
requested size and StorageClass. Once the HwameiStor control plane provisions a PersistenceVolume
for the required StorageClass and requested capacity, the claim is set as satisfied. Kubernetes
will then mount the PersistentVolume and launch the MySQL Deployment.
