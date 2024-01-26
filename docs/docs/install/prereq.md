---
sidebar_position: 1
sidebar_label: "Prerequisites"
---

# Prerequisites

## Kubernetes

- Kubernetes `1.18+`
- CoreDNS is deployed

### Unsupported platforms

- OpenShift
- Rancher

:::note
The aforementioned platforms are currently unsupported; however,
they will be supported in the future.
:::

## Hosts

### Supported linux distributions

- CentOS/RHEL `7.4+`
- Rocky Linux `8.4+`
- Ubuntu `18+`
- Kylin `V10`

### Supported processors

- x86_64
- ARM64

### Required dependencies

- `LVM2` is installed.
- For HA features, it is necessary to install `kernel-devel` with a compatible version to the current `kernel`.
- The HA features module cannot be automatically installed on nodes with certain kernel versions and needs to be installed manually.

  <details>
  <summary>Click to view all compatible kernel versions</summary>

  ```text
  5.8.0-1043-azure
  5.8.0-1042-azure
  5.8.0-1041-azure
  5.4.17-2102.205.7.2.el7uek
  5.4.17-2011.0.7.el8uek
  5.4.0-91
  5.4.0-90
  5.4.0-89
  5.4.0-88
  5.4.0-86
  5.4.0-84
  5.4.0-1064-azure
  5.4.0-1063-azure
  5.4.0-1062-azure
  5.4.0-1061-azure
  5.4.0-1060-aws
  5.4.0-1059-azure
  5.4.0-1059-aws
  5.4.0-1058-azure
  5.4.0-1058-aws
  5.4.0-1057-aws
  5.4.0-1056-aws
  5.4.0-1055-aws
  5.4.247-1.el7
  5.3.18-57.3
  5.3.18-22.2
  5.14.0-1.7.1.el9
  5.11.0-1022-azure
  5.11.0-1022-aws
  5.11.0-1021-azure
  5.11.0-1021-aws
  5.11.0-1020-azure
  5.11.0-1020-aws
  5.11.0-1019-aws
  5.11.0-1017-aws
  5.11.0-1016-aws
  5.10.0-8
  5.10.0-7
  5.10.0-6
  4.9.215-36.el7
  4.9.212-36.el7
  4.9.206-36.el7
  4.9.199-35.el7
  4.9.188-35.el7
  4.4.92-6.30.1
  4.4.74-92.38.1
  4.4.52-2.1
  4.4.27-572.565306
  4.4.0-217
  4.4.0-216
  4.4.0-214
  4.4.0-213
  4.4.0-210
  4.4.0-1133-aws
  4.4.0-1132-aws
  4.4.0-1131-aws
  4.4.0-1128-aws
  4.4.0-1121-aws
  4.4.0-1118-aws
  4.19.19-5.0.8
  4.19.0-8
  4.19.0-6
  4.19.0-5
  4.19.0-16
  4.18.0-80.1.2.el8_0
  4.18.0-348.el8
  4.18.0-305.el8
  4.18.0-240.1.1.el8_3
  4.18.0-193.el8
  4.18.0-147.el8
  4.15.0-163
  4.15.0-162
  4.15.0-161
  4.15.0-159
  4.15.0-158
  4.15.0-156
  4.15.0-112-lowlatency
  4.15.0-1113-azure
  4.15.0-1040-azure
  4.15.0-1036-azure
  4.14.35-2047.502.5.el7uek
  4.14.35-1902.4.8.el7uek
  4.14.35-1818.3.3.el7uek
  4.14.248-189.473.amzn2
  4.14.128-112.105.amzn2
  4.13.0-1018-azure
  4.12.14-95.3.1
  4.12.14-25.25.1
  4.12.14-197.29
  4.12.14-120.1
  4.1.12-124.49.3.1.el7uek
  4.1.12-124.26.3.el6uek
  4.1.12-124.21.1.el6uek
  3.10.0-957.el7
  3.10.0-862.el7
  3.10.0-693.el7
  3.10.0-693.21.1.el7
  3.10.0-693.17.1.el7
  3.10.0-514.6.2.el7
  3.10.0-514.36.5.el7
  3.10.0-327.el7
  3.10.0-229.1.2.el7
  3.10.0-123.20.1.el7
  3.10.0-1160.el7
  3.10.0-1127.el7
  3.10.0-1062.el7
  3.10.0-1049.el7
  3.0.101-108.13.1
  2.6.32-754.el6
  2.6.32-696.el6
  2.6.32-696.30.1.el6
  2.6.32-696.23.1.el6
  2.6.32-642.1.1.el6
  2.6.32-573.1.1.el6
  2.6.32-504.el6
  ```

  </details>

- For VolumeResize features, a tool to resize the filesystem is required.
  By default, `xfs` is used as the volume filesystem. Therefore, you need to install `xfs_growfs` on the host.

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

<Tabs>
<TabItem value="centos" label="CentOS/RHEL, Rocky, and Kylin">

```bash
yum install -y lvm2
yum install -y kernel-devel-$(uname -r)
yum install -y xfsprogs
```

</TabItem>
<TabItem value="ubuntu" label="Ubuntu">

```bash
apt-get install -y lvm2
apt-get install -y linux-headers-$(uname -r)
apt-get install -y xfsprogs
```

</TabItem>
</Tabs>

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

:::note
For virtual machine environments, please ensure that the disk serial number feature is enabled for each virtual machine. This will assist HwameiStor in better identifying and managing the disks on the host.

To avoid disk identification conflicts, please ensure that the provided virtual disk serial numbers are not duplicated.
:::

### Network

For production, it is recommended to have a redundant `10Giga TCP/IP` network, if the HA feature is enabled.
