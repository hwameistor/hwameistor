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
