package storage

//const (
//	// HDD: rotational disk, SSD: non-rotational disk
//	diskRotational    = "1"
//	diskNonRotational = "0"
//)

//type localDiskManager struct {
//	cmdExec exechelper.Executor
//	lm      *LocalManager
//	logger  *log.Entry
//}

//func newLocalDiskManager(lm *LocalManager) LocalDiskManager {
//	return &localDiskManager{
//		cmdExec: nsexecutor.New(),
//		logger:  log.WithField("Module", "NodeManager/LocalDiskManager"),
//		lm:      lm,
//	}
//}

//func (mgr *localDiskManager) GetLocalDisks() (map[string]*localstoragev1alpha1.LocalDisk, error) {
//	disks := make(map[string]*localstoragev1alpha1.LocalDisk)
//
//	disksInUse, err := mgr.discoverDisksInUse()
//	if err != nil {
//		return disks, err
//	}
//
//	diskCapacities, err := mgr.discoverDiskCapacities()
//	if err != nil {
//		return disks, err
//	}
//
//	params := exechelper.ExecParams{
//		CmdName: "lsblk",
//		CmdArgs: []string{"-d", "-n", "-o", "name,type,rota"},
//	}
//	res := mgr.cmdExec.RunCommand(params)
//	if res.ExitCode != 0 {
//		return disks, res.Error
//	}
//	pciDisks, err := utils.GetPCIDisks(mgr.cmdExec)
//	if err != nil {
//		return disks, err
//	}
//	for _, line := range strings.Split(res.OutBuf.String(), "\n") {
//		items := regexp.MustCompile(" +").Split(strings.TrimPrefix(line, " "), -1)
//		if len(items) >= 3 && items[1] == "disk" {
//			if _, ok := pciDisks[items[0]]; !ok {
//				mgr.logger.Debugf("Disk %s is not SCSI device.", items[0])
//				continue
//			}
//			devpath := fmt.Sprintf("/dev/%s", items[0])
//
//			var diskClass string
//			if pciDisks[items[0]].IsNVMe() {
//				diskClass = localstoragev1alpha1.DiskClassNameNVMe
//			} else {
//				switch items[2] {
//				case diskRotational:
//					diskClass = localstoragev1alpha1.DiskClassNameHDD
//				case diskNonRotational:
//					diskClass = localstoragev1alpha1.DiskClassNameSSD
//				}
//			}
//			disk := &localstoragev1alpha1.LocalDisk{
//				DevPath: devpath,
//				Class:   diskClass,
//				State:   localstoragev1alpha1.DiskStateAvailable,
//			}
//			if capacity, ok := diskCapacities[devpath]; ok {
//				disk.CapacityBytes = capacity
//			} else {
//				mgr.logger.WithFields(log.Fields{"disk": devpath}).Error("Disk capacity not found.")
//			}
//			if _, ok := disksInUse[devpath]; !ok {
//				disk.State = localstoragev1alpha1.DiskStateAvailable
//				disks[devpath] = disk
//			} else {
//				disk.State = localstoragev1alpha1.DiskStateInUse
//				disks[devpath] = disk
//			}
//		}
//	}
//	return disks, nil
//}

//func (mgr *localDiskManager) discoverDisksInUse() (map[string]bool, error) {
//	disks := map[string]bool{}
//	/* e.g.
//	[root@local-storage-10-6-161-17 ~]# blkid
//	/dev/mapper/centos-root: UUID="6b9703de-7c73-42ee-a412-4c4d03de3501" TYPE="xfs"
//	/dev/sda2: UUID="gGiQzu-zdr7-cW4d-D2FA-VQAd-SfeT-9h3IW0" TYPE="LVM2_member"
//	/dev/sdb: UUID="c9ae4157-57e6-4041-acd8-b5161becf5f6" TYPE="xfs"
//	/dev/sda1: UUID="5c71d732-88dc-46b5-8c37-a475075757d4" TYPE="xfs"
//	/dev/sdc: UUID="ipuLhF-A2GS-177J-qM46-YjXh-Uhp2-snXhcr" TYPE="LVM2_member"
//	*/
//	params := exechelper.ExecParams{
//		CmdName: "blkid",
//	}
//	res := mgr.cmdExec.RunCommand(params)
//	if res.ExitCode != 0 {
//		mgr.logger.WithError(res.Error).Error("Failed to execute blkid command")
//		return disks, res.Error
//	}
//	for _, line := range strings.Split(res.OutBuf.String(), "\n") {
//		items := strings.Split(strings.TrimPrefix(line, " "), ":")
//		if len(items) == 0 || !strings.HasPrefix(items[0], "/dev") || strings.Contains(items[0], "mapper") {
//			continue
//		}
//		rightIndex := len(items[0])
//		if strings.HasPrefix(items[0], "/dev/nvme") {
//			diviceSep := strings.LastIndex(items[0], "p")
//			if diviceSep != -1 {
//				rightIndex = diviceSep
//			}
//		} else {
//			for unicode.IsDigit(rune(items[0][rightIndex-1])) {
//				rightIndex--
//			}
//		}
//		disks[items[0][:rightIndex]] = true
//	}
//	mgr.logger.Debugf("Disk %+v in used. count %d.", disks, len(disks))
//	return disks, nil
//}

//// DiscoverAvailableDisks Discover all free disks, including HDD, SSD, NVMe
//func (mgr *localDiskManager) DiscoverAvailableDisks() ([]*localstoragev1alpha1.LocalDisk, error) {
//	availableDisks := []*localstoragev1alpha1.LocalDisk{}
//
//	localDisks, err := mgr.GetLocalDisks()
//	if err != nil {
//		return nil, err
//	}
//
//	for _, disk := range localDisks {
//		if disk.State == localstoragev1alpha1.DiskStateAvailable {
//			availableDisks = append(availableDisks, disk)
//		}
//	}
//
//	return availableDisks, nil
//}

//func (mgr *localDiskManager) discoverDiskCapacities() (map[string]int64, error) {
//	capacities := map[string]int64{}
//
//	/* e.g.
//	[root@local-storage-10-6-161-17 ~]# fdisk -l | grep Disk | grep dev | grep -v "mapper"
//	Disk /dev/sda: 137.4 GB, 137438953472 bytes, 268435456 sectors
//	Disk /dev/sdb: 128.8 GB, 128849018880 bytes, 251658240 sectors
//	Disk /dev/sdc: 107.4 GB, 107374182400 bytes, 209715200 sectors
//	*/
//	params := exechelper.ExecParams{
//		CmdName: "fdisk",
//		CmdArgs: []string{"-l"},
//	}
//	res := mgr.cmdExec.RunCommand(params)
//	if res.ExitCode != 0 {
//		return capacities, res.Error
//	}
//	for _, line := range strings.Split(res.OutBuf.String(), "\n") {
//		if !strings.Contains(line, "Disk") || !strings.Contains(line, "/dev/") || strings.Contains(line, "mapper") {
//			continue
//		}
//		mgr.logger.Debugf("fdisk command output line: %s", line)
//		str := strings.TrimSpace(strings.TrimPrefix(line, "Disk"))
//		items := strings.Split(str, ":")
//		if len(items) < 2 {
//			continue
//		}
//		devPath := items[0]
//		for _, cap := range strings.Split(items[1], ",") {
//			if strings.Contains(cap, "bytes") {
//				capStr := strings.Split(strings.TrimSpace(cap), " ")[0]
//				capInt64, err := strconv.ParseInt(capStr, 10, 64)
//				if err != nil {
//					mgr.logger.WithFields(log.Fields{"disk": devPath, "capacityBytes": capStr}).Errorf("Failed to convert capacity: %s\n", capStr)
//					return capacities, err
//				}
//				capacities[devPath] = capInt64
//			}
//		}
//	}
//
//	return capacities, nil
//}
