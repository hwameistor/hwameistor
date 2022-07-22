# HwameiStor

Hwameistor is a cloud-native storage system. It creates a local storage resource pool for centrally managing all disks such as HDD, SSD, and NVMe. It uses the CSI architecture to provide distributed services with local volumes and provides data persistence capabilities for stateful cloud-native workloads or components.

![System architecture](docs/docs/img/architecture.png)

## Current Status

## Build Status

## Release Status

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

## Community
### Blog

Please follow our weekly blogs [here](https://hwameistor.io/blog).
### Slack
### WeChat
### Meetup

## Discussion

## Requests and Issues

## License

Copyright (c) 2014-2021 The HwameiStor Authors

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

