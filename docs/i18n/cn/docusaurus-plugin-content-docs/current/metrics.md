---
sidebar_position: 12
sidebar_label: "指标"
---

# 指标

HwameiStor 通过 hwameistor-exporter 暴露出了丰富的指标，可以通过下列方式获取指标：

  ```bash
  curl hwameistor-exporter:80/metrics
  ```

获取的指标样例如下：


### 本地磁盘指标
```
# HELP hwameistor_localdisk_capacity The capacity of the localdisk.
# TYPE hwameistor_localdisk_capacity gauge
hwameistor_localdisk_capacity{devPath="/dev/sda",nodeName="k8s-master",owner="system",reserved="true",status="Bound",type="HDD"} 1.37438953472e+11
```

| 名称                            | 描述       | 标签(值)                                                                                                                                   | 指标种类  |
|-------------------------------|----------|-----------------------------------------------------------------------------------------------------------------------------------------|-------|
| hwameistor_localdisk_capacity | 本地磁盘容量指标 | devPath,nodeName,owner(system,local-disk-manager,local-storage,Unknown),reserved(true,false),status(Available,Bound),type(HDD,SSD,NVMe) | gauge |


### 本地存储节点指标
```
# HELP hwameistor_localstoragenode_capacity The storage capacity of the localstoragenode.
# TYPE hwameistor_localstoragenode_capacity gauge
hwameistor_localstoragenode_capacity{kind="Free",nodeName="k8s-master",poolName="HDD"} 1.07369988096e+11

# HELP hwameistor_localstoragenode_status The status of the localstoragenode.
# TYPE hwameistor_localstoragenode_status gauge
hwameistor_localstoragenode_status{nodeName="k8s-master",status="Ready"} 1

# HELP hwameistor_localstoragenode_volumecount The volume count of the localstoragenode.
# TYPE hwameistor_localstoragenode_volumecount gauge
hwameistor_localstoragenode_volumecount{kind="Free",nodeName="k8s-master",poolName="HDD"} 1000
```



| 名称                                      | 描述          | 标签(值)                                           | 指标种类  |
|-----------------------------------------|-------------|-------------------------------------------------|-------|
| hwameistor_localstoragenode_capacity    | 本地存储节点容量指标  | kind(Free,Used),nodeName,poolName(HDD,SSD,NVMe) | gauge |
| hwameistor_localstoragenode_status      | 本地存储节点状态指标  | nodeName,status(Ready,NotReady)                 | gauge |
| hwameistor_localstoragenode_volumecount | 本地存储节点数据卷指标 | kind(Free,Used),nodeName,poolName(HDD,SSD,NVMe) | gauge |






### 本地存储数据卷指标
```
# HELP hwameistor_localvolume_capacity The capacity of the localvolume.
# TYPE hwameistor_localvolume_capacity gauge
hwameistor_localvolume_capacity{kind="Allocated",mountedOn="k8s-node2",poolName="HDD",type="Convertible",volumeName="pvc-d1964bc9-9b0b-456d-be1a-6d0de9b47589"} 1.073741824e+09

# HELP hwameistor_localvolume_status The status summary of the localvolume.
# TYPE hwameistor_localvolume_status gauge
hwameistor_localvolume_status{mountedOn="k8s-node2",poolName="HDD",status="Ready",type="Convertible",volumeName="pvc-d1964bc9-9b0b-456d-be1a-6d0de9b47589"} 1
```



| 名称                              | 描述          | 标签(值)                                                                                      | 指标种类  |
|---------------------------------|-------------|--------------------------------------------------------------------------------------------|-------|
| hwameistor_localvolume_capacity | 本地存储数据卷容量指标 | kind(Allocated,Used),mountedOn,poolName(HDD,SSD,NVMe),type(Convertible,NonHA),volumeName   | gauge |
| hwameistor_localvolume_status   | 本地存储数据卷状态指标 | mountedOn,poolName(HDD,SSD,NVMe),status(Ready,NotReady),type(Convertible,NonHA),volumeName | gauge |                                                                             |       |





### 本地存储数据卷副本指标
```
# HELP hwameistor_localvolumereplica_capacity The capacity of the localvolumereplica.
# TYPE hwameistor_localvolumereplica_capacity gauge
hwameistor_localvolumereplica_capacity{nodeName="k8s-node2",poolName="HDD",volumeName="pvc-d1964bc9-9b0b-456d-be1a-6d0de9b47589"} 1.073741824e+09

# HELP hwameistor_localvolumereplica_status The status of the localvolumereplica.
# TYPE hwameistor_localvolumereplica_status gauge
hwameistor_localvolumereplica_status{nodeName="k8s-node2",poolName="HDD",status="Ready",volumeName="pvc-d1964bc9-9b0b-456d-be1a-6d0de9b47589"} 1
```



| 名称                                     | 描述            | 标签(值)                                                    | 指标种类  |
|----------------------------------------|---------------|----------------------------------------------------------|-------|
| hwameistor_localvolumereplica_capacity | 本地存储数据卷副本容量指标 | nodeName,poolName(HDD,SSD,NVMe),volumeName               | gauge |
| hwameistor_localvolumereplica_status   | 本地存储数据卷副本状态指标 | poolName(HDD,SSD,NVMe),status(Ready,NotReady),volumeName | gauge |




