// Code generated by MockGen. DO NOT EDIT.
// Source: internal/api/orders.go
//
// Generated by this command:
//
//	mockgen -source=internal/api/orders.go -destination=internal/mocks/order_storage_mock.go -package=mocks
//

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	models "github.com/evgfitil/gophermart.git/internal/models"
	gomock "go.uber.org/mock/gomock"
)

// MockOrderStorage is a mock of OrderStorage interface.
type MockOrderStorage struct {
	ctrl     *gomock.Controller
	recorder *MockOrderStorageMockRecorder
}

// MockOrderStorageMockRecorder is the mock recorder for MockOrderStorage.
type MockOrderStorageMockRecorder struct {
	mock *MockOrderStorage
}

// NewMockOrderStorage creates a new mock instance.
func NewMockOrderStorage(ctrl *gomock.Controller) *MockOrderStorage {
	mock := &MockOrderStorage{ctrl: ctrl}
	mock.recorder = &MockOrderStorageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockOrderStorage) EXPECT() *MockOrderStorageMockRecorder {
	return m.recorder
}

// GetOrders mocks base method.
func (m *MockOrderStorage) GetOrders(ctx context.Context, userID int) ([]models.Order, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetOrders", ctx, userID)
	ret0, _ := ret[0].([]models.Order)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetOrders indicates an expected call of GetOrders.
func (mr *MockOrderStorageMockRecorder) GetOrders(ctx, userID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOrders", reflect.TypeOf((*MockOrderStorage)(nil).GetOrders), ctx, userID)
}

// ProcessOrder mocks base method.
func (m *MockOrderStorage) ProcessOrder(ctx context.Context, order models.Order) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ProcessOrder", ctx, order)
	ret0, _ := ret[0].(error)
	return ret0
}

// ProcessOrder indicates an expected call of ProcessOrder.
func (mr *MockOrderStorageMockRecorder) ProcessOrder(ctx, order any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProcessOrder", reflect.TypeOf((*MockOrderStorage)(nil).ProcessOrder), ctx, order)
}
