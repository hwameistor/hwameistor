package auditor

import (
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	runtimecache "sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	localstorageclientset "github.com/hwameistor/hwameistor/pkg/apis/client/clientset/versioned"
	localstorageinformers "github.com/hwameistor/hwameistor/pkg/apis/client/informers/externalversions"
)

// Auditor interface
type Auditor interface {
	Run(informerCache runtimecache.Cache, stopCh <-chan struct{}) error
}

type auditor struct {
}

// New an assistant instance
func New(clientset *kubernetes.Clientset) Auditor {
	return &auditor{}
}

func (ad *auditor) Run(informersCache runtimecache.Cache, stopCh <-chan struct{}) error {

	// Initialize HwameiStor LocalStorage resources
	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		log.WithError(err).Fatal("Failed to get kubernetes cluster config")
	}

	log.Debug("start local storage informer factory")
	lsClientSet := localstorageclientset.NewForConfigOrDie(cfg)
	lsFactory := localstorageinformers.NewSharedInformerFactory(lsClientSet, 0)
	lsFactory.Start(stopCh)

	eventStore := NewEventStore()
	eventStore.Run(lsClientSet, lsFactory, stopCh)

	newAuditorForCluster(eventStore).Run(informersCache, stopCh)

	newAuditorForLocalStorageNode(eventStore).Run(lsFactory, stopCh)

	newAuditorForLocalVolume(eventStore).Run(lsFactory, stopCh)
	newAuditorForLocalVolumeMigrate(eventStore).Run(lsFactory, stopCh)
	newAuditorForLocalVolumeConvert(eventStore).Run(lsFactory, stopCh)
	newAuditorForLocalVolumeExpand(eventStore).Run(lsFactory, stopCh)

	newAuditorForLocalDisk(eventStore).Run(lsFactory, stopCh)

	<-stopCh
	return nil
}
