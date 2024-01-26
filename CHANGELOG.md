v0.14.1/ 2024-1-26
========================


## LVM volume management enhancements
- fixed a LVG issue: recreating a PVC can't find the correct LVG #1353(@sun7927 )
- skip volume when failurepolicy is ignore #1379(@SSmallMonster )
- skip volume when failurepolicy is ignore #1380(@SSmallMonster )
## Apiserver
- improve(apiserver):check localdisk exist before using #1339(@hikariwo )
- update snapshot client #1367(@SSmallMonster )
- Hwameictl snapshot #1331(@peng9808 )
- fix getnode api #1375(@peng9808 )
- fix ctl disk_list #1373(@peng9808 )
- fxi snapshot ctl bug #1333(@peng9808 )
- add cluster insatll ctl #1369(@peng9808 )
## Tests
- [unit-test]fix run failed volume_migrate_task test #1334 (@hikariwo )
- [unit-test]fix run failed volume_convert_task test #1336(@hikariwo )
## Documentation
- Added documentation for DRBDs adapted kernel version #1347(@FloatXD )
- Beautify document structure #1350(@FloatXD )
- Add carlory into MAINTAINERS.md #1355(@carlory )
- [docs] update readme #1365(@SSmallMonster )
- update support list #1378(@FloatXD )
- [docs] Update /install/prereq.md #1370(@windsonsea )
## Others
- specify the golang version in the builder #1357(@sun7927 )



v0.14.0/ 2023-12-26
========================

## LVM volume management enhancements
- fix(volume-group): always keep accessibility consistent with all volumes in group #1284(@SSmallMonster )
- fix(volume-group): fix node accessability incorrect caused by point reference #1291(@SSmallMonster )
- remove meaningless poolType using #1318(@buffalo1024 )
## Disk Management Enhancements
- change client type for supporting fake client to test #1292(@hikariwo )
## Volume Migration
- fix bug pending when pruning replica for localvolumemigrate #1290(@buffalo1024 )
## Volume Clone
- update volumeclone #1279(@SSmallMonster )
## Apiserver
- Update readme_zh.md #1265(@hikariwo )
- improve(apiserver):move ldHandler for improving cohesion #1268(@hikariwo )
- fix(apiserver): add exit system statement when indexer setup failed #1270(@hikariwo )
- fix(apiserver): getAvailableDiskCapacity found an empty capacity #1274(@hikariwo )
- fix(apiserver): fix StorageNodePoolDiskGet get localdisk from a wrong path #1275(@hikariwo )
- [unit-test] add more test for disk internal #1303(@hikariwo )
- improve(apiserver): improve pool-controller query efficiency #1305(@hikariwo )
- fix pool disk show bug #1308(@peng9808 )
- fix(csi-volume): fix volume leaks when user delete pvc but pv not created #1316(@SSmallMonster )
- improve(apiserver):check LocalDisks length when ListLocalDiskByNodeDevicePath returns #1321(@hikariwo )
- add VolumeSnapshotClass #1325(@peng9808 )
- Hwameictl snapshot #1331(@peng9808 )
- fxi snapshot ctl bug #1333(@peng9808 )
## Tests
- [unit-test] add updateLocalVolumeGroupAccessibility test #1281(@hikariwo )
- add ut for localdiskvolume handler #1282(@buffalo1024 )
- [unit-test]add smartctl command result example #1288(@hikariwo )
- update migrate e2e #1289(@FloatXD )
- change client type for supporting fake client to test #1292(@hikariwo )
- [unit-test] add more localdisknode test #1293(@hikariwo )
- [unit-test] add more test for localDiskVolume #1294 (@hikariwo )
- [unit-test] add disk interal test #1295(@hikariwo )
- [unit-test]add volume internal test #1296(@hikariwo )
- [unit-test] add localdiskclaim worker test #1297(@hikariwo )
- update migrate test #1299(@FloatXD )
- add more migrate test #1300(@FloatXD )
- [unit-test] add more test for disk internal #1303(@hikariwo )
- Use simplified sc for testing #1317(@FloatXD )
## Documentation
- [en] Add website preview instructions to README #1264(@windsonsea )
- Update readme_zh.md #1265(@windsonsea )
- update FAQ #1286(@FloatXD )
- Remove architecture page and improve some text #1302(@windsonsea )
- fix bullets consistency in faq page #1332(@windsonsea )


v0.13.1/ 2023-11-22
========================


## LVM volume management enhancements
- fix(snap-restore): filter replicaSnapRestoreName before commit tasks #1193(@SSmallMonster )
- Feat(volume-clone): Support VolumeClone #1194(@SSmallMonster )
- fix the localdisk.partitionInfo.path not display correctly #1199(@hikariwo )
- optimize datacopy job name generating #1196(@buffalo1024 )
- fix(LocalStorage): LocalVolumeConvert state transition error #1217(@SSmallMonster )
- fix bug missing pvc in cache map #1235(@buffalo1024 )
- fix(localStorage): imcomplete volumePath #1241(@hikariwo )
- fix(local-storage): potential data race in registry #1238(@hikariwo )
## Disk Management Enhancements
- add more fleid validation on struct Device on udev_test #1204(@hikariwo )
- fix: exit when indexer add failed #1211(@SSmallMonster )
- add more events for localdiskclaim #1249(@hikariwo )
- improve(diskclaim): only record disk claim events when disk is Available #1260(@SSmallMonster )
## Volume Migration
- fix(migrate): use storage node ip when migrate volume #1229(@SSmallMonster )
- fix(data-copy): prune replica after unpublish #1231(@SSmallMonster )
- fix(datacopy): overwrite node ip when create sync job #1232(@SSmallMonster )
- fix(dcp): only update source unpublished in src node #1236(@SSmallMonster )
- Migrate #1239(@peng9808 )
- fixed the migrate prune #1246(@sun7927 )
- add evict migrates into queue when evictor starts #1250(@sun7927 )
## Volume Snapshot
- feat(snapshot): delay volume deletion when snapshots found #1245(@SSmallMonster )
- make indexer spec.sourceVolume for LocalVolumeSnapshot #1247(@SSmallMonster )
- make indexer for snapshots #1248(@SSmallMonster )
## Scheduler
- feat(scheduler): filter node according to sourcevolume accessibility #1203(@SSmallMonster )
## Apiserver
- fix apiserver getnodedisk bug and add set-diskowner api #1188(@peng9808 )
- fix(apiserver): filter VolumeState when list replicas #1202(@hikariwo )
- add snapshot,expand api #1234(@peng9808 )
## Tests
- Disable k8s1.23.3 related tests #1190 (@FloatXD )
- Temporarily remove adaptation test for version 1.23 #1214(@FloatXD )
- e2e-test: add clone test #1215(@FloatXD )
- Upgrade the k8s version for adaptation test #1223(@FloatXD )
- add drbd parse event test example #1226(@hikariwo )
## Documentation
- docs: add user guide for volume clone #1205 (@SSmallMonster )
- add docs for pvc autoresizing #1206 (@buffalo1024 )
- update cli status as completed #1212(@SSmallMonster )
- fix snapshot doc #1233(@FloatXD )
- fix typos #1253(@yojay11717 )
- [docs] add fault management in roadmap #1257(@SSmallMonster )
- Clean up advanced_features #1179(@windsonsea )
- Update docs: pvc_autoresizing.md, volume_clone, and volume_provisioned_io #1208(@windsonsea )
- [i18n/cn] update the nav structure for docs #1256(@windsonsea )
- [en] Update nav structure (a big update) #1258(@windsonsea )

v0.13.0/ 2023-10-18
========================


## LVM snaprestore enhancement
- fix(snaprestore): set restore timeout 600s as default #1178 (@SSmallMonster )

## Apiserver
- fix nodeName filter error (@hikariwo )

## Adapt k8s v1.28
- Adapt k8s v1.28 #1154 (@buffalo1024 )

## Others
- Clean up advanced_features #1179 (@windsonsea )
- add resources value in values.extra.prod.yaml for new components #1180 (@buffalo1024 )
- set member is default log container #1183 (@zgfh )


v0.12.4/ 2023-10-11
========================


## LVM volume management enhancements
- added an option of juicesync to migrate data #1148 (@sun7927 )

## Disk Management Enhancements
- fix: correct safelyMount check #1165(@SSmallMonster )
- chore(disk-filter): add logger for debugging disk claim process #1162(@SSmallMonster )
- feat: parse disk type of NVMe #1160(@SSmallMonster )

## Scheduler
- fix getAssociatedVolumes double count error #1171(@SSmallMonster )
- fix(scheduler): schedule the published node when ha-volume is published #1167(@SSmallMonster )

## Tests
- fix e2e-test #1153 (@FloatXD )
- add 1.28 Adaptation-test #1164 (@FloatXD )
- fix ad-test #1166 (@FloatXD )
- add 1.24 ad-test #1172 (@FloatXD )

## Others
- Add hwameictl release ci #1146 (@Vacant2333）
- fixed the workflow for tools #1156 (@sun7927 )
- chroe: fix golangci-lint #1163(@SSmallMonster )
- chore(lint): add golangci-lint #1161(@SSmallMonster )


v0.12.3/ 2023-9-18
========================

## LVM volume management enhancements

- fix: allow lvm block volume to extend #1125 (@AmazingPangWei )
- stop mount process when duplicate device link found #1131(@SSmallMonster )
- feat: enable specify localDiskName for localdiskclaim #1089(@AmazingPangWei )
- show LD's owner directly #1121(@sun7927 )
- refact: remove CDevice struct #1139(@SSmallMonster )
- don't update disk type when ldm start up #1138 (@AmazingPangWei )

## Volume Snapshot and Restore

- rename SnapshotRecover to SnapshotRestore #1126 (@SSmallMonster )
- restore volume without judging snapshot exist #1141 (@SSmallMonster )
- scheduler: adapt volume create from snapshot #1134 (@SSmallMonster )

## PVC auto-resize

- modify pvc-autoresizer #1120 (@buffalo1024 )
- fix pvc autoresize failed #1142 (@buffalo1024 )
- fix nil pointer panic when pvc has no storageclass #1143 (@buffalo1024 )

## Tests

- add build local-disk-action-controller in e2e test #1123 (@FloatXD )
- add pvc autoresize test in e2e test #1127 (@FloatXD )
- add snapshot restore test in e2e test #1115 (@FloatXD )
- update recover to restore #1129 (@FloatXD )
- fix rollback test #1132 (@FloatXD )
- fix resizer test #1136 (@FloatXD )

## Documents

- update faqs.md #1078 (@windsonsea )
- Update documents related to crd and post_check #1119 (@FloatXD )
- Sync en faqs #1118 (@windsonsea )
- chore: remove link #1135 (@yyzxw )

## Others

- style: code style #1133 (@luckymrwang )

v0.12.2/ 2023-9-1
========================


## PVC auto-resize
- resizepolicy select pvc by label selector #1107 ( @buffalo1024 )

## Documents
- added the documents for audit and failover #1104 ( @sun7927 )



v0.12.1/ 2023-8-29
========================


## LVM volume management enhancements

- feat(local-storage): support ext{x} filesystem #1033 (@SSmallMonster )
- fix bug of displaying NVMe disk of LocalStoragePool node #1041 (@buffalo1024 )
- fix migrate operation missing update replica status #1047 (@Vacant2333 )
- The big enhancement of LocalVolumeGroup feature #1062 (@sun7927 )
- fix panic err when notfound #1093 (@SSmallMonster )

## Disk Management Enhancements

- add more attributes to localdisk #1032 (@AmazingPangWei )
- Update LocalDisk/LocalDiskVolume to identify Disk/DiskVolume #1039 (@SSmallMonster )
- fix(disk-claim): don't assign disks when owner in claim is empty #1053 (@SSmallMonster )
- add field to record device history info #1057 (@SSmallMonster )
- Use Serial/IDPATH to Identify Disk #1058 (@SSmallMonster )
- yum install xfsprogs when build local-disk-manager container image #1071 (@buffalo1024 )
- fix: remove stale localdisk during start #1088 (@AmazingPangWei )

## Volume Snapshot and Restore

- Develop snapshot #1090 @SSmallMonster
- [Feat][Snapshot] Restore Snapshot #981 @SSmallMonster


## PVC auto-resize

- add new component pvc-autoresizer #1016 @buffalo1024


## Volume QoS management

- Add cgroupsv2 support #1083 @carlory

## App Failover

- Failover #1008 @sun7927
- fixed failover assistant helm issues #1018 @sun7927
- fixed some deploy issues for failover assistant #1023 @sun7927
- fixed the Makefile issue for failover #1026 @sun7927
- deleted the CR which caused the installation failure #1060 @sun7927

## Audit

- Added a feature of audit for system resources, including cluster, storagenode, volume, disk #1056 @sun7927

## UI

- fix(ui yaml): ui run as nginx user #1081 @lsq645599166

## Tests

- update e2e test for throughput test #1009 @FloatXD
- update e2e test #1022 @FloatXD
- update AD test #1025 @FloatXD
- update AD test #1029 @FloatXD
- update AD test for failover-assistant #1030 @FloatXD
- update pr test paths #1034 @FloatXD
- Modify pr test coverage #1037 @FloatXD
- add lv node check in convert tests #1043 @FloatXD
- add convert tests in periodcheck #1046 @FloatXD
- update checkout job #1050 @FloatXD
- add unit tests #1010 @Vacant2333
- fix compile error on darwin #1101 @carlory
- fixed the unit tests bugs #1098 @sun7927

## Documents

- update operator.md #1011 @windsonsea
- update uninstall commands #1021 @windsonsea
- Add membership.md and members.yaml #1036 @windsonsea
- update output of kubectl get #1048 @windsonsea
- Fix the output of kubectl get in node expansion #1055 @windsonsea
- update faqs.md #1078@windsonsea
- update qos doc #1084 @carlory
- add doc for snapshot #1085 @FloatXD
- add doc for snapshot roollback #1094 @FloatXD
- update snapshot docs #1091 @SSmallMonster
- correct the information in readme #1100,#1099 @sun7927


v0.11.2/ 2023-7-25
========================

## LocalStorage
* refactor volume qos #966  (@carlory )

##  LocalDiskManager
* fix(disk-node): use /etc as device root path #994  (@SSmallMonster )
* update(ldm): mount /etc/hwameistor to container #999   (@SSmallMonster )

##  Apiserver
* Fix typo in docs #972  (@Vacant2333 )
* fix list local storage node #991  (@Vacant2333 )

## Other
* Add more tests  ( add qos e2e test #971 update e2e test #976 update e2e test #977 update qos test #978 update e2e test #998 @FloatXD  )
* Update Docs (Fix the Chinese documentation of Disk Expansion #964 Fix The Chinese documentation of Volume Provisioned IO has content du… #967 @FloatXD Fix doc issues #968 [Docs] Polish text in creating statefulset and uninstallation #980 update post_check.md #992 @windsonsea Fix doc issues #968 [Docs] Add disk owner description #969 @SSmallMonster Fix typo in docs #972 @calvin-puram add doc for reserving disk while iinstalling #990 @buffalo1024  ）

##  Admission,Scheduler,Evictor,Exporter
N/A


v0.11.1 / 2023-7-5
========================


## LocalStorage
* support auto-detect cgroup version #959 (@carlory )


## Other
* Add more tests ( update e2e test #941 @FloatXD )
* Update Docs (add volume_provisioned_io.md #932 @carlory docs(README): mark IO Throtting as Completed in Roadmap at v0.11.0 #937 docs(README): keep consitent with english #938 [Docs] update slack info #951 @SSmallMonster Modify the document information #939 @Seaiii docs: Improved command line format #936 @my-git9 update maintainers #943 change to cncf code of conduct #944 add cncf to readme #946 @windsonsea add cncf logo #945 add cncf logo and banners #949 @SAMZONG fix localdisk status docs #956 @wawa0210）

##  LocalDiskManager,Apiserver,Admission,Scheduler,Evictor,Exporter
N/A




v0.11.0 / 2023-6-25
========================

## LocalStorage
* LocalVolume implement IO Throttling or QoS #803 (@carlory )
* fix typo #898 (@panguicai008 )
* optimize(log): make funcs - candidate predicate log readable #913  (@SSmallMonster )

## LocalDiskManager
* Use */virtual/* detection for identifying virtual block devices #924 (@LucaDev )

## Apiserver
* list storagepool createtime field inconsistent #894  (@panguicai008 )
* fix ui migrate error #915  (@Vacant2333 )

## Other
* Add more tests   ( Hwameistor was installed by Operator ,some probelem happened when uninstalled  #887 update e2e for FlakeAttempts #905 update doc #907 update e2e #908 fix e2e test #911 fix pr e2e #927@FloatXD )
* Update Docs (Command format optimization #882 @my-git9 [docs] fix docs translations,add installation ways #883 @Vacant2333 add description of auto creating storageclass #895 @buffalo1024 Change some text in /operator.md #899 [docs] update section intro #916 add zh readme #922 update maintainer list #923 @windsonsea fix kubectl and storage typo #900 @panguicai008 ）
* fix apiserver deploy yaml #892 (@Vacant2333 )
* Fix scheduler config template, made hostpaths configurable in Helm Chart #925  (@LucaDev )

##  Admission,Scheduler,Evictor,Exporter
N/A



v0.10.3 / 2023-5-26
========================
## ApiServer
* update_apiserver_auth_helm_setting #878(@Vacant2333)

## Other
* update hwameistor-ui version #879(@Vacant2333 )

##  LocalStorage,LocalDiskManager,Admission,Scheduler,Evictor,Exporter
N/A


v0.10.2 / 2023-5-25
========================
## LocalDiskManager
* [Fix] Sanitize nodeName #875(@SSmallMonster )

## ApiServer
* Add auth to apiserver #864(@Vacant2333)

## Other
* fix get node name #872(@Vacant2333 )

##  LocalStorage,Admission,Scheduler,Evictor,Exporter
N/A


v0.10.1 / 2023-5-18
========================
## Other
* update helm for ui tag #869 (@FloatXD )

##  LocalStorage,LocalDiskManager,Admission,Apiserver,Scheduler,Evictor,Exporter
N/A


v0.10.0 / 2023-5-17
========================


## LocalDiskManager
* recognize and setup disks managed by local-disk-manager owner #840(@SSmallMonster )
* [Feat] mark disk state to inactive when receive remove events #841(@SSmallMonster )
* update disk smart pannels #856(@SSmallMonster )
* [Feat] Implement LocalDisk_Pool{HDD,SSD,NVMe} #701 (@SSmallMonster )

## Apiserver
* regard lv in use only when publishedNode is the same as srcNode in mi… #829 (@buffalo1024 )
* update update_non-standard_code #851 (@Vacant2333 )
* fix bug of apiserver showing components status #860 (@buffalo1024 )



## Other
* Add more tests ( update e2e #842 update e2e #843 update Period check go version #845 update ad e2e #847 update ad e2e #852 @FloatXD )
* Update Docs (remove step which apply cluster cr in operator.md doc #826 @buffalo1024 [documents] update quickstart with operator #832 revert doc of installing by operator #835 @SAMZONG update /create_stateful/basic/local.md #837 update a command in post_check.md #863 @windsonsea [Docs] sync ldm module description #858 [Docs] Add Volumes and Nodes #861 fix typo #865@SSmallMonster ）
* updated the roadmap with the new features #827 (@sun7927 )
* add namespace env #862 (@Vacant2333 )

## LocalStorage,Admission,Apiserver,Scheduler,Evictor,Exporter
N/A

v0.9.3 / 2023-4-24
========================

## LocalStorage
* fix: ignore state when list bounded disks #813(@SSmallMonster )

## LocalDiskManager
* populate disk owner when empty #762 (@SSmallMonster )
* remove hostNetwork #776 (@SSmallMonster )
* disable metrics serving #779 (@SSmallMonster )
* dismiss not found error #782(@SSmallMonster )
* [Fix] merge disk self attrs when triggerd by udev events #788 (@SSmallMonster )
* add labels on service #812 (@SSmallMonster )
* [Enhance] separate disk assign and disk status update process #815 (@SSmallMonster )

## Exporter
* rename exporter port to http-metrics #798 (@SSmallMonster )
* add port name on exporter service #821 (@SSmallMonster )

## UI
* fix ui template error #756 (@SSmallMonster )

## Other
* Add more tests ( update relok8s #752 update e2e #758 update e2e #764 update e2e #765 @FloatXD )
* Update Docs ([zh-docs] sync /quick_start/install/operator.md #751 @windsonsea update doc #784 update doc #800 @FloatXD updated the documents for the latest features #787 @SSmallMonster updated the documents #789 updated the document by removing scheduler configuration #790 removed the document for upgrade section #791 @sun7927 ）
* mark roadmap for observability and operator as completed in v0.9.x #761(@SSmallMonster )
* ignore hwameistor/Chart.yaml when trriger relok8s check action #824 (@SSmallMonster )

## Admission,Apiserver,Scheduler,Evictor
N/A

v0.9.2 / 2023-3-28
========================

## Other
* Add ui relok8s  (@FloatXD)


## LocalDiskManager,Scheduler,Apiserver,Evictor,Exporter,Admission，LocalStorage
N/A


v0.9.1 / 2023-3-28
========================

## LocalStorage
* enabled the volume stats capability #741 (@sun7927 )
* corrected the local-storage deploy #742 (@sun7927 )

## LocalDiskManager,Scheduler,Apiserver,Evictor,Exporter,Admission
N/A


v0.9.0 / 2023-3-28
========================

## LocalStorage
* track the volume's used capacity #667 (@sun7927 )
* optimize log level args(default: 4 - info) #691 (@SSmallMonster )
* remove debug call #692 (@SSmallMonster )

## LocalDiskManager
* Feat: Support specify disk owner #681 (@SSmallMonster )

## Scheduler
* ignore NotFound error according to FailurePolicy #671 (@SSmallMonster )
* [Scheduler] skip score if no new volumes found #724 (@SSmallMonster )

## Apiserver
* to merge new codes for apiserver #694 (@SSmallMonster )

## Admission
* fix start error caused by args parse #698 (@SSmallMonster )
* update apiserver #732 (@buffalo1024 )

## Other
* Add grafana dashboard (#733 #735 #736 @sun7927 )
* Add more tests  ( #673 #674 #680 #683 #695 #704 #705 #714 @FloatXD )
* Update Docs (#675 #676 #677 #693 #697 @windsonsea )
* delete unused file #665 (@SSmallMonster )
* bump up the hwameistor-operator version #666 (@sun7927 )
* add module.go to support import crds directory #668 (@SSmallMonster )
* fix wrong describe for update pvc. #670 (@yanggangtony )
* add ui deployment #679 (@SSmallMonster )
* Improve docs for kubectl command #706 (@mengjiao-liu )
* ui: add app label to ui service #710 (@SSmallMonster )
* [Docs] Update migrate.md #711 (@nameYULI )
* update helm icon #712 (@FloatXD )
* Update release status in README #715 (@Zhuzhenghao )
* use NewClientBuilder instead of the deprecated NewFakeClientWithScheme #716 (@Fish-pro )
* clean up duplicate package imports #717 (@Fish-pro )
* fix fatal misspellings #718 (@Fish-pro )
* set drbdStartPort 43001 #723 (@SSmallMonster )
* Update volume_eviction.md #729 (@yanzhifa )
* sync owner in charts and docs #731 (@SSmallMonster )


## Evictor,Exporter
N/A


v0.8.0 / 2023-2-23
========================

## LocalDiskManager

* Added metrics for volume operations #639 (@sun7927 )

* [Feat] Delete the Claim after it has been consumed #641 (@SSmallMonster )

## Other

* Add more api tests ( fix test #636 fix e2e #649 fix e2e #655 @FloatXD )
* update registry #638 (@FloatXD )
* fix chinese doc #645 (@FloatXD )
* added maintainers #650 (@sun7927 )
* [Chart] Render chart values.yaml  #651 (@SSmallMonster )
* [relok8s] Add relok8s hint config #652 (@SSmallMonster )
* update latest version and roadmap #653 (@SSmallMonster )
* update hwameistor image registry to ghcr.io #654 (@SSmallMonster )
* set default failurePolicy:Ignore in admission config #657 (@SSmallMonster )

## LocalStorage,Scheduler,Evictor,Admission,Exporter
N/A

v0.7.2 / 2023-2-8
========================

##LocalDiskManager
* rename MoveMountPoint into RemoveMountPoint and simplify range expression in localdiskvolume #610 (@carlory )
* fix potenitial panic in raid #609 (@carlory )
* improve resultCodeIsOk comment #608 (@carlory )
* added some fields in localvolume status to record storage usage #600 (@sun7927 )

##LocalStorage
* make localregistry rebuilding process correct #619 (@SSmallMonster )
* Resize StoragePool Capacity when disk capacity changed #618 (@SSmallMonster )

##Scheduler
* update scheduler config #602 (@SSmallMonster )
* Extend Scheduler Score Plugin #601 (@SSmallMonster )

##Exporter
* renamed metrics collector to exporter and refined the collector manager #621 (@sun7927 )

##Evictor,Admission
N/A

##Other
* Add more api tests ( [test]update api test #625 [test]update api test #623 [test]update api test #622 [test]update api test #616 [test]add api test #615 [test]update api test #616 @FloatXD )
* make for link helm crds #631 (@SSmallMonster )
* Update fs-resize tools in pre-requirements docs #628 (@SSmallMonster )
* fix go-import-lint #605 (@carlory )
* remove reduant gitkeep #606 (@carlory )
* add print headers for the localdisk resource #607 (@carlory )

v0.7.1 / 2023-01-06
========================
* Check for hwameistor GC JOB before processing (#591 #593 @sun7927 @SSmallMonster)

## Evictor,Admission,LocalStorage,Scheduler,Metrics
N/A

v0.7.0 / 2022-12-27
========================

## LocalDiskManager
* Feat: Collect S.M.A.R.T metrics #545 (@SSmallMonster )
* fix disk status bug in diskvolume mode #552 (@SSmallMonster )
* Feat: Expose S.M.A.R.T metrics #554 (@SSmallMonster )
* [S.M.A.R.T] Save SMART result to configmap #563 (@SSmallMonster )

## Apiserver
* added apiserver module #556 ( @sun7927 )
* add[apiserver]: add apiserver interface refracture #561 ( @angel0507 )
* add[apiserver]: add apiserver interface param update #562 ( @angel0507 )

## Metrics
* add metrics feature #546 ( @sun7927 )

## Other
* Add more e2e tests ( [test]add reliability test #535 [test]Add comprehensive test #537 [test]fix test #541 [test]add auto ad test #549 @FloatXD )
* [docs] update folders of terms and quickstart #538 ( @windsonsea ）
* correct testcases that wrong in logic #540 ( @buffalo1024 )
* removed servicemonitor from Helm #553 ( @sun7927 )
* added swag make command in Makefile #557 ( @sun7927 )
* removed apiserver from unit test #558 ( @sun7927 )
* added apiserver swag and run in Makefile #559 ( @sun7927 )
* optimized the Makefile for swag init #569 ( @sun7927 )
* generated metadata field of the swagger json by swag v1.8.9 #570 ( @sun7927 )
* generate swagger json in builder Docker image #571 ( @sun7927 )


## Evictor,Admission,LocalStorage,Scheduler
N/A

v0.6.1 / 2022-12-08
========================

## LocalDiskManager
* Reconcile localdiskclaim when no disks found in Bound status  #530 ( @SSmallMonster )
* Query LocalDiskClaim Directly After Updating LocalDisk  #529 ( @SSmallMonster )
* Setup attachnode in localdisk without dot  #525 ( @SSmallMonster )
* Optimize logic when localdiskclaim Bound  #519 ( @SSmallMonster )
* Check localdisk twice when Bound already  #518 ( @SSmallMonster )

## Scheduler
* Remove kubeconfig in scheduler config  #520 ( @SSmallMonster )

## Other
* Add more e2e tests  ( #516 #515 #514 @FloatXD )
* Add more unit tests (#510 #509 #512 #513 @buffalo1024 )
* fix Chinese document  #511 ( @FloatXD )
* update readme  #508 ( @FloatXD )


v0.6.0 / 2022-11-28
========================
## LocalDiskManager
* Support Health Check By S.M.A.R.T (#496 @SSmallMonster )
* Optimize the DISK State and reconstruct the state flow process (#464 @SSmallMonster )

## Evictor
* added an option to disable storage node volume eviction (#493 @sun7927 )

## Scheduler,Admission,LocalStorage
N/A

## Other
* [Docs] Add step about changelog (#500 @SSmallMonster )
* updated CODE_OF_CONDUCT.md @#466 @windsonsea )
* updated the document for eviction (#494 @sun7927 )

v0.5.0 / 2022-11-24
========================
## Other
* Upgrade version to kubev1.21 to support kube1.25 (#427 @SSmallMonster )
* Add more e2e tests(#489, #487, #487 @FloatXD )

v0.4.3 / 2022-11-14
========================
## Admission
* Fix panic error when label is empty on Namespace (#447 @SSmallMonster )

## Scheduler,LocalStorage,LocalDiskManager,Evictor
N/A

## Other
* Improve the docs style of README(#446 @windsonsea )
* Update e2e test (#459 @FloatXD )
* Add CHANGELOG.md & release/v0.4.{0,1,2}/changelog (#461 @SSmallMonster )

v0.4.2 / 2022-11-02
========================
## LocalStorage
Fix allocate config error (#439 @sun7927 )

v0.4.1 / 2022-11-02
========================
# Feature / Major changes
## LocalStorage
* Reduce conditions in LocalStorageNode (#435 @SSmallMonster)
* Fix error when VolumeExpand (#433 @sun7927 )

## Evictor
* Refactor the configuration for VolumeMigrate (#429 @sun7927 )

# Others
* Update Roadmap (#414, #416 @sun7927 )

v0.4.0 / 2022-10-28
========================
# Feature / Major changes
## Evictor
> We now support automatically migrate HwameiStor volumes.

* Support eviction for node and pod(#386, #389, #328 @sun7927 )
* Add docs about eviction(#398 @sun7927 )

## LocalStorage
* Optime volume migrate(#380, #361 @sun7927 )
* Fix capacity leakage(@sun7927)
* Add events and message about LSN(#327 @SSmallMonster )

## LocalDiskManager
* Optimize disk claim logic(#318, #317 @SSmallMonster )

## AdmissionController
* Optimize namespace selector label and ignore kube-system by default(#400 @SSmallMonster )
* Panic if create admission config fail(#355 @SSmallMonster )

# Other changes
* Add release process docs(#399, #409 @SSmallMonster )
* Add unit test for localdiskclaim(#383, #379, #378 @SSmallMonster )
* Support image scan with trivy (#362 @SSmallMonster )
* Fix CVE-2021-43527, CVE-2022-1996(#373, #372 @SSmallMonster )
* Correct annations in storageclass when helm install(#309 @SSmallMonster )
* Add more e2e test(#314 @FloatXD )
* Add e2e test for drbd installer (#297 @FloatXD )
* Add e2e test for performance (#374 @FloatXD )
* Improve docs (#408 @Michelle951 )
