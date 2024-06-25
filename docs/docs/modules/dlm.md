---
sidebar_position: 2
sidebar_label: "DataLoad Manager"
---

# DataLoad Manager

DataLoad Manager is a module of DataStor, which is a cloud-native local storage system acceleration solution in AI scenarios. It combines p2p technology to provide the ability to quickly load remote data.

## Applicable scenarios

DataloadManager supports multiple data loading protocols: s3, nfs, ftp, http, ssh

In AI data training scenarios, data can be loaded into local cache volumes faster.
Especially when the data set supports s3 protocol pull, p2p technology can be combined to significantly improve data loading.
## Usage with DataLoad Manager

DataLoad Manager is a component of HwameiStor and must work with the [DataLoad Manager](./dsm.md) module.

