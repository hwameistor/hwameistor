---
sidebar_position: 12
sidebar_label: "Metrics"
---

# Metrics

HwameiStor exposes a variety of metrics through hwameistor-exporter, which can be obtained in the following ways:

```bash
curl hwameistor-exporter:80/metrics
```

The following are some examples of the metrics obtained:

## Localdisk Metrics

```bash
# HELP hwameistor_localdisk_capacity The capacity of the localdisk.
# TYPE hwameistor_localdisk_capacity gauge
hwameistor_localdisk_capacity{devPath="/dev/sda",nodeName="k8s-master",owner="system",reserved="true",status="Bound",type="HDD"} 1.37438953472e+11
```

| Name | Description | Labels (Values) | Metric type |
|------|-------------|-----------------|-------------|
| hwameistor_localdisk_capacity | Localdisk capacity metric | devPath,nodeName,owner(system,local-disk-manager,local-storage,Unknown),reserved(true,false),status(Available,Bound),type(HDD,SSD,NVMe) | gauge |

## Localstoragenode Metrics

```bash
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

| Name | Description | Labels (Values) | Metric type |
|------|-------------|-----------------|-------------|
| hwameistor_localstoragenode_capacity | Localstoragenode capacity metric | kind(Free,Used),nodeName,poolName(HDD,SSD,NVMe) | gauge |
| hwameistor_localstoragenode_status | Localstoragenode status metric | nodeName,status(Ready,NotReady) | gauge |
| hwameistor_localstoragenode_volumecount | Localstoragenode volume metric | kind(Free,Used),nodeName,poolName(HDD,SSD,NVMe) | gauge |

## Localvolume Metrics

```bash
# HELP hwameistor_localvolume_capacity The capacity of the localvolume.
# TYPE hwameistor_localvolume_capacity gauge
hwameistor_localvolume_capacity{kind="Allocated",mountedOn="k8s-node2",poolName="HDD",type="Convertible",volumeName="pvc-d1964bc9-9b0b-456d-be1a-6d0de9b47589"} 1.073741824e+09

# HELP hwameistor_localvolume_status The status summary of the localvolume.
# TYPE hwameistor_localvolume_status gauge
hwameistor_localvolume_status{mountedOn="k8s-node2",poolName="HDD",status="Ready",type="Convertible",volumeName="pvc-d1964bc9-9b0b-456d-be1a-6d0de9b47589"} 1
```

| Name | Description | Labels (Values) | Metric type |
|------|-------------|-----------------|-------------|
| hwameistor_localvolume_capacity | Localvolume capacity metric | kind(Allocated,Used),mountedOn,poolName(HDD,SSD,NVMe),type(Convertible,NonHA),volumeName | gauge |
| hwameistor_localvolume_status | Localvolume status metric | mountedOn,poolName(HDD,SSD,NVMe),status(Ready,NotReady),type(Convertible,NonHA),volumeName | gauge |

## Localvolumereplica Metrics

```bash
# HELP hwameistor_localvolumereplica_capacity The capacity of the localvolumereplica.
# TYPE hwameistor_localvolumereplica_capacity gauge
hwameistor_localvolumereplica_capacity{nodeName="k8s-node2",poolName="HDD",volumeName="pvc-d1964bc9-9b0b-456d-be1a-6d0de9b47589"} 1.073741824e+09

# HELP hwameistor_localvolumereplica_status The status of the localvolumereplica.
# TYPE hwameistor_localvolumereplica_status gauge
hwameistor_localvolumereplica_status{nodeName="k8s-node2",poolName="HDD",status="Ready",volumeName="pvc-d1964bc9-9b0b-456d-be1a-6d0de9b47589"} 1
```

| Name | Description | Labels (Values) | Metric type |
|------|-------------|-----------------|-------------|
| hwameistor_localvolumereplica_capacity | Localvolumereplica capacity metric | nodeName,poolName(HDD,SSD,NVMe),volumeName | gauge |
| hwameistor_localvolumereplica_status | Localvolumereplica status metric | poolName(HDD,SSD,NVMe),status(Ready,NotReady),volumeName | gauge |
