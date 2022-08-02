# HwameiStor

Hwameistor is a cloud-native storage system. It creates a local storage resource pool for centrally managing all disks such as HDD, SSD, and NVMe. It uses the CSI architecture to provide distributed services with local volumes and provides data persistence capabilities for stateful cloud-native workloads or components.

![System architecture](docs/docs/img/architecture.png)

## Current Status
>At present, HwameiStor is still in the alpha stage.

The latest release of HwameiStor is [![hwameistor-releases](https://img.shields.io/github/v/release/hwameistor/hwameistor.svg?include_prereleases)](https://github.com/hwameistor/hwameistor/releases)
## Build Status
![period-check](https://github.com/hwameistor/hwameistor/actions/workflows/period-check.yml/badge.svg) [![codecov](https://codecov.io/gh/hwameistor/hwameistor/branch/main/graph/badge.svg?token=AWRUI46FEX)](https://codecov.io/gh/hwameistor/hwameistor) [![OpenSSF Best Practices](https://bestpractices.coreinfrastructure.org/projects/5685/badge)](https://bestpractices.coreinfrastructure.org/projects/5685)
## Release Status
| Release  | Version | Type   |    
|----------|---------|--------|
| v0.3     | v0.3.2  | latest |

## Modules and Code

HwameiStor contains modules local-disk-manager, local-storage, scheduler.

### Local-Disk-Manager
Local-Disk-Manager(LDM) is designed to hold the management of disks on nodes.Other modules such as Local-Storage can take advantage of the management of disks by LDM.See more at LDM

### Local-Storage
Local-Storage provides a cloud native local storage system.It aims to provision high performance persistent LVM volume with local access to applicatios.See more at LS

### Scheduler
The Scheduler is to automatically schedule the Pod to the correct node which has the associated HwameiStor volume

## Documentation

A full documentation is hosted at our project website [hwameistor.io](https://hwameistor.io/docs/intro).

## Roadmap
| Feature                                  	| Status    	| Release 	|  Description                                     	|
|------------------------------------------	|-----------	|---------	|--------------------------------------------------	|
| CSI for LVM volume                       	| Completed 	| v0.3.2  	| Provision volume with lvm                        	|
| CSI for disk volume                      	| Completed 	| v0.3.2  	| Provision volume with disk                       	|
| HA LVM Volume                            	| Completed 	| v0.3.2  	| Volume with HA                                   	|
| HA LVM Volume expansion                  	| Completed 	| v0.3.2  	| Expand LVM volume capacity online                	|
| LVM Volume migration                     	| Completed 	| v0.3.2  	| Migrate a LVM volume replica to a different node 	|
| LVM Volume conversion                    	| Completed 	| v0.3.2  	| Convert a non-HA LVM volume to the HA            	|
| Volume Group                             	| Completed 	| v0.3.2  	| Support volume group allocation                  	|
| Observability                            	| Planed    	|         	| Observability                                    	|
| non-HA LVM volume mirror                 	| Planed    	|         	| Mirror LVM volume                                	|
| non-HA LVM volume clone                  	| Planed    	|         	| Clone LVM volume                                 	|
| non-HA LVM volume snapshot               	| Planed    	|         	| Snapshot of LVM volume                           	|
| non-HA LVM volume thin provision support 	| Planed    	|         	| non-HA LVM volume thin provision support         	|
| non-HA LVM volume stripe writing support 	| Planed    	|         	| non-HA LVM volume stripe writing support         	|
| data encryption                          	| Planed    	|         	| Data encryption                                  	|
| Disk health check                        	| Planed    	|         	| Fault prediction, status information reporting   	|


## Community
### Blog

Please follow our weekly blogs [here](https://hwameistor.io/blog).
### Slack
### WeChat
### Meetup

## Discussion

## Requests and Issues
Please submit any feedback and issue at: [Issues](https://github.com/hwameistor/hwameistor/issues)
## License

Copyright (c) 2014-2021 The HwameiStor Authors

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

