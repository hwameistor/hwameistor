package datacopy

import (
	"context"
	"encoding/json"
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	k8sinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	k8scache "k8s.io/client-go/tools/cache"
	k8sworkqueue "k8s.io/client-go/util/workqueue"
	k8sruntimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	k8sconfig "sigs.k8s.io/controller-runtime/pkg/client/config"
)

const (
	DataCopyStatusPending = "pending"
	DataCopyStatusSuccess = "success"
	DataCopyStatusRunning = "running"
	DataCopyStatusFailed  = "failed"
)

// TODO
type Progress struct{}

type DataCopyStatus struct {
	UserData string
	JobName  string
	Phase    string
	Event    string
	Message  string
	// TODO
	Progress *Progress
}

type statusGenerator struct {
	dcm                          *DataCopyManager
	dataCopyStatusAnnotationName string
	statusGenerator              k8scache.SharedIndexInformer
	queue                        k8sworkqueue.Interface
	statusCh                     chan *DataCopyStatus
	// map[jobName]chan *DataCopyStatus
	relatedJobWithResultCh map[string]chan *DataCopyStatus
}

func newStatusGenerator(dcm *DataCopyManager,
	dataCopyStatusAnnotationName string,
	statusCh chan *DataCopyStatus) (*statusGenerator, error) {
	statusGenerator := &statusGenerator{
		dcm:                          dcm,
		dataCopyStatusAnnotationName: dataCopyStatusAnnotationName,
		queue:                        k8sworkqueue.New(),
		statusCh:                     statusCh,
		relatedJobWithResultCh:       make(map[string]chan *DataCopyStatus),
	}

	config, err := k8sconfig.GetConfig()
	if err != nil {
		return nil, err
	}

	k8sClientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	factory := k8sinformers.NewSharedInformerFactory(k8sClientset, 0)
	informer := factory.Batch().V1().Jobs().Informer()
	informer.AddEventHandler(k8scache.ResourceEventHandlerFuncs{
		AddFunc:    statusGenerator.onAdd,
		UpdateFunc: statusGenerator.onUpdate,
		//DeleteFunc: statusGenerator.onDelete,
	})

	factory.Start(dcm.ctx.Done())
	factory.WaitForCacheSync(dcm.ctx.Done())
	statusGenerator.statusGenerator = informer

	return statusGenerator, nil
}

type jobWithEvent struct {
	Job   *batchv1.Job
	Event string
}

func (statusGenerator *statusGenerator) onAdd(obj interface{}) {
	job := obj.(*batchv1.Job)

	//if !statusGenerator.isReleted(job) {
	//	return
	//}

	eventJob := &jobWithEvent{
		Job:   job,
		Event: "ADD",
	}

	statusGenerator.queue.Add(eventJob)
}

func (statusGenerator *statusGenerator) onUpdate(_, obj interface{}) {
	job := obj.(*batchv1.Job)

	//if !statusGenerator.isReleted(job) {
	//	return
	//}

	eventJob := &jobWithEvent{
		Job:   job,
		Event: "UPDATE",
	}

	statusGenerator.queue.Add(eventJob)
}

func (statusGenerator *statusGenerator) onDelete(obj interface{}) {
	job := obj.(*batchv1.Job)

	//if !statusGenerator.isReleted(job) {
	//	return
	//}

	eventJob := &jobWithEvent{
		Job:   job,
		Event: "DELETE",
	}

	statusGenerator.queue.Done(eventJob)
}

func (statusGenerator *statusGenerator) Run() {
	logger.Debugf("statusGenerator Run")
	go wait.UntilWithContext(statusGenerator.dcm.ctx, statusGenerator.processLoop, 5)
}

func (statusGenerator *statusGenerator) processLoop(ctx context.Context) {

	untyped, qClosed := statusGenerator.queue.Get()
	if qClosed {
		logger.Fatal("Unexpcted queue close")
	}
	statusGenerator.queue.Done(untyped)
	jobNeedProcess := untyped.(*jobWithEvent)
	logger.Infof(
		"Start processing job:%s, namespace:%s, event:%s, qlen: %d",
		jobNeedProcess.Job.Name,
		jobNeedProcess.Job.Namespace,
		jobNeedProcess.Event,
		statusGenerator.queue.Len(),
	)

	runningStatus := statusGenerator.getJobRunningStatus(jobNeedProcess.Job)
	if runningStatus == nil {
		logger.Debugf(
			"Rollback runing status not exists on %s, namespace %s",
			jobNeedProcess.Job.Name,
			jobNeedProcess.Job.Namespace,
		)
		return
	}
	switch jobNeedProcess.Event {
	case "ADD":
		logger.Debugf("Job %s running", jobNeedProcess.Job.Name)
		runningStatus.Phase = DataCopyStatusRunning
		statusGenerator.statusCh <- runningStatus
	case "UPDATE":
		status, event := calculateJobRunningStatus(jobNeedProcess.Job)
		if status == DataCopyStatusSuccess {
			logger.Debugf("Job %s successful finished", jobNeedProcess.Job.Name)
			runningStatus.Phase = DataCopyStatusSuccess
			runningStatus.Event = event
			statusGenerator.gc(jobNeedProcess.Job, runningStatus)
			statusGenerator.statusCh <- runningStatus
			if resCh, exists := statusGenerator.relatedJobWithResultCh[jobNeedProcess.Job.Name]; exists {
				resCh <- runningStatus
			}
		} else if status == DataCopyStatusFailed {
			logger.Debugf("Job %s failed", jobNeedProcess.Job.Name)
			runningStatus.Phase = DataCopyStatusFailed
			runningStatus.Event = event
			// TODO error message
			statusGenerator.statusCh <- runningStatus
			if resCh, exists := statusGenerator.relatedJobWithResultCh[jobNeedProcess.Job.Name]; exists {
				resCh <- runningStatus
			}
		}
	}
}

func (statusGenerator *statusGenerator) gc(job *batchv1.Job, runningStatus *DataCopyStatus) {
	// delete job
	if err := statusGenerator.delObject(
		job.Name,
		job.Namespace,
		&batchv1.Job{},
	); err != nil {
		//statusGenerator.queue.Add(untyped)
		logger.WithError(err).Errorf(
			"Failed to delete Job %s, namesapce %s",
			job.Name,
			job.Namespace,
		)
		return
	}

	// delete job pod, Surprisely it was alive after delete job
	jobPods := &corev1.PodList{}
	err := statusGenerator.dcm.k8sControllerClient.List(statusGenerator.dcm.ctx, jobPods)
	if err != nil {
		logger.WithError(err).Errorf(
			"Failed to list job, job name %s, namesapce %s",
			job.Name,
			job.Namespace,
		)
		return
	}
	for _, pod := range jobPods.Items {
		if pod.Labels["job-name"] != job.Name {
			continue
		}
		logger.Debugf("Start deleting pod %s, namesapce %s", pod.Name, pod.Namespace)
		if err := statusGenerator.delObject(
			pod.Name,
			pod.Namespace,
			&corev1.Pod{},
		); err != nil {
			//statusGenerator.queue.Add(untyped)
			logger.WithError(err).Errorf("Failed to delete Job pod %s, namesapce %s", pod.Name, pod.Namespace)
			return
		}
	}
}

func calculateJobRunningStatus(job *batchv1.Job) (string, string) {
	if job.Status.Active != 0 {
		return DataCopyStatusRunning, ""
	} else if job.Status.Failed != 0 {
		event := "Unknown"
		if len(job.Status.Conditions) > 0 {
			event = fmt.Sprintf("Message: %s, Reason:%s", job.Status.Conditions[0].Message, job.Status.Conditions[0].Reason)
		}
		return DataCopyStatusFailed, event
	} else if job.Status.Succeeded != 0 {
		return DataCopyStatusSuccess, ""
	}
	return DataCopyStatusPending, ""
}

func (statusGenerator *statusGenerator) delObject(name, namespace string, obj k8sruntime.Object) error {
	objectKey := k8sruntimeclient.ObjectKey{
		Namespace: namespace,
		Name:      name,
	}
	if err := statusGenerator.dcm.k8sControllerClient.Get(statusGenerator.dcm.ctx, objectKey, obj); err != nil {
		if k8serrors.IsNotFound(err) {
			logger.Debugf("Object %s not found, skip deleting, namespace %s", name, namespace)
			return nil
		}
		logger.WithError(err).Error("Failed to get instance")
		return err
	}

	if err := statusGenerator.dcm.k8sControllerClient.Delete(statusGenerator.dcm.ctx, obj); err != nil {
		logger.WithError(err).Error("Failed to delete instance")
		return err
	}

	return nil
}

// TODO This func should as a filter handler in front of the informer
func (statusGenerator *statusGenerator) isReleted(job *batchv1.Job) bool {
	if dataCopyStatusJSON, has := job.Annotations[statusGenerator.dataCopyStatusAnnotationName]; has {
		dataCopyStatus := &DataCopyStatus{}
		if err := json.Unmarshal([]byte(dataCopyStatusJSON), dataCopyStatus); err != nil {
			return false
		}
		if _, exists := statusGenerator.relatedJobWithResultCh[dataCopyStatus.JobName]; exists {
			return true
		}
	}

	return false
}

func (statusGenerator *statusGenerator) getJobRunningStatus(job *batchv1.Job) *DataCopyStatus {
	rcloneRunningStatus := &DataCopyStatus{}
	if untyped, has := job.Annotations[statusGenerator.dataCopyStatusAnnotationName]; has {
		json.Unmarshal([]byte(untyped), rcloneRunningStatus)
	}
	return rcloneRunningStatus
}
