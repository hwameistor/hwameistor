---
sidebar_position: 3
sidebar_label: "通过 hwameistor-operator 安装"
---

# 通过 hwameistor-operator 安装

hwameistor-operator 用于自动化管理和安装各个 HwameiStor 组件。

- HwameiStor 组件的全生命周期管理 (LCM)：
  - Apiserver
  - LocalStorage
  - LocalDiskManager
  - Scheduler
  - AdmissionController
  - VolumeEvictor
  - Exporter
- 自动化本地磁盘声明确保 HwameiStor 准备就绪
- 管理 HwameiStor 卷验证所用的准入控制配置

## 安装步骤

1. 安装 hwameistor-operator

   ```bash
   helm repo add hwameistor-operator https://hwameistor.io/hwameistor-operator
   ```

2. 使用 hwameistor-operator 安装 HwameiStor

   ```bash
   helm repo update hwameistor-operator
   helm install hwameistor-operator hwameistor-operator/hwameistor-operator
   kubectl apply -f https://raw.githubusercontent.com/hwameistor/hwameistor-operator/main/config/samples/hwameistor.io_hmcluster.yaml
   ```
