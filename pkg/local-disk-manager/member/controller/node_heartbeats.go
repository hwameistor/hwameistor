package controller

import (
	"context"
	"github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	v1 "k8s.io/api/coordination/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"time"
)

const (
	// heartBeatsDuration is same as leaderLeaseDuration
	heartBeatsDuration = 30 * time.Second
)

// startHeartBeatsDetection is responsible for detect if StorageNode can be schedule normally.
//
// if components on the node is offline, mark it as NotReady and scheduler won't consider it
// during pod scheduling period until it becomes Ready again.
func (m *nodeManager) startHeartBeatsDetection(c context.Context) {
	m.logger.WithField("duration", heartBeatsDuration.Seconds()).Info("Start timer to detect node heartbeats")

	detectNodeHeartBeats := func() {
		m.logger.Info("Start detecting node heartbeats")

		diskNodeList := v1alpha1.LocalDiskNodeList{}
		err := m.k8sClient.List(context.TODO(), &diskNodeList)
		if err != nil {
			m.logger.WithError(err).Error("Failed to list LocalDiskNode")
			return
		}

		leaseList := v1.LeaseList{}
		err = m.k8sClient.List(context.TODO(), &leaseList, client.InNamespace(m.namespace))
		if err != nil {
			m.logger.WithError(err).Error("Failed to list worker node Lease")
			return
		}
		workNodeLease := map[string]*v1.Lease{}
		for _, lease := range leaseList.Items {
			if !strings.HasPrefix(lease.Name, "hwameistor-local-disk-manager-worker") {
				continue
			}
			if lease.Spec.HolderIdentity != nil {
				workNodeLease[*lease.Spec.HolderIdentity] = lease.DeepCopy()
			}
		}

		// don't update node status during loop
		needUpdateNodes := map[v1alpha1.State][]v1alpha1.LocalDiskNode{}
		for _, node := range diskNodeList.Items {
			lease, ok := workNodeLease[node.Name]
			if node.Status.State == v1alpha1.NodeStateReady &&
				(!ok || (ok && time.Since(lease.Spec.RenewTime.Time).Seconds() > float64(*lease.Spec.LeaseDurationSeconds))) {
				needUpdateNodes[v1alpha1.NodeStateOffline] = append(needUpdateNodes[v1alpha1.NodeStateOffline], *node.DeepCopy())
			} else if node.Status.State == v1alpha1.NodeStateOffline &&
				ok && time.Since(lease.Spec.RenewTime.Time).Seconds() < float64(*lease.Spec.LeaseDurationSeconds) {
				needUpdateNodes[v1alpha1.NodeStateReady] = append(needUpdateNodes[v1alpha1.NodeStateReady], *node.DeepCopy())
			}
		}

		// mark node offline or ready
		for updateState, updateNodes := range needUpdateNodes {
			for _, node := range updateNodes {
				nodeNew := node.DeepCopy()
				nodeNew.Status.State = updateState
				err = retry.OnError(retry.DefaultRetry, errors.IsTimeout, func() error {
					err = m.k8sClient.Status().Patch(context.TODO(), nodeNew, client.MergeFrom(&node))
					if err != nil && !errors.IsNotFound(err) {
						m.logger.WithField("node", node.Name).WithError(err).Errorf("Failed to mark node as %s", updateState)
						// mock timeout error here to ensure retry anyway
						err = errors.NewTimeoutError(err.Error(), 1)
					}
					return err
				})

				if err != nil {
					m.logger.WithField("node", node.Name).WithError(err).Errorf("Failed to mark node as %s after try 5 times", updateState)
				} else {
					m.logger.WithField("node", node.Name).Errorf("Succeed to mark node as %s", updateState)
				}
			}
		}
	}

	wait.Until(detectNodeHeartBeats, heartBeatsDuration, c.Done())
	m.logger.Info("Stop detecting node heartbeats")
}
