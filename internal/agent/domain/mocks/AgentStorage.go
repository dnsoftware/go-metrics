// Code generated by mockery v2.20.2. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// AgentStorage is an autogenerated mock type for the AgentStorage type
type AgentStorage struct {
	mock.Mock
}

// GetCounter provides a mock function with given fields: name
func (_m *AgentStorage) GetCounter(name string) (int64, error) {
	ret := _m.Called(name)

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (int64, error)); ok {
		return rf(name)
	}
	if rf, ok := ret.Get(0).(func(string) int64); ok {
		r0 = rf(name)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetGauge provides a mock function with given fields: name
func (_m *AgentStorage) GetGauge(name string) (float64, error) {
	ret := _m.Called(name)

	var r0 float64
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (float64, error)); ok {
		return rf(name)
	}
	if rf, ok := ret.Get(0).(func(string) float64); ok {
		r0 = rf(name)
	} else {
		r0 = ret.Get(0).(float64)
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SetCounter provides a mock function with given fields: name, value
func (_m *AgentStorage) SetCounter(name string, value int64) {
	_m.Called(name, value)
}

// SetGauge provides a mock function with given fields: name, value
func (_m *AgentStorage) SetGauge(name string, value float64) {
	_m.Called(name, value)
}

type mockConstructorTestingTNewAgentStorage interface {
	mock.TestingT
	Cleanup(func())
}

// NewAgentStorage creates a new instance of AgentStorage. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewAgentStorage(t mockConstructorTestingTNewAgentStorage) *AgentStorage {
	mock := &AgentStorage{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}