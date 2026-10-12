//go:build scheduler_integration

package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	kubefake "k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/events"
	volumehelpers "k8s.io/component-helpers/storage/volume"
	kubescheduler "k8s.io/kubernetes/pkg/scheduler"
	config "k8s.io/kubernetes/pkg/scheduler/apis/config"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins/feature"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins/volumebinding"
	frameworkruntime "k8s.io/kubernetes/pkg/scheduler/framework/runtime"
)

// Run explicitly with -tags=scheduler_integration. These tests run the real
// scheduler queue, scheduling/binding cycles, and VolumeBinding plugin against
// an in-memory API. A controlled provisioner completes PVC binding after a
// wall-clock delay; no CSI service or cluster resources are modified.
type slowSchedulingCounts struct {
	cycles, filters, newChecks, prebind, prebindErrors, binds, pvcUpdates atomic.Int64
	bindNanos                                                             atomic.Int64
	started                                                               time.Time
	t                                                                     *testing.T
}

func (c *slowSchedulingCounts) snapshot(label string) {
	data, _ := json.Marshal(map[string]interface{}{
		"case": c.t.Name(), "version": os.Getenv("HWAMEISTOR_TEST_VERSION"),
		"label": label, "elapsed_seconds": time.Since(c.started).Seconds(),
		"cycles": c.cycles.Load(), "hwameistor_filters": c.filters.Load(),
		"new_capacity_checks": c.newChecks.Load(), "prebind_calls": c.prebind.Load(),
		"prebind_errors": c.prebindErrors.Load(), "pod_binds": c.binds.Load(),
		"pvc_updates": c.pvcUpdates.Load(), "pod_bound_after_seconds": float64(c.bindNanos.Load()) / float64(time.Second),
	})
	c.t.Log(string(data))
}

type observedVolumeBinding struct {
	*volumebinding.VolumeBinding
	counts *slowSchedulingCounts
}

func (p *observedVolumeBinding) Name() string { return "ObservedVolumeBinding" }
func (p *observedVolumeBinding) PreBind(ctx context.Context, state *framework.CycleState, pod *corev1.Pod, node string) *framework.Status {
	p.counts.prebind.Add(1)
	p.counts.snapshot("prebind-enter")
	status := p.VolumeBinding.PreBind(ctx, state, pod, node)
	if !status.IsSuccess() {
		p.counts.prebindErrors.Add(1)
		p.counts.t.Logf("PreBind failed: %s", status.Message())
	}
	p.counts.snapshot("prebind-return")
	return status
}

type observedHwameistor struct {
	*Plugin
	counts *slowSchedulingCounts
}

func (p *observedHwameistor) Filter(ctx context.Context, state *framework.CycleState, pod *corev1.Pod, node *framework.NodeInfo) *framework.Status {
	p.counts.filters.Add(1)
	return p.Plugin.Filter(ctx, state, pod, node)
}

func TestSlowProvisioningScheduling(t *testing.T) {
	seconds := 180
	if raw := os.Getenv("HWAMEISTOR_TEST_DELAY_SECONDS"); raw != "" {
		var err error
		seconds, err = strconv.Atoi(raw)
		if err != nil || seconds < 6 {
			t.Fatal("HWAMEISTOR_TEST_DELAY_SECONDS must be at least 6")
		}
	}
	delay := time.Duration(seconds) * time.Second
	fixed := os.Getenv("HWAMEISTOR_TEST_VERSION") != "before"
	for _, tc := range []struct {
		name                                     string
		selected, capacityConsumed, shortTimeout bool
	}{
		{name: "fresh"},
		{name: "reentry", selected: true},
		{name: "reentry-capacity-used", selected: true, capacityConsumed: true},
		{name: "timeout", shortTimeout: true},
	} {
		tc := tc // This repository uses Go 1.21 loop-variable semantics.
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithTimeout(context.Background(), delay+60*time.Second)
			defer cancel()
			counts := &slowSchedulingCounts{started: time.Now(), t: t}
			pvc := testPVC("claim", "lvm.hwameistor.io", "", "", corev1.ClaimPending)
			pvc.ResourceVersion = "1"
			pvc.UID = "claim-uid"
			pvc.Spec.AccessModes = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
			delete(pvc.Annotations, volumehelpers.AnnSelectedNode)
			if tc.selected {
				pvc.Annotations[volumehelpers.AnnSelectedNode] = "node-a"
			}
			mode := storagev1.VolumeBindingWaitForFirstConsumer
			sc := &storagev1.StorageClass{
				ObjectMeta:  metav1.ObjectMeta{Name: *pvc.Spec.StorageClassName, ResourceVersion: "1"},
				Provisioner: "lvm.hwameistor.io", VolumeBindingMode: &mode,
			}
			cs := kubefake.NewSimpleClientset(pvc, sc,
				&corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node-a", UID: "node-a"}},
				&corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node-b", UID: "node-b"}},
				&storagev1.CSINode{ObjectMeta: metav1.ObjectMeta{Name: "node-a"}},
				&storagev1.CSINode{ObjectMeta: metav1.ObjectMeta{Name: "node-b"}},
			)
			var rv atomic.Int64
			rv.Store(1)
			cs.PrependReactor("update", "persistentvolumeclaims", func(a ktesting.Action) (bool, runtime.Object, error) {
				obj := a.(ktesting.UpdateAction).GetObject().(*corev1.PersistentVolumeClaim)
				obj.ResourceVersion = strconv.FormatInt(rv.Add(1), 10)
				counts.pvcUpdates.Add(1)
				return false, nil, nil
			})
			// Emulate API-server resourceVersion changes on scheduler status patches.
			cs.PrependReactor("patch", "pods", func(a ktesting.Action) (bool, runtime.Object, error) {
				handled, obj, err := ktesting.ObjectReaction(cs.Tracker())(a)
				if err != nil || !handled {
					return handled, obj, err
				}
				pod := obj.(*corev1.Pod)
				pod.ResourceVersion = strconv.FormatInt(rv.Add(1), 10)
				err = cs.Tracker().Update(a.GetResource(), pod, a.GetNamespace())
				return true, pod, err
			})
			bound := make(chan string, 1)
			cs.PrependReactor("create", "pods", func(a ktesting.Action) (bool, runtime.Object, error) {
				if a.GetSubresource() != "binding" {
					return false, nil, nil
				}
				binding := a.(ktesting.CreateAction).GetObject().(*corev1.Binding)
				obj, err := cs.Tracker().Get(a.GetResource(), binding.Namespace, binding.Name)
				if err != nil {
					return true, nil, err
				}
				pod := obj.(*corev1.Pod)
				pod.Spec.NodeName = binding.Target.Name
				pod.ResourceVersion = strconv.FormatInt(rv.Add(1), 10)
				if err := cs.Tracker().Update(a.GetResource(), pod, pod.Namespace); err != nil {
					return true, nil, err
				}
				counts.binds.Add(1)
				counts.bindNanos.Store(time.Since(counts.started).Nanoseconds())
				bound <- binding.Target.Name
				return true, binding, nil
			})
			factory := informers.NewSharedInformerFactory(cs, 0)
			bindTimeout := int64(600)
			if tc.shortTimeout {
				bindTimeout = int64(seconds / 3)
			}
			plugins := &config.Plugins{
				QueueSort: config.PluginSet{Enabled: []config.Plugin{{Name: "PrioritySort"}}},
				PreFilter: config.PluginSet{Enabled: []config.Plugin{{Name: "ObservedVolumeBinding"}}},
				Filter:    config.PluginSet{Enabled: []config.Plugin{{Name: "ObservedVolumeBinding"}, {Name: Name}}},
				Reserve:   config.PluginSet{Enabled: []config.Plugin{{Name: "ObservedVolumeBinding"}, {Name: Name}}},
				PreBind:   config.PluginSet{Enabled: []config.Plugin{{Name: "ObservedVolumeBinding"}}},
				Bind:      config.PluginSet{Enabled: []config.Plugin{{Name: "DefaultBinder"}}},
			}
			registry := frameworkruntime.Registry{
				"ObservedVolumeBinding": func(_ runtime.Object, h framework.Handle) (framework.Plugin, error) {
					p, err := volumebinding.New(&config.VolumeBindingArgs{BindTimeoutSeconds: bindTimeout}, h, feature.Features{})
					if err != nil {
						return nil, err
					}
					return &observedVolumeBinding{VolumeBinding: p.(*volumebinding.VolumeBinding), counts: counts}, nil
				},
				Name: func(_ runtime.Object, h framework.Handle) (framework.Plugin, error) {
					s := &Scheduler{
						pvLister:      factory.Core().V1().PersistentVolumes().Lister(),
						pvcLister:     factory.Core().V1().PersistentVolumeClaims().Lister(),
						scLister:      factory.Storage().V1().StorageClasses().Lister(),
						diskScheduler: &testVolumeScheduler{driver: "disk.hwameistor.io"},
						lvmScheduler: &testVolumeScheduler{driver: "lvm.hwameistor.io", filter: func(volumes []string, claims []*corev1.PersistentVolumeClaim, node *corev1.Node) (bool, error) {
							if len(volumes) > 0 {
								pvc, err := factory.Core().V1().PersistentVolumeClaims().Lister().PersistentVolumeClaims("test").Get("claim")
								if err != nil {
									return false, err
								}
								if node.Name != pvc.Annotations[volumehelpers.AnnSelectedNode] {
									return false, fmt.Errorf("test: replica is on the original selected node")
								}
							}
							if len(claims) > 0 {
								counts.newChecks.Add(1)
								if tc.capacityConsumed && claims[0].Annotations[volumehelpers.AnnSelectedNode] != "" {
									return false, fmt.Errorf("test: capacity already consumed by the in-progress volume")
								}
							}
							return true, nil
						}},
					}
					return &observedHwameistor{Plugin: &Plugin{scheduler: s}, counts: counts}, nil
				},
			}
			sched, err := kubescheduler.New(cs, factory, nil, func(string) events.EventRecorder { return events.NewFakeRecorder(10000) }, ctx.Done(),
				kubescheduler.WithProfiles(config.KubeSchedulerProfile{SchedulerName: "test-scheduler", Plugins: plugins}),
				kubescheduler.WithFrameworkOutOfTreeRegistry(registry),
			)
			if err != nil {
				t.Fatal(err)
			}
			originalSchedule := sched.SchedulePod
			sched.SchedulePod = func(ctx context.Context, f framework.Framework, state *framework.CycleState, pod *corev1.Pod) (kubescheduler.ScheduleResult, error) {
				counts.cycles.Add(1)
				counts.snapshot("schedule")
				return originalSchedule(ctx, f, state, pod)
			}
			factory.Start(ctx.Done())
			for kind, ok := range factory.WaitForCacheSync(ctx.Done()) {
				if !ok {
					t.Fatalf("cache did not sync: %v", kind)
				}
			}
			runDone := make(chan struct{})
			go func() {
				sched.Run(ctx)
				close(runDone)
			}()
			defer func() {
				cancel()
				// Wake a blocked NextPod; Scheduler.Run closes its own queue.
				now := metav1.Now()
				_ = sched.SchedulingQueue.Add(&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{Name: "shutdown", Namespace: "test", UID: "shutdown", DeletionTimestamp: &now},
					Spec:       corev1.PodSpec{SchedulerName: "test-scheduler"},
				})
				<-runDone
			}()
			provisioning := make(chan error, 1)
			go func() {
				provisioning <- completeSlowPVC(ctx, cs, pvc, delay, counts)
			}()
			priority := int32(0)
			_, err = cs.CoreV1().Pods("test").Create(ctx, &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "pod", Namespace: "test", UID: types.UID(tc.name), ResourceVersion: "1"},
				Spec: corev1.PodSpec{SchedulerName: "test-scheduler", Priority: &priority,
					Volumes: []corev1.Volume{{Name: "data", VolumeSource: corev1.VolumeSource{PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: pvc.Name}}}},
				},
			}, metav1.CreateOptions{})
			if err != nil {
				t.Fatal(err)
			}
			select {
			case err := <-provisioning:
				if err != nil {
					t.Fatal(err)
				}
			case <-ctx.Done():
				t.Fatal("provisioning never completed", ctx.Err())
			}
			select {
			case node := <-bound:
				if tc.selected && node != "node-a" {
					t.Fatalf("bound on wrong node %s", node)
				}
			case <-ctx.Done():
				t.Fatal("pod never bound", ctx.Err())
			}
			counts.snapshot("result")
			if counts.binds.Load() != 1 {
				t.Fatalf("pod bound %d times", counts.binds.Load())
			}
			if counts.bindNanos.Load() < delay.Nanoseconds() {
				t.Fatal("pod was bound before provisioning completed")
			}
			if !tc.shortTimeout && (!tc.capacityConsumed || fixed) && counts.cycles.Load() != 1 {
				t.Fatalf("unexpected repeated scheduling: %d cycles", counts.cycles.Load())
			}
			if tc.selected && fixed && counts.newChecks.Load() != 0 {
				t.Fatal("selected PVC re-entered new capacity checking")
			}
			if tc.shortTimeout && counts.prebindErrors.Load() == 0 {
				t.Fatal("expected a bind timeout")
			}
			if !tc.shortTimeout && counts.prebindErrors.Load() != 0 {
				t.Fatal("unexpected PreBind error within the default timeout")
			}
			if tc.capacityConsumed && !fixed && counts.cycles.Load() <= 1 {
				t.Fatal("expected the original capacity check to reject the in-progress volume")
			}
		})
	}
}

func completeSlowPVC(ctx context.Context, cs *kubefake.Clientset, initial *corev1.PersistentVolumeClaim, delay time.Duration, counts *slowSchedulingCounts) error {
	if err := wait.PollImmediateUntil(20*time.Millisecond, func() (bool, error) {
		pvc, err := cs.CoreV1().PersistentVolumeClaims(initial.Namespace).Get(ctx, initial.Name, metav1.GetOptions{})
		return err == nil && pvc.Annotations[volumehelpers.AnnSelectedNode] != "", err
	}, ctx.Done()); err != nil {
		return err
	}
	counts.snapshot("provisioning-start")
	timer := time.NewTimer(delay)
	defer timer.Stop()
	// Progress updates exercise real informer events while PreBind is waiting.
	progress := time.NewTicker(delay / 18)
	defer progress.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-progress.C:
			pvc, err := cs.CoreV1().PersistentVolumeClaims(initial.Namespace).Get(ctx, initial.Name, metav1.GetOptions{})
			if err != nil {
				return err
			}
			pvc.Annotations["test.hwameistor.io/progress"] = time.Now().Format(time.RFC3339Nano)
			if _, err := cs.CoreV1().PersistentVolumeClaims(pvc.Namespace).Update(ctx, pvc, metav1.UpdateOptions{}); err != nil {
				return err
			}
			counts.snapshot("provisioning-progress")
		case <-timer.C:
			pvc, err := cs.CoreV1().PersistentVolumeClaims(initial.Namespace).Get(ctx, initial.Name, metav1.GetOptions{})
			if err != nil {
				return err
			}
			pv := &corev1.PersistentVolume{ObjectMeta: metav1.ObjectMeta{Name: "pv", ResourceVersion: "1"},
				Spec: corev1.PersistentVolumeSpec{
					StorageClassName: *pvc.Spec.StorageClassName,
					ClaimRef:         &corev1.ObjectReference{Namespace: pvc.Namespace, Name: pvc.Name, UID: pvc.UID},
					AccessModes:      pvc.Spec.AccessModes, Capacity: pvc.Spec.Resources.Requests,
					PersistentVolumeSource: corev1.PersistentVolumeSource{CSI: &corev1.CSIPersistentVolumeSource{Driver: "lvm.hwameistor.io", VolumeHandle: "lv"}},
				},
			}
			if _, err := cs.CoreV1().PersistentVolumes().Create(ctx, pv, metav1.CreateOptions{}); err != nil {
				return err
			}
			pvc.Spec.VolumeName = pv.Name
			pvc.Status.Phase = corev1.ClaimBound
			pvc.Annotations[volumehelpers.AnnBindCompleted] = "yes"
			_, err = cs.CoreV1().PersistentVolumeClaims(pvc.Namespace).Update(ctx, pvc, metav1.UpdateOptions{})
			counts.snapshot("provisioning-complete")
			return err
		}
	}
}
