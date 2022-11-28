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
