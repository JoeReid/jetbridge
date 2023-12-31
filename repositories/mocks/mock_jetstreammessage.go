// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/JoeReid/jetbridge/repositories (interfaces: JetstreamMessage)

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	jetbridge "github.com/JoeReid/jetbridge"
	gomock "github.com/golang/mock/gomock"
)

// MockJetstreamMessage is a mock of JetstreamMessage interface.
type MockJetstreamMessage struct {
	ctrl     *gomock.Controller
	recorder *MockJetstreamMessageMockRecorder
}

// MockJetstreamMessageMockRecorder is the mock recorder for MockJetstreamMessage.
type MockJetstreamMessageMockRecorder struct {
	mock *MockJetstreamMessage
}

// NewMockJetstreamMessage creates a new mock instance.
func NewMockJetstreamMessage(ctrl *gomock.Controller) *MockJetstreamMessage {
	mock := &MockJetstreamMessage{ctrl: ctrl}
	mock.recorder = &MockJetstreamMessageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockJetstreamMessage) EXPECT() *MockJetstreamMessageMockRecorder {
	return m.recorder
}

// Ack mocks base method.
func (m *MockJetstreamMessage) Ack() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Ack")
	ret0, _ := ret[0].(error)
	return ret0
}

// Ack indicates an expected call of Ack.
func (mr *MockJetstreamMessageMockRecorder) Ack() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Ack", reflect.TypeOf((*MockJetstreamMessage)(nil).Ack))
}

// Nak mocks base method.
func (m *MockJetstreamMessage) Nak() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Nak")
	ret0, _ := ret[0].(error)
	return ret0
}

// Nak indicates an expected call of Nak.
func (mr *MockJetstreamMessageMockRecorder) Nak() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Nak", reflect.TypeOf((*MockJetstreamMessage)(nil).Nak))
}

// Payload mocks base method.
func (m *MockJetstreamMessage) Payload() jetbridge.JetstreamLambdaPayload {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Payload")
	ret0, _ := ret[0].(jetbridge.JetstreamLambdaPayload)
	return ret0
}

// Payload indicates an expected call of Payload.
func (mr *MockJetstreamMessageMockRecorder) Payload() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Payload", reflect.TypeOf((*MockJetstreamMessage)(nil).Payload))
}
