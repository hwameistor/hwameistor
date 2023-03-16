package exporter

import (
	localstorageclientset "github.com/hwameistor/hwameistor/pkg/apis/client/clientset/versioned"
	localstorageinformers "github.com/hwameistor/hwameistor/pkg/apis/client/informers/externalversions"
	localstorageinformersv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/client/informers/externalversions/hwameistor/v1alpha1"
	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type metricsCache struct {
	lsClientset       *localstorageclientset.Clientset
	lsnInformer       localstorageinformersv1alpha1.LocalStorageNodeInformer
	lvInformer        localstorageinformersv1alpha1.LocalVolumeInformer
	lvrInformer       localstorageinformersv1alpha1.LocalVolumeReplicaInformer
	lvMigrateInformer localstorageinformersv1alpha1.LocalVolumeMigrateInformer
	lvConvertInformer localstorageinformersv1alpha1.LocalVolumeConvertInformer
	lvExpandInformer  localstorageinformersv1alpha1.LocalVolumeExpandInformer
	ldInformer        localstorageinformersv1alpha1.LocalDiskInformer
	ldcInformer       localstorageinformersv1alpha1.LocalDiskClaimInformer
	ldvInformer       localstorageinformersv1alpha1.LocalDiskVolumeInformer
}

func newCache() *metricsCache {
	return &metricsCache{}
}

func (mc *metricsCache) run(stopCh <-chan struct{}) {
	log.Debug("start local storage informer factory")
	cfg, err := config.GetConfig()
	if err != nil {
		log.WithError(err).Fatal("Failed to get kubernetes cluster config")
	}

	mc.lsClientset = localstorageclientset.NewForConfigOrDie(cfg)
	lsFactory := localstorageinformers.NewSharedInformerFactory(mc.lsClientset, 0)
	lsFactory.Start(stopCh)

	mc.lsnInformer = lsFactory.Hwameistor().V1alpha1().LocalStorageNodes()
	go mc.lsnInformer.Informer().Run(stopCh)

	mc.lvInformer = lsFactory.Hwameistor().V1alpha1().LocalVolumes()
	go mc.lvInformer.Informer().Run(stopCh)

	mc.lvrInformer = lsFactory.Hwameistor().V1alpha1().LocalVolumeReplicas()
	go mc.lvrInformer.Informer().Run(stopCh)

	mc.lvMigrateInformer = lsFactory.Hwameistor().V1alpha1().LocalVolumeMigrates()
	go mc.lvMigrateInformer.Informer().Run(stopCh)

	mc.lvConvertInformer = lsFactory.Hwameistor().V1alpha1().LocalVolumeConverts()
	go mc.lvConvertInformer.Informer().Run(stopCh)

	mc.lvExpandInformer = lsFactory.Hwameistor().V1alpha1().LocalVolumeExpands()
	go mc.lvExpandInformer.Informer().Run(stopCh)

	mc.ldInformer = lsFactory.Hwameistor().V1alpha1().LocalDisks()
	go mc.ldInformer.Informer().Run(stopCh)

	mc.ldcInformer = lsFactory.Hwameistor().V1alpha1().LocalDiskClaims()
	go mc.ldcInformer.Informer().Run(stopCh)

	mc.ldvInformer = lsFactory.Hwameistor().V1alpha1().LocalDiskVolumes()
	go mc.ldvInformer.Informer().Run(stopCh)

}
