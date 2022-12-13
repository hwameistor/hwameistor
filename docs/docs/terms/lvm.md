---
sidebar_position: 6
sidebar_label: "LVM"
---

# LVM

The full name of LVM is Logical Volume Manager. It adds a logical layer between
the disk partition and the file system, provides an abstract disk volume for the
file system to shield the underlying disk partition layout, and establishes a file
system on the disk volume.

With LVM, you can dynamically resize file systems without repartitioning the disk,
and file systems managed with LVM can span disks. When a new disk is added to the
server, the administrator does not have to move the original files to the new disk,
but directly extends the file system across the disk through LVM. That is, by
encapsulating the underlying physical hard disk, and then presenting it to the
upper-level application in the form of a logical volume.

LVM encapsulates the underlying hard disk. When we operate the underlying physical
hard disk, we no longer operate on the partition, but perform the underlying disk
management operation on it through something called a logical volume.

## Basic functionality

- **Physical media (PM):** LVM storage media can be partitions, disks, RAID arrays, or SAN disks.

- **Physical volume (PV):** Physical volume is the basic storage logical block of LVM,
  but compared with basic physical storage media (such as partitions, disks, etc.),
  it contains management parameters related to LVM. A physical volume can be partitioned
  by a disk, or it can be the disk itself. Disks must be initialized as LVM physical
  volumes to be used with LVM.

- **Volume groups (VG):** It can be resized online by absorbing new physical volumes (PVs)
  or ejecting existing ones.

- **Logical volumes (LV):** It can be resized online by concatenating extents onto them or
  truncating extents from them.

- **Physical extents (PE):** The smallest storage unit that can be allocated in the physical volume.
  The size of PE can be specified, and the default is 4MB.

- **Logical extents (LE):** The smallest storage unit that can be allocated in an logical volume.
  In the same volume group, the size of LE is the same as that of PE, and there is a one-to-one correspondence.

## Advantages

- Use volume groups to make multiple hard drives look like one big hard drive
- Using logical volumes, partitions can span multiple hard disk spaces sdb1 sdb2 sdc1 sdd2 sdf
- Using logical volumes, you can dynamically resize it if the storage space is insufficient
- When resizing a logical volume, you need not to consider the location of the logical volume
  on a hard disk, and you need not to worry about no contiguous space available
- LV and VG can be created, deleted, and resized online, and the file system on LVM also needs to be resized
- You can create snapshots, which can be used to back up file systems
- RAID + LVM combined: LVM is a software method of volume management, while RAID is a method of
  disk management. For important data, RAID is used to protect physical disks from failures and
  services are not interrupted, and LVM is used to achieve a good volume management and better
  use of disk resources.

## Basic procedure to use LVM

1. Format a physical disk as PVs, that is, the space is divided into PEs. A PV contains multiple PEs.
1. Add different PVs to the same VG, that is, the PEs of different PVs all enter the PE pool of the VG.
   A VG contains multiple PVs.
1. Create logical volumes in the VG. This creation process is based on PE, so the PEs that make up the LV
   may come from different physical disks. LV is created based on PE.
1. Directly format the LV and mount it for use.
1. The scaling in / out of an LV is actually to increase or decrease the number of PEs that make up the LV
   without losing the original data.
1. Format the LV and mount it for use.

## LV expansion

First, determine if there is available space for expansion, because space is created
from VG, and LVs cannot be expanded across VGs. If the VG has no capacity, you need to
expand the VG first. Perform the following steps:

```bash
$ vgs
VG #PV #LV #SN Attr VSize VFree
vg-sdb1 1 8 1 wz--n- <16.00g <5.39g
$ lvextend -L +100M  -r /dev/vg-sdb1/lv-sdb1     #将 /dev/vg-sdb1/lv-sdb 扩容 100M
```

## VG expansion

If there is not sufficient space in the VG and you need to add a new disk, run the following commands in sequence:

```bash
pvcreate /dev/sdc
vgextend vg-sdb1 /dev/sdb3
```

## LV snapshots

The LVM mechanism provides the function of snapshotting LVs to obtain a state-consistent
backup of the file system. LVM adopts Copy-On-Write (COW) technology, which can be backed
up without stopping the service or setting the logical volume as read-only. Using the LVM,
snapshot function can enable consistent backup without affecting the availability of the server.

The copy-on-write adopted by LVM means that when creating an LVM snapshot, only the metadata
in the original volume is copied. In other words, when an LVM logical volume is created,
no physical replication of the data occurs. In another words, only the metadata is copied,
not the physical data, so the snapshot creation is almost real-time. When a write operation
is performed on the original volume, the snapshot will track the changes to the blocks in
the original volume. At this time, the data that will be changed on the original volume
will be copied to the space reserved by the snapshot before the change.
