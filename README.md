# HwameiStor

Hwameistor is a cloud-native storage system. It creates a local storage resource pool for centrally managing all disks such as HDD, SSD, and NVMe. It uses the CSI architecture to provide distributed services with local volumes and provides data persistence capabilities for stateful cloud-native workloads or components.

![System architecture](docs/docs/img/architecture.png)

## Current Status

## Build Status

## Release Status

## Modules and Code

HwameiStor contains 3 modules:
* local-disk-manager
* local-storage
* scheduler

### Local-Disk-Manager
Local-Disk-Manager (LDM) is designed to hold the management of disks on nodes. Other modules such as Local-Storage can take advantage of the management of disks by LDM.

### Local-Storage
Local-Storage (LS) provides a cloud-native local storage system. It aims to provision high-performance persistent LVM volume with local access to applicatios.

### Scheduler
The Scheduler is to automatically schedule the Pod to the correct node which has the associated HwameiStor volume.

## Documentation

Full documentation is hosted at our project website [hwameistor.io](https://hwameistor.io/docs/intro).

## Roadmap

## Community

### Blog

Please follow our weekly blogs [here](https://hwameistor.io/blog).

### Slack

Our slack channel is [here](https://join.slack.com/t/hwameistor/shared_invite/zt-1dkabcq2c-KIRBJDBc_GgZZfeLrooK6g)

### WeChat
HwameiStor tech-talk group:

![QR code for Wechat](./docs/docs/img/wechat.png)

## Discussion

Welcome to follow our roadmap discussions [here](https://github.com/hwameistor/hwameistor/discussions)

## Requests and Issues

Please feel free to raise requests on chats or by a PR.  

We will try our best to respond to every issue reported on community channels, but the issues reported [here](https://github.com/hwameistor/hwameistor/discussions) on this repo will be addressed first.

## License

Copyright (c) 2014-2021 The HwameiStor Authors

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0