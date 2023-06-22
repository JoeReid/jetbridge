// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/JoeReid/jetbridge/repositories (interfaces: MessageSource)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	repositories "github.com/JoeReid/jetbridge/repositories"
	gomock "github.com/golang/mock/gomock"
)

// MockMessageSource is a mock of MessageSource interface.
type MockMessageSource struct {
	ctrl     *gomock.Controller
	recorder *MockMessageSourceMockRecorder
}

// MockMessageSourceMockRecorder is the mock recorder for MockMessageSource.
type MockMessageSourceMockRecorder struct {
	mock *MockMessageSource
}

// NewMockMessageSource creates a new mock instance.
func NewMockMessageSource(ctrl *gomock.Controller) *MockMessageSource {
	mock := &MockMessageSource{ctrl: ctrl}
	mock.recorder = &MockMessageSourceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMessageSource) EXPECT() *MockMessageSourceMockRecorder {
	return m.recorder
}

// FetchJetstreamMessages mocks base method.
func (m *MockMessageSource) FetchJetstreamMessages(arg0 context.Context, arg1 repositories.JetstreamBinding) ([]repositories.JetstreamMessage, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FetchJetstreamMessages", arg0, arg1)
	ret0, _ := ret[0].([]repositories.JetstreamMessage)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FetchJetstreamMessages indicates an expected call of FetchJetstreamMessages.
func (mr *MockMessageSourceMockRecorder) FetchJetstreamMessages(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FetchJetstreamMessages", reflect.TypeOf((*MockMessageSource)(nil).FetchJetstreamMessages), arg0, arg1)
}
