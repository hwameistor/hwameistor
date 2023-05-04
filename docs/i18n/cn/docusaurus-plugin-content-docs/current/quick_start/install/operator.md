---
sidebar_position: 2
sidebar_label: "HwameiStor-Operator 安装"
---

# 通过 HwameiStor-Operator 安装

Hwameistor-Operator 负责自动化安装并管理 HwameiStor 系统。

- HwameiStor 组件的全生命周期管理 (LCM)：
    - LocalDiskManager
    - LocalStorage
    - Scheduler
    - AdmissionController
    - VolumeEvictor
    - Exporter
    - Apiserver
    - Graph UI

- 根据不同目的和用途配置节点磁盘
- 自动发现节点磁盘的类型，并以此自动创建 HwameiStor 存储池
- 根据 HwameiStor 系统的配置和功能自动创建相应的 StorageClass

## 安装步骤

1. 安装 hwameistor-operator

   ```console
   helm repo add hwameistor-operator https://hwameistor.io/hwameistor-operator
   helm repo update hwameistor-operator
   helm install -n hwameistor hwameistor-operator hwameistor-operator/hwameistor-operator
   ```

2. 创建 HwameiStor 存储系统

   ```console
   kubectl -n hwameistor apply -f https://raw.githubusercontent.com/hwameistor/hwameistor-operator/main/config/samples/hwameistor.io_hmcluster.yaml
   ```
