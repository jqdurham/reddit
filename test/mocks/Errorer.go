// Code generated by mockery v2.42.2. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// Errorer is an autogenerated mock type for the Errorer type
type Errorer struct {
	mock.Mock
}

// Err provides a mock function with given fields:
func (_m *Errorer) Err() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Err")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Err2 provides a mock function with given fields:
func (_m *Errorer) Err2() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Err2")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewErrorer creates a new instance of Errorer. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewErrorer(t interface {
	mock.TestingT
	Cleanup(func())
}) *Errorer {
	mock := &Errorer{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}