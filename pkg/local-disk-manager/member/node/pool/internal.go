package pool

import (
	"fmt"
	"github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/member/types"
	log "github.com/sirupsen/logrus"
	"k8s.io/kubernetes/pkg/volume/util/hostutil"
	"os"
	"strings"
)

// diskPool implement interface pool.Manager
type diskPool struct {
	name string
	hu   hostutil.HostUtils
}

func (p *diskPool) Init() error {
	for _, poolType := range types.DefaultDevTypes {
		if err := p.CreatePool(poolType); err != nil {
			return err
		}
	}
	return nil
}

func (p *diskPool) CreatePool(poolType types.DevType) error {
	for _, poolPath := range []string{types.GetLocalDiskPoolPath(poolType), types.GetPoolDiskPath(poolType), types.GetPoolVolumePath(poolType)} {
		exist, _ := p.hu.PathExists(poolPath)
		if !exist {
			if err := os.MkdirAll(poolPath, 0755); err != nil {
				return err
			}
			log.Debugf("Succeed to create %s(mode: 0755) directory", poolPath)
			continue
		}
		log.Debugf("Directory %s already exist", poolPath)
	}
	return nil
}

func (p *diskPool) PoolExist(poolName string) (bool, error) {
	return p.hu.PathExists(types.GetLocalDiskPoolPath(poolName))
}

func (p *diskPool) GetPool(poolName string) (*Pool, error) {
	//TODO implement me
	panic("implement me")
}

func (p *diskPool) ExtendPool(poolName string, devLinks []string, serial string) (bool, error) {
	actualDeviceLink, err := findSuitableDevLink(devLinks, serial)
	if err != nil {
		return false, err
	}

	devName := strings.Split(actualDeviceLink, "/")[len(strings.Split(actualDeviceLink, "/"))-1]
	poolDevicePath := types.ComposePoolDevicePath(poolName, devName)
	exist, err := p.hu.PathExists(devName)
	if exist || err != nil {
		return exist, err
	}

	log.Infof("create symlink %s point to %s", poolDevicePath, actualDeviceLink)
	err = os.Symlink(actualDeviceLink, poolDevicePath)
	if err != nil {
		return false, err
	}

	return true, nil
}

func convertToDevLinkMap(devLinks []string) map[v1alpha1.DevLinkType][]string {
	devLinksMap := make(map[v1alpha1.DevLinkType][]string, 0)
	for _, devLink := range devLinks {
		switch {
		case strings.Contains(devLink, v1alpha1.LinkByPath):
			devLinksMap[v1alpha1.LinkByPath] = append(devLinksMap[v1alpha1.LinkByPath], devLink)
		case strings.Contains(devLink, v1alpha1.LinkByID):
			devLinksMap[v1alpha1.LinkByID] = append(devLinksMap[v1alpha1.LinkByID], devLink)
		case strings.Contains(devLink, v1alpha1.LinkByUUID):
			devLinksMap[v1alpha1.LinkByUUID] = append(devLinksMap[v1alpha1.LinkByUUID], devLink)
		default:
			continue
		}
	}
	return devLinksMap
}

func findSuitableDevLink(devLinks []string, serial string) (string, error) {
	devLinkMap := convertToDevLinkMap(devLinks)

	deviceLink := ""
	// use dev link order: by-id(serial) -> by-path
	if serial != "" && devLinkMap[v1alpha1.LinkByID] != nil {
		linksByID := devLinkMap[v1alpha1.LinkByID]
		for _, link := range linksByID {
			if strings.HasSuffix(link, serial) {
				deviceLink = link
				break
			}
		}
	} else if devLinkMap[v1alpha1.LinkByPath] != nil {
		linksByPath := devLinkMap[v1alpha1.LinkByPath]
		if len(linksByPath) == 0 {
			return "", fmt.Errorf("this device does not exist by-id(serial) or by-path symlink, devLinks: %v, serial: %s", devLinks, serial)
		}
		deviceLink = linksByPath[0]
	}

	return deviceLink, nil
}

func New() Manager {
	return &diskPool{
		hu: hostutil.NewHostUtil(),
	}
}

// /dev/sdc -> sda
func getDeviceName(devPath string) string {
	ss := strings.Split(devPath, "/")
	return ss[len(ss)-1]
}
