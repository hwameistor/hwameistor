package thinpoolclaim

import (
	"context"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	v1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-storage/utils"
)

var (
	fakeThinPoolClaimName       = "thin-pool-claim-example"
	fakeThinPoolClaimUID        = "thin-pool-claim-example-uid"
	fakeNamespace               = "thin-pool-manager-test"
	fakeNodename                = "10-6-118-10"
	poolName                    = "pool1"
	apiversion                  = "hwameistor.io/v1alpha1"
	thinPoolClaimKind           = "ThinPoolClaim"
	cap100G               int64 = 100
	cap10G                int64 = 10
	overProvisionRatio          = "1.5"
	metadataSize                = uint(1)
	fakeRecorder                = record.NewFakeRecorder(100)
)

func TestReconcileThinPoolClaim_Reconcile(t *testing.T) {
	cli, s := CreateFakeClient()
	// Create a Reconcile for ThinPoolClaim
	r := ReconcileThinPoolClaim{
		Client:   cli,
		Scheme:   s,
		Recorder: fakeRecorder,
	}

	// Create LocalStorageNode with pool
	node := GenFakeLocalStorageNodeObject()
	err := r.Create(context.Background(), node)
	if err != nil {
		t.Errorf("Create LocalStorageNode fail %v", err)
	}
	defer r.DeleteFakeLocalStorageNode(t, node)

	// Create ThinPoolClaim
	claim := GenFakeThinPoolClaimObject(v1alpha1.ThinPoolClaimPhaseEmpty)
	err = r.Create(context.Background(), claim)
	if err != nil {
		t.Errorf("Create ThinPoolClaim fail %v", err)
	}
	defer r.DeleteFakeThinPoolClaim(t, claim)

	// Mock ThinPoolClaim request
	req := reconcile.Request{NamespacedName: types.NamespacedName{Namespace: claim.GetNamespace(), Name: claim.GetName()}}

	// First reconcile - should move from Empty to Pending
	_, err = r.Reconcile(context.TODO(), req)
	if err != nil {
		t.Errorf("Reconcile fail %v", err)
	}

	// Second reconcile - should move from Pending to ToBeConsumed (if enough capacity)
	_, err = r.Reconcile(context.TODO(), req)
	if err != nil {
		t.Errorf("Reconcile fail %v", err)
	}

	// Update claim
	err = r.Get(context.Background(), req.NamespacedName, claim)
	if err != nil {
		t.Errorf("Get thin pool claim fail %v", err)
	}

	// Check claim status, should be ToBeConsumed
	if claim.Status.Status != v1alpha1.ThinPoolClaimPhaseToBeConsumed {
		t.Errorf("Expected status %v but got %v", v1alpha1.ThinPoolClaimPhaseToBeConsumed, claim.Status.Status)
	}
}

func TestReconcileThinPoolClaim_Reconcile_NotEnoughCapacity(t *testing.T) {
	cli, s := CreateFakeClient()
	// Create a Reconcile for ThinPoolClaim
	r := ReconcileThinPoolClaim{
		Client:   cli,
		Scheme:   s,
		Recorder: fakeRecorder,
	}

	// Create LocalStorageNode with small pool capacity
	node := GenFakeLocalStorageNodeObject()
	node.Status.Pools[poolName] = v1alpha1.LocalPool{
		FreeCapacityBytes: 10 * utils.Gi, // Only 10G free
	}
	err := r.Create(context.Background(), node)
	if err != nil {
		t.Errorf("Create LocalStorageNode fail %v", err)
	}
	defer r.DeleteFakeLocalStorageNode(t, node)

	// Create ThinPoolClaim requesting 100G
	claim := GenFakeThinPoolClaimObject(v1alpha1.ThinPoolClaimPhaseEmpty)
	err = r.Create(context.Background(), claim)
	if err != nil {
		t.Errorf("Create ThinPoolClaim fail %v", err)
	}
	defer r.DeleteFakeThinPoolClaim(t, claim)

	// Mock ThinPoolClaim request
	req := reconcile.Request{NamespacedName: types.NamespacedName{Namespace: claim.GetNamespace(), Name: claim.GetName()}}

	// First reconcile - should move from Empty to Pending
	_, err = r.Reconcile(context.TODO(), req)
	if err != nil {
		t.Errorf("Reconcile fail %v", err)
	}

	// Second reconcile - should stay Pending due to insufficient capacity
	_, err = r.Reconcile(context.TODO(), req)
	if err == nil {
		t.Error("Expected error for insufficient capacity but got none")
	}

	// Update claim
	err = r.Get(context.Background(), req.NamespacedName, claim)
	if err != nil {
		t.Errorf("Get thin pool claim fail %v", err)
	}

	// Check claim status, should still be Pending
	if claim.Status.Status != v1alpha1.ThinPoolClaimPhasePending {
		t.Errorf("Expected status %v but got %v", v1alpha1.ThinPoolClaimPhasePending, claim.Status.Status)
	}
}

func TestReconcileThinPoolClaim_Reconcile_InvalidOverProvisionRatio(t *testing.T) {
	cli, s := CreateFakeClient()
	// Create a Reconcile for ThinPoolClaim
	r := ReconcileThinPoolClaim{
		Client:   cli,
		Scheme:   s,
		Recorder: fakeRecorder,
	}

	// Create LocalStorageNode
	node := GenFakeLocalStorageNodeObject()
	err := r.Create(context.Background(), node)
	if err != nil {
		t.Errorf("Create LocalStorageNode fail %v", err)
	}
	defer r.DeleteFakeLocalStorageNode(t, node)

	// Create ThinPoolClaim with invalid over provision ratio
	claim := GenFakeThinPoolClaimObject(v1alpha1.ThinPoolClaimPhaseEmpty)
	invalidRatio := "0.5" // Less than 1.0 is invalid
	claim.Spec.Description.OverProvisionRatio = &invalidRatio
	err = r.Create(context.Background(), claim)
	if err != nil {
		t.Errorf("Create ThinPoolClaim fail %v", err)
	}
	defer r.DeleteFakeThinPoolClaim(t, claim)

	// Mock ThinPoolClaim request
	req := reconcile.Request{NamespacedName: types.NamespacedName{Namespace: claim.GetNamespace(), Name: claim.GetName()}}

	// Reconcile - should fail due to invalid ratio
	_, err = r.Reconcile(context.TODO(), req)
	if err == nil {
		t.Error("Expected error for invalid over provision ratio but got none")
	}

	// Update claim
	err = r.Get(context.Background(), req.NamespacedName, claim)
	if err != nil {
		t.Errorf("Get thin pool claim fail %v", err)
	}

	// Check claim status, should still be Empty since validation failed
	if claim.Status.Status != v1alpha1.ThinPoolClaimPhaseEmpty {
		t.Errorf("Expected status %v but got %v", v1alpha1.ThinPoolClaimPhaseEmpty, claim.Status.Status)
	}
}

func TestReconcileThinPoolClaim_Reconcile_AllStatusTransitions(t *testing.T) {
	cli, s := CreateFakeClient()
	r := ReconcileThinPoolClaim{
		Client:   cli,
		Scheme:   s,
		Recorder: fakeRecorder,
	}

	testCases := []struct {
		description    string
		node           *v1alpha1.LocalStorageNode
		claim          *v1alpha1.ThinPoolClaim
		reconcileTimes int
		expectState    v1alpha1.ThinPoolClaimPhase
	}{
		{
			description:    "Status empty to pending",
			node:           GenFakeLocalStorageNodeObject(),
			claim:          GenFakeThinPoolClaimObject(v1alpha1.ThinPoolClaimPhaseEmpty),
			reconcileTimes: 1,
			expectState:    v1alpha1.ThinPoolClaimPhasePending,
		},
		{
			description:    "Status pending to toBeConsumed",
			node:           GenFakeLocalStorageNodeObject(),
			claim:          GenFakeThinPoolClaimObject(v1alpha1.ThinPoolClaimPhasePending),
			reconcileTimes: 1,
			expectState:    v1alpha1.ThinPoolClaimPhaseToBeConsumed,
		},
		{
			description:    "Status toBeConsumed to consumed",
			node:           GenFakeLocalStorageNodeObject(),
			claim:          GenFakeThinPoolClaimObject(v1alpha1.ThinPoolClaimPhaseToBeConsumed),
			reconcileTimes: 1,
			expectState:    v1alpha1.ThinPoolClaimPhaseToBeConsumed,
		},
		{
			description:    "Status consumed to toBeDeleted",
			node:           GenFakeLocalStorageNodeObject(),
			claim:          GenFakeThinPoolClaimObject(v1alpha1.ThinPoolClaimPhaseConsumed),
			reconcileTimes: 1,
			expectState:    v1alpha1.ThinPoolClaimPhaseToBeDeleted,
		},
		{
			description:    "Status toBeDeleted to deleted",
			node:           GenFakeLocalStorageNodeObject(),
			claim:          GenFakeThinPoolClaimObject(v1alpha1.ThinPoolClaimPhaseToBeDeleted),
			reconcileTimes: 1,
			expectState:    v1alpha1.ThinPoolClaimPhaseDeleted,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			err := r.Create(context.Background(), testCase.node)
			if err != nil {
				t.Errorf("Create LocalStorageNode fail %v", err)
			}
			defer r.DeleteFakeLocalStorageNode(t, testCase.node)

			err = r.Create(context.Background(), testCase.claim)
			if err != nil {
				t.Errorf("Create ThinPoolClaim fail %v", err)
			}
			if testCase.expectState != v1alpha1.ThinPoolClaimPhaseDeleted {
				defer r.DeleteFakeThinPoolClaim(t, testCase.claim)
			}

			req := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Namespace: testCase.claim.GetNamespace(),
					Name:      testCase.claim.GetName(),
				},
			}

			for i := 1; i <= testCase.reconcileTimes; i++ {
				_, err = r.Reconcile(context.TODO(), req)
				if err != nil {
					t.Errorf("Reconcile fail %v, times: %v", err, i)
				}
			}

			// refresh thinPoolClaim
			if err = r.Get(context.Background(), types.NamespacedName{
				Namespace: testCase.claim.GetNamespace(),
				Name:      testCase.claim.GetName(),
			}, testCase.claim); err != nil && testCase.expectState != v1alpha1.ThinPoolClaimPhaseDeleted {
				t.Errorf("Failed to refresh thinPoolClaim %s for err %v", req.NamespacedName, err)
			}

			if testCase.expectState != testCase.claim.Status.Status {
				t.Errorf("Expected ThinPoolClaim State %v but got State %v", testCase.expectState, testCase.claim.Status.Status)
			}
		})
	}
}

// DeleteFakeLocalStorageNode
func (r *ReconcileThinPoolClaim) DeleteFakeLocalStorageNode(t *testing.T, node *v1alpha1.LocalStorageNode) {
	if err := r.Delete(context.Background(), node); err != nil {
		t.Errorf("Delete LocalStorageNode %v fail %v", node.GetName(), err)
	}
}

// DeleteFakeThinPoolClaim
func (r *ReconcileThinPoolClaim) DeleteFakeThinPoolClaim(t *testing.T, tpc *v1alpha1.ThinPoolClaim) {
	if err := r.Delete(context.Background(), tpc); err != nil {
		t.Errorf("Delete ThinPoolClaim %v fail %v", tpc.GetName(), err)
	}
}

// GenFakeThinPoolClaimObject Create claim request
func GenFakeThinPoolClaimObject(status v1alpha1.ThinPoolClaimPhase) *v1alpha1.ThinPoolClaim {
	tpc := &v1alpha1.ThinPoolClaim{}

	TypeMeta := metav1.TypeMeta{
		Kind:       thinPoolClaimKind,
		APIVersion: apiversion,
	}

	ObjectMata := metav1.ObjectMeta{
		Name:              fakeThinPoolClaimName,
		Namespace:         "",
		ResourceVersion:   "",
		UID:               types.UID(fakeThinPoolClaimUID),
		CreationTimestamp: metav1.Time{Time: time.Now()},
	}

	Spec := v1alpha1.ThinPoolClaimSpec{
		NodeName: fakeNodename,
		Description: v1alpha1.ThinPoolClaimDescription{
			PoolName:           poolName,
			Capacity:           cap100G,
			OverProvisionRatio: &overProvisionRatio,
			PoolMetadataSize:   &metadataSize,
		},
	}

	tpc.ObjectMeta = ObjectMata
	tpc.TypeMeta = TypeMeta
	tpc.Spec = Spec
	tpc.Status.Status = status
	return tpc
}

// GenFakeLocalStorageNodeObject Create LocalStorageNode with pool
func GenFakeLocalStorageNodeObject() *v1alpha1.LocalStorageNode {
	lsn := &v1alpha1.LocalStorageNode{}

	TypeMeta := metav1.TypeMeta{
		APIVersion: apiversion,
		Kind:       "LocalStorageNode",
	}

	ObjectMata := metav1.ObjectMeta{
		Name:              fakeNodename,
		Namespace:         "",
		ResourceVersion:   "",
		UID:               types.UID("local-storage-node-uid"),
		CreationTimestamp: metav1.Time{Time: time.Now()},
	}

	// Create pool with 200G free capacity
	pools := make(map[string]v1alpha1.LocalPool)
	pools[poolName] = v1alpha1.LocalPool{
		Name:                     poolName,
		Class:                    "HDD",
		Type:                     "REGULAR",
		TotalCapacityBytes:       200 * utils.Gi,
		UsedCapacityBytes:        0,
		FreeCapacityBytes:        200 * utils.Gi,
		VolumeCapacityBytesLimit: 200 * utils.Gi,
		TotalVolumeCount:         0,
		UsedVolumeCount:          0,
		FreeVolumeCount:          0,
	}

	Status := v1alpha1.LocalStorageNodeStatus{
		State: v1alpha1.NodeStateReady,
		Pools: pools,
	}

	lsn.TypeMeta = TypeMeta
	lsn.ObjectMeta = ObjectMata
	lsn.Status = Status
	return lsn
}

// CreateFakeClient Create LocalStorageNode and ThinPoolClaim resource
func CreateFakeClient() (client.Client, *runtime.Scheme) {
	node := GenFakeLocalStorageNodeObject()
	nodeList := &v1alpha1.LocalStorageNodeList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: apiversion,
			Kind:       "LocalStorageNodeList",
		},
	}

	claim := GenFakeThinPoolClaimObject(v1alpha1.ThinPoolClaimPhaseEmpty)
	claimList := &v1alpha1.ThinPoolClaimList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: apiversion,
			Kind:       thinPoolClaimKind,
		},
	}

	s := scheme.Scheme
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, node)
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, nodeList)
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, claim)
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, claimList)
	return fake.NewClientBuilder().WithScheme(s).Build(), s
}
