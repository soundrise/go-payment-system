// Code generated by MockGen. DO NOT EDIT.
// Source: controller.go

// Package payment is a generated GoMock package.
package payment

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	payment "github.com/soundrise/go-payment-system/payment"
)

// MockController is a mock of Controller interface.
type MockController struct {
	ctrl     *gomock.Controller
	recorder *MockControllerMockRecorder
}

// MockControllerMockRecorder is the mock recorder for MockController.
type MockControllerMockRecorder struct {
	mock *MockController
}

// NewMockController creates a new mock instance.
func NewMockController(ctrl *gomock.Controller) *MockController {
	mock := &MockController{ctrl: ctrl}
	mock.recorder = &MockControllerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockController) EXPECT() *MockControllerMockRecorder {
	return m.recorder
}

// Add mocks base method.
func (m *MockController) Add(arg0 payment.HandlerFunc) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Add", arg0)
}

// Add indicates an expected call of Add.
func (mr *MockControllerMockRecorder) Add(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Add", reflect.TypeOf((*MockController)(nil).Add), arg0)
}

// Run mocks base method.
func (m *MockController) Run() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Run")
}

// Run indicates an expected call of Run.
func (mr *MockControllerMockRecorder) Run() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Run", reflect.TypeOf((*MockController)(nil).Run))
}