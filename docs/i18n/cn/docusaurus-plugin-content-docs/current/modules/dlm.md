---
sidebar_position: 2
sidebar_label: "数据加载管理器"
---

# 数据加载管理器

数据加载管理器是DataStor的一个模块，DataStor是AI场景下的云原生本地存储系统加速解决方案，结合p2p技术，提供快速加载远程数据的能力。

## 适用场景

数据加载管理器支持多种数据加载协议：s3、nfs、ftp、http、ssh

在AI数据训练场景中，可以更快的将数据加载到本地缓存卷中。
特别是当数据集支持s3协议拉取时，可以结合p2p技术，大幅提升数据加载速度。
## 与 数据集管理器 一起使用

DataSet Manager 是 HwameiStor 的一个组件，必须与[数据集管理器](./dsm.md) 模块配合使用.

