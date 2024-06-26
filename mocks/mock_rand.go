// Code generated by MockGen. DO NOT EDIT.
// Source: math/rand (interfaces: Source)

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockSource is a mock of Source interface.
type MockSource struct {
	ctrl     *gomock.Controller
	recorder *MockSourceMockRecorder
}

// MockSourceMockRecorder is the mock recorder for MockSource.
type MockSourceMockRecorder struct {
	mock *MockSource
}

// NewMockSource creates a new mock instance.
func NewMockSource(ctrl *gomock.Controller) *MockSource {
	mock := &MockSource{ctrl: ctrl}
	mock.recorder = &MockSourceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSource) EXPECT() *MockSourceMockRecorder {
	return m.recorder
}

// Int63 mocks base method.
func (m *MockSource) Int63() int64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Int63")
	ret0, _ := ret[0].(int64)
	return ret0
}

// Int63 indicates an expected call of Int63.
func (mr *MockSourceMockRecorder) Int63() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Int63", reflect.TypeOf((*MockSource)(nil).Int63))
}

// Seed mocks base method.
func (m *MockSource) Seed(arg0 int64) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Seed", arg0)
}

// Seed indicates an expected call of Seed.
func (mr *MockSourceMockRecorder) Seed(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Seed", reflect.TypeOf((*MockSource)(nil).Seed), arg0)
}
