// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package util

import (
	os "os"

	mock "github.com/stretchr/testify/mock"
)

// MockControllerInterface is an autogenerated mock type for the ControllerInterface type
type MockControllerInterface struct {
	mock.Mock
}

// CallGetFriendsListAPI provides a mock function with given fields: steamID, apiKey
func (_m *MockControllerInterface) CallGetFriendsListAPI(steamID string, apiKey string) (FriendsStruct, error) {
	ret := _m.Called(steamID, apiKey)

	var r0 FriendsStruct
	if rf, ok := ret.Get(0).(func(string, string) FriendsStruct); ok {
		r0 = rf(steamID, apiKey)
	} else {
		r0 = ret.Get(0).(FriendsStruct)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(steamID, apiKey)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CallIsAPIKeyValidAPI provides a mock function with given fields: apiKeys
func (_m *MockControllerInterface) CallIsAPIKeyValidAPI(apiKeys string) (string, error) {
	ret := _m.Called(apiKeys)

	var r0 string
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(apiKeys)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(apiKeys)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CallPlayerSummaryAPI provides a mock function with given fields: steamID, apiKey
func (_m *MockControllerInterface) CallPlayerSummaryAPI(steamID string, apiKey string) (UserStatsStruct, error) {
	ret := _m.Called(steamID, apiKey)

	var r0 UserStatsStruct
	if rf, ok := ret.Get(0).(func(string, string) UserStatsStruct); ok {
		r0 = rf(steamID, apiKey)
	} else {
		r0 = ret.Get(0).(UserStatsStruct)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(steamID, apiKey)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CreateFile provides a mock function with given fields: fileName
func (_m *MockControllerInterface) CreateFile(fileName string) (*os.File, error) {
	ret := _m.Called(fileName)

	var r0 *os.File
	if rf, ok := ret.Get(0).(func(string) *os.File); ok {
		r0 = rf(fileName)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*os.File)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(fileName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FileExists provides a mock function with given fields: steamID
func (_m *MockControllerInterface) FileExists(steamID string) bool {
	ret := _m.Called(steamID)

	var r0 bool
	if rf, ok := ret.Get(0).(func(string) bool); ok {
		r0 = rf(steamID)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Open provides a mock function with given fields: fileName
func (_m *MockControllerInterface) Open(fileName string) (*os.File, error) {
	ret := _m.Called(fileName)

	var r0 *os.File
	if rf, ok := ret.Get(0).(func(string) *os.File); ok {
		r0 = rf(fileName)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*os.File)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(fileName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// OpenFile provides a mock function with given fields: fileName, flag, perm
func (_m *MockControllerInterface) OpenFile(fileName string, flag int, perm os.FileMode) (*os.File, error) {
	ret := _m.Called(fileName, flag, perm)

	var r0 *os.File
	if rf, ok := ret.Get(0).(func(string, int, os.FileMode) *os.File); ok {
		r0 = rf(fileName, flag, perm)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*os.File)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, int, os.FileMode) error); ok {
		r1 = rf(fileName, flag, perm)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// WriteGzip provides a mock function with given fields: file, content
func (_m *MockControllerInterface) WriteGzip(file *os.File, content string) error {
	ret := _m.Called(file, content)

	var r0 error
	if rf, ok := ret.Get(0).(func(*os.File, string) error); ok {
		r0 = rf(file, content)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
