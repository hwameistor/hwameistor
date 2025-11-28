package manager

import (
	"testing"
)

func TestDiskInfo_GenerateUUID(t *testing.T) {
	testCases := []struct {
		Description string
		DiskInfoA   *DiskInfo
		DiskInfoB   *DiskInfo
		ExpectEqual bool
	}{
		{
			Description: "These are the different sets of device links",
			DiskInfoA: &DiskInfo{
				Attribute: Attribute{
					DevLinks: []string{
						"/dev/disk/by-diskseq/9",
						"/dev/disk/by-path/pci-0000:00:07.0",
						"/dev/disk/by-path/virtio-pci-0000:00:07.0",
					},
				},
			},
			DiskInfoB: &DiskInfo{
				Attribute: Attribute{
					DevLinks: []string{
						"/dev/disk/by-diskseq/10",
						"/dev/disk/by-id/lvm-pv-uuid-j0SJMZ-AdTs-wE89-CDCC-zx0T-qXaL-18ZkFA",
						"/dev/disk/by-path/pci-0000:00:0a.0",
						"/dev/disk/by-path/virtio-pci-0000:00:0a.0",
					},
				},
			},
			ExpectEqual: false,
		},
		{
			Description: "This is the same order of `by-path` device links",
			DiskInfoA: &DiskInfo{
				Attribute: Attribute{
					DevLinks: []string{
						"/dev/disk/by-path/virtio-pci-0000:00:0a.0",
						"/dev/disk/by-diskseq/10",
						"/dev/disk/by-path/pci-0000:00:0a.0",
						"/dev/disk/by-id/lvm-pv-uuid-j0SJMZ-AdTs-wE89-CDCC-zx0T-qXaL-18ZkFA",
					},
				},
			},
			DiskInfoB: &DiskInfo{
				Attribute: Attribute{
					DevLinks: []string{
						"/dev/disk/by-path/virtio-pci-0000:00:0a.0",
						"/dev/disk/by-diskseq/10",
						"/dev/disk/by-id/lvm-pv-uuid-j0SJMZ-AdTs-wE89-CDCC-zx0T-qXaL-18ZkFA",
						"/dev/disk/by-path/pci-0000:00:0a.0",
					},
				},
			},
			ExpectEqual: true,
		},
		{
			Description: "This is the different order of `by-path` device links",
			DiskInfoA: &DiskInfo{
				Attribute: Attribute{
					DevLinks: []string{
						"/dev/disk/by-path/virtio-pci-0000:00:0a.0",
						"/dev/disk/by-path/pci-0000:00:0a.0",
					},
				},
			},
			DiskInfoB: &DiskInfo{
				Attribute: Attribute{
					DevLinks: []string{
						"/dev/disk/by-path/pci-0000:00:0a.0",
						"/dev/disk/by-path/virtio-pci-0000:00:0a.0",
					},
				},
			},
			ExpectEqual: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Description, func(t *testing.T) {
			uuidA := testCase.DiskInfoA.GenerateUUID()
			uuidB := testCase.DiskInfoB.GenerateUUID()
			if uuidA == uuidB != testCase.ExpectEqual {
				t.Fatal("Unexpected UUIDs equality relation")
			}
		})
	}
}
