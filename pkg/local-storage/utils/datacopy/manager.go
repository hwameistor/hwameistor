// Design for copy data from source PVC to destination PVC, continuously push statue into status channel for notifications
package datacopy

import (
	"context"
	"time"

	"github.com/hwameistor/hwameistor/pkg/local-storage/exechelper/nsexecutor"
	log "github.com/sirupsen/logrus"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	DefaultCopyTimeout       = time.Hour * 48
	rcloneMountContainerName = "data-src-mount-container"
)

var (
	logger = log.WithField("module", "util-job")
)

type DataCopyManager struct {
	dataCopyJobStatusAnnotationName string
	statusGenerator                 *statusGenerator
	k8sControllerClient             k8sclient.Client
	ctx                             context.Context
	progressWatchingFunc            func() *Progress
}

// NewDataCopyManager return DataCopyManager instance
//
// It will feedback copy process status continuously through statusCh,
// so it dose not need ResourceReady to poll resource status
func NewDataCopyManager(ctx context.Context,
	dataCopyJobStatusAnnotationName string,
	client k8sclient.Client,
	statusCh chan *DataCopyStatus) (*DataCopyManager, error) {
	dcm := &DataCopyManager{
		dataCopyJobStatusAnnotationName: dataCopyJobStatusAnnotationName,
		k8sControllerClient:             client,
		ctx:                             ctx,
	}

	statusGenerator, err := newStatusGenerator(dcm, dataCopyJobStatusAnnotationName, statusCh)
	if err != nil {
		logger.WithError(err).Error("Failed to init StatusGenerator")
		return nil, err
	}

	dcm.statusGenerator = statusGenerator
	return dcm, nil
}

func (dcm *DataCopyManager) UseRclone(
	rcloneImage,
	rcloneConfigMapName,
	rcloneKeyConfigMapName,
	rcloneConfigMapNamespace,
	rcloneConfigMapKey, rcloneCrtKey string) *Rclone {
	rclone := &Rclone{
		rcloneImage:              rcloneImage,
		rcloneMountContainerName: rcloneMountContainerName,
		rcloneConfigMapName:      rcloneConfigMapName,
		rcloneKeyConfigMapName:   rcloneKeyConfigMapName,
		rcloneConfigMapNamespace: rcloneConfigMapNamespace,
		rcloneConfigMapKey:       rcloneConfigMapKey,
		rcloneCertKey:            rcloneCrtKey,
		skipRcloneConfiguration:  true,
		dcm:                      dcm,
		cmdExec:                  nsexecutor.New(),
	}

	dcm.progressWatchingFunc = rclone.progressWatchingFunc

	return rclone
}

func (dcm *DataCopyManager) Run() {
	logger.Debugf("DataCopyManager Run start")
	dcm.statusGenerator.Run()
}

func (dcm *DataCopyManager) RegisterRelatedJob(jobName string, resultCh chan *DataCopyStatus) {
	dcm.statusGenerator.relatedJobWithResultCh[jobName] = resultCh
}

func (dcm *DataCopyManager) DeregisterRelatedJob(jobName string) {
	delete(dcm.statusGenerator.relatedJobWithResultCh, jobName)
}
