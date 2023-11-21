---
sidebar_position: 11
sidebar_label: "APIs"
---

# APIs

## CRD Object Classes

HwameiStor defines some object classes to associate PV/PVC with local disks.

| Name                               | Abbr                       | Kind                              | Function                                                             |
|------------------------------------|----------------------------|-----------------------------------|----------------------------------------------------------------------|
| clusters                           | hmcluster                  | Cluster                           | HwameiStor cluster                                                   |
| events                             | evt                        | Event                             | Audit information of HwameiStor cluster                              |
| localdiskclaims                    | ldc                        | LocalDiskClaim                    | Filter and allocate local data disks                                 |
| localdisknodes                     | ldn                        | LocalDiskNode                     | Storage pool for disk volumes                                        |
| localdisks                         | ld                         | LocalDisk                         | Data disks on nodes and automatically find which disks are available |
| localdiskvolumes                   | ldv                        | LocalDiskVolume                   | Disk volumes                                                         |
| localstoragenodes                  | lsn                        | LocalStorageNode                  | Storage pool for lvm volumes                                         |
| localvolumeconverts                | lvconvert                  | LocalVolumeConvert                | Convert common LVM volume to highly available LVM volume             |
| localvolumeexpands                 | lvexpand                   | LocalVolumeExpand                 | Expand local volume storage capacity                                 |                                                        |
| localvolumegroups                  | lvg                        | LocalVolumeGroup                  | LVM volume groups                                                    |                                                          |
| localvolumemigrates                | lvmigrate                  | LocalVolumeMigrate                | Migrate LVM volume                                                   |
| localvolumereplicas                | lvr                        | LocalVolumeReplica                | Replicas of LVM volume                                               |
| localvolumereplicasnapshotrestores | lvrsrestore,lvrsnaprestore | LocalVolumeReplicaSnapshotRestore | Restore snapshots of LVM volume Replicas                             |
| localvolumereplicasnapshots        | lvrs                       | LocalVolumeReplicaSnapshot        | Snapshots of LVM volume Replicas                                     |
| localvolumes                       | lv                         | LocalVolume                       | LVM local volumes                                                    |
| localvolumesnapshotrestores        | lvsrestore,lvsnaprestore   | LocalVolumeSnapshotRestore        | Restore snapshots of LVM volume                                      |
| localvolumesnapshots               | lvs                        | LocalVolumeSnapshot               | Snapshots of LVM volume                                              |                                                      |
| resizepolicies                     |                            | ResizePolicy                      | PVC automatic expansion policy                                       |                      |


