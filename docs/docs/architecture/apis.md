---
sidebar_position: 4
sidebar_label: "APIs"
---

# APIs

## CRD Object Classes

HwameiStor defines some object classes to associate PV/PVC with local disks.

| Kind              | Abbr     | Function                                                             |
|-------------------|----------|----------------------------------------------------------------------|
| LocalDiskNode     | ldn      | Storage pool for disk volumes                                        |
| LocalDisk         | ld       | Data disks on nodes and automatically find which disks are available |
| LocalDiskVolume   | ldv      | Disk volumes                                                         |
| LocalDiskClaim    | ldc      | Filter and allocate local data disks                                 |
| LocalStorageNode  | lsn      | Storage pool for lvm volumes                                         |
| LocalVolume       | lv       | LVM local volumes                                                    |
| LocalVolumeExpand | lvexpand | Expand local volume storage capacity                                 |
