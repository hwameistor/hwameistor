package scheduler

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	framework "k8s.io/kubernetes/pkg/scheduler/framework"
)

type Plugin struct {
	fHandle framework.Handle

	scheduler *Scheduler
}

var _ framework.FilterPlugin = (*Plugin)(nil)

// Name is the name of the plugin used in Registry and configurations.
const (
	Name = "hwameistor-scheduler-plugin"
)

// New initializes a new plugin and returns it.
func New(_ runtime.Object, f framework.Handle) (framework.Plugin, error) {
	time.Sleep(time.Second) // wait for scheduleLabelMgr to be created
	log.SetLevel(log.DebugLevel)

	return &Plugin{
		fHandle:   f,
		scheduler: NewScheduler(f),
	}, nil
}

// Name returns name of the plugin. It is used in logs, etc.
func (p Plugin) Name() string {
	return Name
}

// Filter is the functions invoked by the framework at "filter" extension point.
func (p *Plugin) Filter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, node *framework.NodeInfo) *framework.Status {
	if pod == nil {
		return framework.NewStatus(framework.Unschedulable, "no pod specified")
	}
	if node == nil || node.Node() == nil {
		return framework.NewStatus(framework.Unschedulable, "no node specified")
	}
	log.WithFields(log.Fields{"pod": pod.Name, "node": node.Node().Name}).Debug("filtering a node against a pod")

	if len(pod.Spec.Volumes) == 0 {
		// no pvc, always allowed to be scheduled on this node
		log.Info("no volume in pod's spec, allow it")
		return framework.NewStatus(framework.Success, "no volume to be bound, ok to schedule on any node")
	}

	allowed, err := p.filter(pod, node.Node())
	if err != nil {
		log.WithFields(log.Fields{"pod": pod.Name, "node": node.Node().Name}).WithError(err).Debug("Filtered out the node")
		return framework.NewStatus(framework.Unschedulable, err.Error())
	}

	if allowed {
		log.WithFields(log.Fields{"pod": pod.Name, "node": node.Node().Name}).Debug("Filtered in the node")
		return framework.NewStatus(framework.Success, "can be scheduled on this node")
	}

	log.WithFields(log.Fields{"pod": pod.Name, "node": node.Node().Name}).Debug("Filtered out the node")
	return framework.NewStatus(framework.Unschedulable, "can not be scheduled on this node")
}

// Reserve is the functions invoked by the framework at "reserve" extension point.
func (p *Plugin) Reserve(ctx context.Context, state *framework.CycleState, pod *v1.Pod, node string) *framework.Status {
	if pod == nil {
		return framework.NewStatus(framework.Unschedulable, "no pod specified")
	}
	if node == "" {
		return framework.NewStatus(framework.Unschedulable, "no node specified")
	}
	logCtx := log.Fields{"pod": pod.Name, "node": node}
	log.WithFields(logCtx).Debug("reserving resource for a pod")

	err := p.scheduler.Reserve(pod, node)
	if err != nil {
		log.WithFields(logCtx).WithError(err).Error("failed to reserve resource")
		return framework.NewStatus(framework.Error)
	}

	return framework.NewStatus(framework.Success)
}

// Unreserve is the functions invoked by the framework at "unreserve" extension point.
func (p *Plugin) Unreserve(ctx context.Context, state *framework.CycleState, pod *v1.Pod, node string) {
	if pod == nil {
		log.Debug("no pod specified")
	}
	if node == "" {
		log.Debug("no node specified")
	}
	logCtx := log.Fields{"pod": pod.Name, "node": node}
	log.WithFields(logCtx).Debug("unreserving resource for a pod")

	err := p.scheduler.Unreserve(pod, node)
	if err != nil {
		log.WithFields(logCtx).WithError(err).Error("failed to unreserve resource")
	}
	return
}

func (p *Plugin) filter(pod *v1.Pod, node *v1.Node) (bool, error) {
	return p.scheduler.
		Filter(pod, node)
}
