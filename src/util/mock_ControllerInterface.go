// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package util

import mock "github.com/stretchr/testify/mock"

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
func (_m *MockControllerInterface) CallIsAPIKeyValidAPI(apiKeys string) string {
	ret := _m.Called(apiKeys)

	var r0 string
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(apiKeys)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
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
