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
