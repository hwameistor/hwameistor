package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	localstorageclientset "github.com/hwameistor/hwameistor/pkg/apis/client/clientset/versioned"
	localstorageinformers "github.com/hwameistor/hwameistor/pkg/apis/client/informers/externalversions"
	localstorageinformersv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/client/informers/externalversions/hwameistor/v1alpha1"
	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
)

type MetricsCollector struct {
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

	//lsnLister localstoragelistersv1alpha1.LocalStorageNodeLister

	// for localstoragenode
	localStorageNodeCapacityMetrics    *prometheus.GaugeVec
	localStorageNodeVolumeCountMetrics *prometheus.GaugeVec
	localStorageNodeStatusMetrics      *prometheus.GaugeVec
	// for localvolume
	localVolumeTypeMetrics     *prometheus.GaugeVec
	localVolumeStatusMetrics   *prometheus.GaugeVec
	localVolumeCapacityMetrics *prometheus.GaugeVec
	// for localvolumereplica
	localVolumeReplicaStatusMetrics   *prometheus.GaugeVec
	localVolumeReplicaCapacityMetrics *prometheus.GaugeVec
	// for localdisk
	localDiskStatusMetrics *prometheus.GaugeVec
	// for localdiskvolume
	localDiskVolumeStatusMetrics *prometheus.GaugeVec
	// for localvolumemigrate
	// for localvolumeconvert
	// for localvolumeexpand

	// for S.M.A.R.T
	smartCollector prometheus.Collector
	metricsHandler http.Handler
}

func NewHandler() *MetricsCollector {
	return &MetricsCollector{
		// for localstoragenode
		localStorageNodeCapacityMetrics: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "hwameistor_localstoragenode_capacity",
				Help: "The storage capacity of the localstoragenode.",
			},
			[]string{"nodeName", "poolName", "kind"},
		),
		localStorageNodeVolumeCountMetrics: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "hwameistor_localstoragenode_volumecount",
				Help: "The volume count of the localstoragenode.",
			},
			[]string{"nodeName", "poolName", "kind"},
		),
		localStorageNodeStatusMetrics: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "hwameistor_localstoragenode_status",
				Help: "The status summary of the localstoragenode.",
			},
			[]string{"status"},
		),

		// for localvolume
		localVolumeTypeMetrics: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "hwameistor_localvolume_type",
				Help: "The type of the localvolume.",
			},
			[]string{"poolName", "type"},
		),
		localVolumeStatusMetrics: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "hwameistor_localvolume_status",
				Help: "The status summary of the localvolume.",
			},
			[]string{"poolName", "type", "status"},
		),
		localVolumeCapacityMetrics: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "hwameistor_localvolume_capacity",
				Help: "The capacity of the localvolume.",
			},
			[]string{"poolName", "volumeName", "type"},
		),

		// for localvolumereplica
		localVolumeReplicaStatusMetrics: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "hwameistor_localvolumereplica_status",
				Help: "The status summary of the localvolumereplica.",
			},
			[]string{"nodeName", "poolName", "status"},
		),
		localVolumeReplicaCapacityMetrics: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "hwameistor_localvolumereplica_capacity",
				Help: "The capacity of the localvolumereplica.",
			},
			[]string{"nodeName", "poolName", "volumeName"},
		),

		// for localdisk
		localDiskStatusMetrics: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "hwameistor_localdisk_status",
				Help: "The status summary of the localdisk.",
			},
			[]string{"nodeName", "type", "status"},
		),

		// for localdiskvolume
		localDiskVolumeStatusMetrics: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "hwameistor_localdiskvolume_status",
				Help: "The status summary of the localdiskvolume.",
			},
			[]string{"nodeName", "type", "status"},
		),

		// for S.M.A.R.T
		smartCollector: NewSMARTCollector(),
	}
}

func (mc *MetricsCollector) Run(stopCh <-chan struct{}) {

	mc.registerMetrics()

	mc.setup(stopCh)

}

func (mc *MetricsCollector) ServeHTTP(resp http.ResponseWriter, req *http.Request) {

	mc.collect()

	mc.metricsHandler.ServeHTTP(resp, req)
}

func (mc *MetricsCollector) registerMetrics() {
	registry := prometheus.NewRegistry()

	// for localstoragenode
	registry.MustRegister(mc.localStorageNodeCapacityMetrics)
	registry.MustRegister(mc.localStorageNodeVolumeCountMetrics)
	registry.MustRegister(mc.localStorageNodeStatusMetrics)
	// for localvolume
	registry.MustRegister(mc.localVolumeTypeMetrics)
	registry.MustRegister(mc.localVolumeStatusMetrics)
	registry.MustRegister(mc.localVolumeCapacityMetrics)
	// for localvolumereplica
	registry.MustRegister(mc.localVolumeReplicaStatusMetrics)
	registry.MustRegister(mc.localVolumeReplicaCapacityMetrics)
	// for localdisk
	registry.MustRegister(mc.localDiskStatusMetrics)
	// for localdiskvolume
	registry.MustRegister(mc.localDiskVolumeStatusMetrics)
	// for S.M.A.R.T
	registry.MustRegister(mc.smartCollector)

	mc.metricsHandler = promhttp.HandlerFor(registry, promhttp.HandlerOpts{})

}

func (mc *MetricsCollector) setup(stopCh <-chan struct{}) {
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

func (mc *MetricsCollector) collect() {
	mc.collectForLocalStorageNodes()
	mc.collectForLocalVolumes()
	mc.collectForLocalVolumeReplicas()
	mc.collectForLocalVolumeMigrates()
	mc.collectForLocalVolumeConverts()
	mc.collectForLocalVolumeExpands()
	mc.collectForLocalDisks()
	mc.collectForLocalDiskClaims()
	mc.collectForLocalDiskVolumes()
}

func (mc *MetricsCollector) collectForLocalStorageNodes() {
	log.Debug("Collecting metrics for LocalStorageNode ...")
	nodes, err := mc.lsnInformer.Lister().List(labels.NewSelector())
	if err != nil || len(nodes) == 0 {
		log.WithError(err).Debug("Not found LocalStorageNode")
		return
	}
	nodeCount := map[string]int64{}
	nodeCount[string(apisv1alpha1.NodeStateReady)] = 0
	nodeCount[string(apisv1alpha1.NodeStateMaintain)] = 0
	nodeCount[string(apisv1alpha1.NodeStateOffline)] = 0

	for _, node := range nodes {
		nodeCount[string(node.Status.State)]++

		for poolName, pool := range node.Status.Pools {
			mc.localStorageNodeCapacityMetrics.WithLabelValues(
				node.Name,
				unifiedPoolName(poolName),
				"Total",
			).Set(float64(pool.TotalCapacityBytes))
			mc.localStorageNodeCapacityMetrics.WithLabelValues(
				node.Name,
				unifiedPoolName(poolName),
				"Free",
			).Set(float64(pool.FreeCapacityBytes))
			mc.localStorageNodeCapacityMetrics.WithLabelValues(
				node.Name,
				unifiedPoolName(poolName),
				"Used",
			).Set(float64(pool.UsedCapacityBytes))

			mc.localStorageNodeVolumeCountMetrics.WithLabelValues(
				node.Name,
				unifiedPoolName(poolName),
				"Total",
			).Set(float64(pool.TotalVolumeCount))
			mc.localStorageNodeVolumeCountMetrics.WithLabelValues(
				node.Name,
				unifiedPoolName(poolName),
				"Free",
			).Set(float64(pool.FreeVolumeCount))
			mc.localStorageNodeVolumeCountMetrics.WithLabelValues(
				node.Name,
				unifiedPoolName(poolName),
				"Used",
			).Set(float64(pool.UsedVolumeCount))
		}
	}

	for state, count := range nodeCount {
		mc.localStorageNodeStatusMetrics.WithLabelValues(state).Set(float64(count))
	}

}

func (mc *MetricsCollector) collectForLocalVolumes() {
	log.Debug("Collecting metrics for LocalVolume ...")
	volumes, err := mc.lvInformer.Lister().List(labels.NewSelector())
	if err != nil || len(volumes) == 0 {
		log.WithError(err).Debug("Not found LocalVolume")
		return
	}
	volumeTypeCount := map[string]map[string]int64{
		apisv1alpha1.DiskClassNameHDD: {
			VolumeTypeNonHA:       0,
			VolumeTypeConvertible: 0,
			VolumeTypeHA:          0,
		},
		apisv1alpha1.DiskClassNameSSD: {
			VolumeTypeNonHA:       0,
			VolumeTypeConvertible: 0,
			VolumeTypeHA:          0,
		},
		apisv1alpha1.DiskClassNameNVMe: {
			VolumeTypeNonHA:       0,
			VolumeTypeConvertible: 0,
			VolumeTypeHA:          0,
		},
	}
	volumeStatusCount := map[string]map[string]map[string]int64{
		apisv1alpha1.DiskClassNameHDD: {
			VolumeTypeNonHA:       {},
			VolumeTypeHA:          {},
			VolumeTypeConvertible: {},
		},
		apisv1alpha1.DiskClassNameSSD: {
			VolumeTypeNonHA:       {},
			VolumeTypeHA:          {},
			VolumeTypeConvertible: {},
		},
		apisv1alpha1.DiskClassNameNVMe: {
			VolumeTypeNonHA:       {},
			VolumeTypeHA:          {},
			VolumeTypeConvertible: {},
		},
	}
	for _, vol := range volumes {
		poolName := unifiedPoolName(vol.Spec.PoolName)
		if !vol.Spec.Convertible {
			volumeTypeCount[poolName][VolumeTypeNonHA]++
			volumeStatusCount[poolName][VolumeTypeNonHA][string(vol.Status.State)]++
		} else if vol.Spec.ReplicaNumber > 1 {
			volumeTypeCount[poolName][VolumeTypeHA]++
			volumeStatusCount[poolName][VolumeTypeHA][string(vol.Status.State)]++
		} else {
			volumeTypeCount[poolName][VolumeTypeConvertible]++
			volumeStatusCount[poolName][VolumeTypeConvertible][string(vol.Status.State)]++
		}
		mc.localVolumeCapacityMetrics.WithLabelValues(poolName, vol.Name, "Allocated").Set(float64(vol.Status.AllocatedCapacityBytes))
		mc.localVolumeCapacityMetrics.WithLabelValues(poolName, vol.Name, "Used").Set(float64(vol.Status.UsedCapacityBytes))
	}
	for poolName, volumeCount := range volumeTypeCount {
		for volType, count := range volumeCount {
			mc.localVolumeTypeMetrics.WithLabelValues(poolName, volType).Set(float64(count))
		}
	}
	for poolName, typeCount := range volumeStatusCount {
		for volType, statusCount := range typeCount {
			for state, count := range statusCount {
				mc.localVolumeStatusMetrics.WithLabelValues(poolName, volType, state).Set(float64(count))
			}
		}
	}
}

func (mc *MetricsCollector) collectForLocalVolumeReplicas() {
	log.Debug("Collecting metrics for LocalVolumeReplica ...")
	replicas, err := mc.lvrInformer.Lister().List(labels.NewSelector())
	if err != nil || len(replicas) == 0 {
		log.WithError(err).Debug("Not found LocalVolumeReplica")
		return
	}

	replicaStatusCount := map[string]map[string]map[string]int64{
		apisv1alpha1.DiskClassNameHDD:  {},
		apisv1alpha1.DiskClassNameSSD:  {},
		apisv1alpha1.DiskClassNameNVMe: {},
	}

	for _, replica := range replicas {
		poolName := unifiedPoolName(replica.Spec.PoolName)
		_, exist := replicaStatusCount[poolName][replica.Spec.NodeName]
		if !exist {
			replicaStatusCount[poolName][replica.Spec.NodeName] = map[string]int64{}
		}
		replicaStatusCount[poolName][replica.Spec.NodeName][string(replica.Status.State)]++

		mc.localVolumeReplicaCapacityMetrics.WithLabelValues(replica.Spec.NodeName, poolName, replica.Spec.VolumeName).Set(float64(replica.Status.AllocatedCapacityBytes))
	}

	for poolName, nodeCount := range replicaStatusCount {
		for nodeName, stateCount := range nodeCount {
			for state, count := range stateCount {
				mc.localVolumeReplicaStatusMetrics.WithLabelValues(nodeName, poolName, state).Set(float64(count))
			}
		}

	}
}

func (mc *MetricsCollector) collectForLocalVolumeMigrates() {
}

func (mc *MetricsCollector) collectForLocalVolumeConverts() {
}

func (mc *MetricsCollector) collectForLocalVolumeExpands() {
}

func (mc *MetricsCollector) collectForLocalDisks() {
}

func (mc *MetricsCollector) collectForLocalDiskClaims() {
}

func (mc *MetricsCollector) collectForLocalDiskVolumes() {
}
