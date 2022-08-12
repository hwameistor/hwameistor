---
sidebar_position: 1
sidebar_label: "Prerequisites"
---

# Prerequisites

## Kubernetes

1. Kubernetes `1.18+`
1. CoreDNS is deployed

## Host

### Linux Distributions

1. CentOS/RHEL `7.4+`
2. Rocky Linux `8.4+`
3. Ubuntu `18+`
4. Kylin `V10`

### Processor Architecture

1. x86_64
1. ARM64

### Package Dependencies

1. `LVM2` is installed
2. For HA features, `kernel-devel` must be installed and match the version of the operating `kernel`

```bash
# CentOS/RHEL, Rocky and Kylin
$ yum install -y lvm2
$ yum install -y kernel-devel-$(uname -r)

# Ubuntu
$ apt-get install -y lvm2
$ apt-get install -y linux-headers-$(uname -r)
```

### Data Disk

HwameiStor supports `HDD`, `SSD` and `NVMe`.

For test, each host must have at least one unused drive with a minimal size of `10GiB`.

For production, it is recommended to have a least one unused drive, protected by RAID1 or RAID5/6, with a minimal size of `200GiB`.

### Network

For production, it is recommended to have a redundant `10Giga TCP/IP` network, if the HA feature is enabled.
