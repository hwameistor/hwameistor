---
slug: minio
title: HwameiStor 对 Minio 的支持
authors: [Simon]
tags: [Test]
---

# HwameiStor 对 Minio 的支持

1. MinIO简介
   1. 功能特性

MinIO是一款高性能，分布式，兼容S3的多云对象存储系统套件。由于原生支持kubernetes，MinIO支持所有的公有云，私有云及边缘计算环境。 MinIO是GNU AGPL v3开源的软件定义产品能够很好地运行在标准硬件如X86等设备上·。

`        `![../Desktop/屏幕快照%202022-06-17%20上午6.25.12.png](Aspose.Words.81ed48c3-d0c6-4cc7-adbc-ce89e7ef2af0.001.png)

MinIO的架构设计从一开始就是针对对性能要求很高的私有云标准，在实现对象存储所需要的全部功能的基础上追求极致的性能。MinIO具备易用性，高效性及高性能，能够以更简单的方式提供具有弹性伸缩能力的云原生对象存储服务。

MinIO在传统对象存储场景（如辅助存储，灾难恢复和归档）方面表现出色，同时在机器学习、大数据、私有云、混合云等方面的存储技术上也独树一帜，包括数据分析、高性能应用负载、原生云应用。


1. 架构设计

MinIOn为云原生架构设计，可以作为轻量级容器运行并由外部编排服务如Kubernetes管理。MinIO整个服务包约为不到100 MB的静态二进制文件，即使在很高负载下也可以高效利用CPU和内存资源并可以在共享硬件上共同托管大量租户。对应的架构图如下：

![../Desktop/屏幕快照%202022-06-17%20上午6.48.22.png](Aspose.Words.81ed48c3-d0c6-4cc7-adbc-ce89e7ef2af0.002.png)

MinIO可以在带有本地驱动器（JBOD / JBOF）的标准服务器上运行。集群为完全对称的体系架构，即所有服务器的功能均相同，没有名称节点或元数据服务器。

MinIO将数据和元数据作为对象一起写入从而无需使用元数据数据库。MinIO以内联，严格一致的操作执行所有功能包括擦除代码，位rotrot检查，加密等。

每个MinIO群集都是分布式MinIO服务器的集合，每个节点一个进程。 MinIO作为单个进程在用户空间中运行，并使用轻量级的协同例程来实现高并发。将驱动器分组到擦除集（默认情况下，每组16个驱动器），然后使用确定性哈希算法将对象放置在这些擦除集上。

MinIO专为大规模，多数据中心云存储服务而设计。每个租户都运行自己的MinIO群集，该群集与其他租户完全隔离，从而使他们能够保护他们免受升级，更新和安全事件的任何干扰。每个租户通过联合跨地理区域的集群来独立扩展。

1. 测试环境
   1. kubernetes集群

本次测试使用了三台虚拟机节点部署了kubernetes集群：1 master + 2 work nodes，kubelete版本为1.22.0

![../Desktop/屏幕快照%202022-07-05%20下午4.06.08.png](Aspose.Words.81ed48c3-d0c6-4cc7-adbc-ce89e7ef2af0.003.png)

1. hwameiStor本地存储

在kubernetes上部署了hwameiStor本地存储

![../Desktop/屏幕快照%202022-07-05%20下午4.08.19.png](Aspose.Words.81ed48c3-d0c6-4cc7-adbc-ce89e7ef2af0.004.png)

两台work节点各配置了五块磁盘sdb/sdc/sdd/sde/sdf用于hwameiStor本地磁盘管理。

![../Desktop/屏幕快照%202022-07-05%20下午4.10.54.png](Aspose.Words.81ed48c3-d0c6-4cc7-adbc-ce89e7ef2af0.005.png)	![../Desktop/屏幕快照%202022-07-05%20下午4.10.24.png](Aspose.Words.81ed48c3-d0c6-4cc7-adbc-ce89e7ef2af0.006.png)

local storage node状态				   ![../Desktop/屏幕快照%202022-07-11%20上午8.05.38.png](Aspose.Words.81ed48c3-d0c6-4cc7-adbc-ce89e7ef2af0.007.png)

创建了storagClass![../Desktop/屏幕快照%202022-07-05%20下午4.13.47.png](Aspose.Words.81ed48c3-d0c6-4cc7-adbc-ce89e7ef2af0.008.png)


1. 分布式多租户源码部署安装（minio operator）
   1. 部署minio operator

复制minio operator仓库到本地

git clone <https://github.com/minio/operator.git>

![../Desktop/屏幕快照%202022-07-11%20上午7.12.37.png](Aspose.Words.81ed48c3-d0c6-4cc7-adbc-ce89e7ef2af0.009.png)

![](Aspose.Words.81ed48c3-d0c6-4cc7-adbc-ce89e7ef2af0.010.png)

进入helm operator目录

/root/operator/helm/operator

![../Desktop/屏幕快照%202022-07-19%20上午8.27.07.png](Aspose.Words.81ed48c3-d0c6-4cc7-adbc-ce89e7ef2af0.011.png)



部署minio-operator实例

helm install minio-operator \

--namespace minio-operator \

--create-namespace \

--generate-name .

--set persistence.storageClass=local-storage-hdd-lvm .

检查minio-operator资源运行情况

![../Desktop/屏幕快照%202022-07-19%20上午8.29.44.png](Aspose.Words.81ed48c3-d0c6-4cc7-adbc-ce89e7ef2af0.012.png)

1. `	`创建租户

进入/root/operator/examples/kustomization/base目录

修改tenant.yaml如下

![../Desktop/屏幕快照%202022-07-20%20上午7.42.14.png](Aspose.Words.81ed48c3-d0c6-4cc7-adbc-ce89e7ef2af0.013.png)

进入/root/operator/helm/tenant/目录

修改values.yaml文件如下

![../Desktop/屏幕快照%202022-07-20%20上午7.48.34.png](Aspose.Words.81ed48c3-d0c6-4cc7-adbc-ce89e7ef2af0.014.png)

进入/root/operator/examples/kustomization/tenant-lite目录

修改kustomization.yaml文件如下：

![../Desktop/屏幕快照%202022-07-20%20上午7.52.31.png](Aspose.Words.81ed48c3-d0c6-4cc7-adbc-ce89e7ef2af0.015.png)

修改tenant.yaml文件如下：

![../Desktop/屏幕快照%202022-07-20%20上午7.52.47.png](Aspose.Words.81ed48c3-d0c6-4cc7-adbc-ce89e7ef2af0.016.png)

修改tenantNamePatch.yaml文件如下

![../Desktop/屏幕快照%202022-07-20%20上午7.53.00.png](Aspose.Words.81ed48c3-d0c6-4cc7-adbc-ce89e7ef2af0.017.png)

创建租户：

kubectl apply –f ./tenant.yaml 

检查租户资源状态：

![../Desktop/屏幕快照%202022-07-20%20上午10.11.24.png](Aspose.Words.81ed48c3-d0c6-4cc7-adbc-ce89e7ef2af0.018.png)


1. `	`HwameiStor本地卷配置

kubectl get statefulset.apps/minio-t1-pool-0 -nminio-tenant -oyaml

![](Aspose.Words.81ed48c3-d0c6-4cc7-adbc-ce89e7ef2af0.019.png)

kubectl get pvc –A

![../Desktop/屏幕快照%202022-07-20%20上午10.06.53.png](Aspose.Words.81ed48c3-d0c6-4cc7-adbc-ce89e7ef2af0.020.png)

kubectl get pvc export-minio6-0 -nminio-6 -oyaml

![../Desktop/屏幕快照%202022-07-11%20上午7.58.00.png](Aspose.Words.81ed48c3-d0c6-4cc7-adbc-ce89e7ef2af0.021.png)

kubectl get pv           

![../Desktop/屏幕快照%202022-07-20%20上午10.07.25.png](Aspose.Words.81ed48c3-d0c6-4cc7-adbc-ce89e7ef2af0.022.png)

kubectl get pvc data0-minio-t1-pool-0-0 -nminio-tenant -oyaml

![../Desktop/屏幕快照%202022-07-20%20上午10.18.49.png](Aspose.Words.81ed48c3-d0c6-4cc7-adbc-ce89e7ef2af0.023.png)

kubectl get lv

![../Desktop/屏幕快照%202022-07-20%20上午10.07.57.png](Aspose.Words.81ed48c3-d0c6-4cc7-adbc-ce89e7ef2af0.024.png)

kubect get lvr

![../Desktop/屏幕快照%202022-07-20%20上午10.08.14.png](Aspose.Words.81ed48c3-d0c6-4cc7-adbc-ce89e7ef2af0.025.png)

1. HwameiStor与MinIo测试验证
   1. `	`基本功能测试

登录minio console：10.6.163.52:30401/login

![../Desktop/屏幕快照%202022-07-11%20上午8.15.59.png](Aspose.Words.81ed48c3-d0c6-4cc7-adbc-ce89e7ef2af0.026.png)

用户名：XulV2mOJh9DKn3F95fdq

密码：PkXqxUJFZ8QYby2ZAVbeXQ9KW4ZAkdXH67HX2msx

![../Desktop/屏幕快照%202022-07-11%20上午8.17.57.png](Aspose.Words.81ed48c3-d0c6-4cc7-adbc-ce89e7ef2af0.027.png)

![../Desktop/屏幕快照%202022-07-11%20上午8.18.28.png](Aspose.Words.81ed48c3-d0c6-4cc7-adbc-ce89e7ef2af0.028.png)

创建bucket

![../Desktop/屏幕快照%202022-07-11%20上午9.46.02.png](Aspose.Words.81ed48c3-d0c6-4cc7-adbc-ce89e7ef2af0.029.png)

![../Desktop/屏幕快照%202022-07-11%20上午9.46.11.png](Aspose.Words.81ed48c3-d0c6-4cc7-adbc-ce89e7ef2af0.030.png)

创建path

![../Desktop/屏幕快照%202022-07-11%20上午9.46.32.png](Aspose.Words.81ed48c3-d0c6-4cc7-adbc-ce89e7ef2af0.012.png)

![../Desktop/屏幕快照%202022-07-11%20上午9.46.44.png](Aspose.Words.81ed48c3-d0c6-4cc7-adbc-ce89e7ef2af0.031.png)

上传文件

![../Desktop/屏幕快照%202022-07-11%20上午9.47.58.png](Aspose.Words.81ed48c3-d0c6-4cc7-adbc-ce89e7ef2af0.032.png)

`   `创建用户

`	`![../Desktop/屏幕快照%202022-07-11%20下午2.05.25.png](Aspose.Words.81ed48c3-d0c6-4cc7-adbc-ce89e7ef2af0.033.png)

`  `![../Desktop/屏幕快照%202022-07-11%20下午2.08.18.png](Aspose.Words.81ed48c3-d0c6-4cc7-adbc-ce89e7ef2af0.034.png)

![../Desktop/屏幕快照%202022-07-11%20下午2.08.27.png](Aspose.Words.81ed48c3-d0c6-4cc7-adbc-ce89e7ef2af0.035.png) 

![../Desktop/屏幕快照%202022-07-11%20下午2.08.38.png](Aspose.Words.81ed48c3-d0c6-4cc7-adbc-ce89e7ef2af0.036.png)

![../Desktop/屏幕快照%202022-07-13%20上午6.08.09.png](Aspose.Words.81ed48c3-d0c6-4cc7-adbc-ce89e7ef2af0.037.png)

用户policy

![../Desktop/屏幕快照%202022-07-13%20上午6.16.35.png](Aspose.Words.81ed48c3-d0c6-4cc7-adbc-ce89e7ef2af0.038.png)

![../Desktop/屏幕快照%202022-07-13%20上午6.16.53.png](Aspose.Words.81ed48c3-d0c6-4cc7-adbc-ce89e7ef2af0.039.png)





`	`删除bucket

`	`![../Desktop/屏幕快照%202022-07-13%20上午5.59.16.png](Aspose.Words.81ed48c3-d0c6-4cc7-adbc-ce89e7ef2af0.040.png)

![../Desktop/屏幕快照%202022-07-13%20上午5.59.36.png](Aspose.Words.81ed48c3-d0c6-4cc7-adbc-ce89e7ef2af0.041.png)

![../Desktop/屏幕快照%202022-07-13%20上午5.59.48.png](Aspose.Words.81ed48c3-d0c6-4cc7-adbc-ce89e7ef2af0.042.png)

![../Desktop/屏幕快照%202022-07-13%20上午6.00.36.png](Aspose.Words.81ed48c3-d0c6-4cc7-adbc-ce89e7ef2af0.043.png)

![../Desktop/屏幕快照%202022-07-13%20上午6.00.46.png](Aspose.Words.81ed48c3-d0c6-4cc7-adbc-ce89e7ef2af0.044.png)

![../Desktop/屏幕快照%202022-07-13%20上午6.01.18.png](Aspose.Words.81ed48c3-d0c6-4cc7-adbc-ce89e7ef2af0.045.png)




4. 结论

本次测试是在kubernetes 1.22平台上部署了minio分布式对象存储并对接hwameiStor本地存储。在此环境中完成了基本能力测试，系统安全测试及运维管理测试。

全部测试通过。
















