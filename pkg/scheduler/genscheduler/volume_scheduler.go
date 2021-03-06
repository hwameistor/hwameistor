// Code generated by MockGen. DO NOT EDIT.
// Source: types.go

// Package genscheduler is a generated GoMock package.
package genscheduler

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	v1 "k8s.io/api/core/v1"
)

// MockVolumeScheduler is a mock of VolumeScheduler interface.
type MockVolumeScheduler struct {
	ctrl     *gomock.Controller
	recorder *MockVolumeSchedulerMockRecorder
}

// MockVolumeSchedulerMockRecorder is the mock recorder for MockVolumeScheduler.
type MockVolumeSchedulerMockRecorder struct {
	mock *MockVolumeScheduler
}

// NewMockVolumeScheduler creates a new mock instance.
func NewMockVolumeScheduler(ctrl *gomock.Controller) *MockVolumeScheduler {
	mock := &MockVolumeScheduler{ctrl: ctrl}
	mock.recorder = &MockVolumeSchedulerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockVolumeScheduler) EXPECT() *MockVolumeSchedulerMockRecorder {
	return m.recorder
}

// CSIDriverName mocks base method.
func (m *MockVolumeScheduler) CSIDriverName() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CSIDriverName")
	ret0, _ := ret[0].(string)
	return ret0
}

// CSIDriverName indicates an expected call of CSIDriverName.
func (mr *MockVolumeSchedulerMockRecorder) CSIDriverName() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CSIDriverName", reflect.TypeOf((*MockVolumeScheduler)(nil).CSIDriverName))
}

// Filter mocks base method.
func (m *MockVolumeScheduler) Filter(existingLocalVolume []string, unboundPVCs []*v1.PersistentVolumeClaim, node *v1.Node) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Filter", existingLocalVolume, unboundPVCs, node)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Filter indicates an expected call of Filter.
func (mr *MockVolumeSchedulerMockRecorder) Filter(existingLocalVolume, unboundPVCs, node interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Filter", reflect.TypeOf((*MockVolumeScheduler)(nil).Filter), existingLocalVolume, unboundPVCs, node)
}
