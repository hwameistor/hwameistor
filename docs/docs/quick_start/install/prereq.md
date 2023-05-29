---
sidebar_position: 1
sidebar_label: "Prerequisites"
---

# Prerequisites

## Kubernetes

- Kubernetes `1.18+`
- CoreDNS is deployed

### Unsupported platforms

1. OpenShift
2. Rancher

:::note
Above platforms are not supported currently but will be in the future.
:::

## Hosts

### Linux distributions

1. CentOS/RHEL `7.4+`
2. Rocky Linux `8.4+`
3. Ubuntu `18+`
4. Kylin `V10`

### Processor architecture

1. x86_64
2. ARM64

### Package dependencies

1. `LVM2` is installed.
2. For HA features, `kernel-devel` shall be installed and has a compatible version with the current `kernel`.
3. For VolumeResize features, a tool to resize the filesystem is required. 
   By default, `xfs` is used as the volume filesystem. Therefore, you need to install `xfs_growfs` on the host.


```console title="CentOS/RHEL, Rocky and Kylin"
$ yum install -y lvm2
$ yum install -y kernel-devel-$(uname -r)
$ yum install -y xfsprogs
```

```console title="Ubuntu"
$ apt-get install -y lvm2
$ apt-get install -y linux-headers-$(uname -r)
$ apt-get install -y xfsprogs
```

### Secure Boot

The HA feature does not support `Secure Boot` currently. Make sure `Secure Boot` is `disabled`：

```console
$ mokutil --sb-state
SecureBoot disabled

$ dmesg | grep secureboot
[    0.000000] secureboot: Secure boot disabled
```

### Data disks

HwameiStor supports `HDD`, `SSD`, and `NVMe`.

For test, each host must have at least one unused drive with a minimal size of `10GiB`.

For production, it is recommended to have at least one unused drive, protected by RAID1 or RAID5/6, with a minimal size of `200GiB`.

### Network

For production, it is recommended to have a redundant `10Giga TCP/IP` network, if the HA feature is enabled.
