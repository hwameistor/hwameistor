---
sidebar_position: 3
sidebar_label: "CSI"
---

# CSI

CSI is the abbreviation of Container Storage Interface. To have a better understanding
of what we're going to do, the first thing we need to know is what the Container
Storage Interface is. Currently, there are still some problems for already existing
storage subsystem within Kubernetes. Storage driver code is maintained in the Kubernetes
core repository which is difficult to test. But beyond that, Kubernetes needs to give
permissions to storage vendors to check code into the Kubernetes core repository.
Ideally, that should be implemented externally.

CSI is designed to define an industry standard that will enable storage providers
who enable CSI to be available across container orchestration systems that support CSI.

The figure below shows a kind of high-level Kubernetes archetypes integrated with CSI.

![CSI](../img/csi.png)

- Three new external components are introduced to decouple Kubernetes and Storage Provider logic
- Blue arrows present the conventional way to call against API Server
- Red arrows present gRPC to call against Volume Driver

## Extend CSI and Kubernetes

In order to enable the feature of expanding volume atop Kubernetes, we should extend several
components including CSI specification, “in-tree” volume plugin, external-provisioner and external-attacher.

## Extend CSI spec

The feature of expanding volume is still undefined in latest CSI 0.2.0. The new 3 RPCs,
including `RequiresFSResize`, `ControllerResizeVolume` and `NodeResizeVolume`, should be introduced.

```jade
service Controller {
  rpc CreateVolume (CreateVolumeRequest)
    returns (CreateVolumeResponse) {}
……
  rpc RequiresFSResize (RequiresFSResizeRequest)
    returns (RequiresFSResizeResponse) {}
  rpc ControllerResizeVolume (ControllerResizeVolumeRequest)
    returns (ControllerResizeVolumeResponse) {}
}
service Node {
  rpc NodeStageVolume (NodeStageVolumeRequest)
    returns (NodeStageVolumeResponse) {}
……
  rpc NodeResizeVolume (NodeResizeVolumeRequest)
    returns (NodeResizeVolumeResponse) {}
}
```

## Extend “In-Tree” Volume Plugin

In addition to the extend CSI specification, the `csiPlugin` interface within Kubernetes
should also implement `expandablePlugin`. The `csiPlugin` interface will expand
`PersistentVolumeClaim` representing for `ExpanderController`.

```jade
type ExpandableVolumePlugin interface {
VolumePlugin
ExpandVolumeDevice(spec Spec, newSize resource.Quantity, oldSize resource.Quantity) (resource.Quantity, error)
RequiresFSResize() bool
}
```

### Implement Volume Driver

Finally, to abstract complexity of the implementation, we should hard code the separate
storage provider management logic into the following functions which is well-defined in the CSI specification:

- CreateVolume
- DeleteVolume
- ControllerPublishVolume
- ControllerUnpublishVolume
- ValidateVolumeCapabilities
- ListVolumes
- GetCapacity
- ControllerGetCapabilities
- RequiresFSResize
- ControllerResizeVolume

## Demonstration

Let’s demonstrate this feature with a concrete user case.

- Create storage class for CSI storage provisioner

  ```yaml
  allowVolumeExpansion: true
  apiVersion: storage.k8s.io/v1
  kind: StorageClass
  metadata:
    name: csi-qcfs
  parameters:
    csiProvisionerSecretName: orain-test
    csiProvisionerSecretNamespace: default
  provisioner: csi-qcfsplugin
  reclaimPolicy: Delete
  volumeBindingMode: Immediate
  ```

- Deploy CSI Volume Driver including storage provisioner `csi-qcfsplugin` across Kubernetes cluster
- Create PVC `qcfs-pvc` which will be dynamically provisioned by storage class `csi-qcfs`

  ```yaml
  apiVersion: v1
  kind: PersistentVolumeClaim
  metadata:
    name: qcfs-pvc
    namespace: default
  ....
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 300Gi
    storageClassName: csi-qcfs
  ```

- Create MySQL 5.7 instance to use PVC `qcfs-pvc`
- In order to mirror the exact same production-level scenario, there are actually two different types of workloads including:
  - Batch insert to make MySQL consuming more file system capacity
  - Surge query request
- Dynamically expand volume capacity through edit pvc `qcfs-pvc` configuration
