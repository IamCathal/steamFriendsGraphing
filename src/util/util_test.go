package util

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type testInput struct {
	steamID    string
	APIKey     string
	shouldFail bool
}

var (
	validUserSummaryResponse UserStatsStruct
)

func TestMain(m *testing.M) {
	os.Setenv("testing", "")
	path, err := os.Getwd()
	CheckErr(err)
	os.Setenv("BWD", fmt.Sprintf("%s/../", path))

	setupStubs()
	code := m.Run()

	os.Exit(code)
}

func setupStubs() {
	validUserSummaryResponse = UserStatsStruct{
		Response: Response{
			Players: []Player{
				{
					Steamid:     "76561198076045001",
					Timecreated: 0,
					Personaname: "expected persona name",
				},
			},
		},
	}
}

func TestGetPlayerSummary(t *testing.T) {
	mockController := &MockControllerInterface{}

	expectedSteamID := validUserSummaryResponse.Response.Players[0].Steamid
	mockController.On("CallPlayerSummaryAPI", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(validUserSummaryResponse, nil)

	userDetails, _ := GetPlayerSummary(mockController, expectedSteamID, "test API key")
	receivedSteamID := userDetails.Response.Players[0].Steamid

	assert.Equal(t, expectedSteamID, receivedSteamID, fmt.Sprintf("expected SteamID: %s but received SteamID: %s", expectedSteamID, receivedSteamID))
}

func TestCheckAPIKeys(t *testing.T) {
	mockController := &MockControllerInterface{}

	mockController.On("CallIsAPIKeyValidAPI", mock.AnythingOfType("string")).Return("valid response")
	mockController.On("IsValidResponseForAPIKey", mock.AnythingOfType("string")).Return(true)

	apiKeysToBeChecked := []string{
		"example API key",
		"another example API key",
		"bags of cans",
	}

	CheckAPIKeys(mockController, apiKeysToBeChecked)
}

func TestValidGetUserDetails(t *testing.T) {
	mockController := &MockControllerInterface{}
	steamID := "search steamID"

	expectedUser := validUserSummaryResponse.Response.Players[0]

	mockController.On("CallPlayerSummaryAPI", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(validUserSummaryResponse, nil)
	receivedUser, _ := GetUserDetails(mockController, "example API key", steamID)

	assert.NotNil(t, receivedUser, "expect to receive mocked user")
	assert.Equal(t, receivedUser["SteamID"], expectedUser.Steamid)
}

func TestGetUserDetailsForNonExistantUser(t *testing.T) {
	mockController := &MockControllerInterface{}
	steamID := "search steamID"
	expectedUserResponse := UserStatsStruct{
		Response: Response{
			Players: []Player{},
		},
	}

	mockController.On("CallPlayerSummaryAPI", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(expectedUserResponse, nil)
	_, err := GetUserDetails(mockController, "example API key", steamID)

	assert.NotNil(t, err, "expect error to be returned when receiving 0 users")
}

func TestGetUsernameValidFormatSteamID(t *testing.T) {
	mockController := &MockControllerInterface{}
	apiKeys := []string{"test API key"}
	steamID := "76561197960287930"

	expectedUsername := validUserSummaryResponse.Response.Players[0].Personaname
	mockController.On("CallPlayerSummaryAPI", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(validUserSummaryResponse, nil)

	receivedUsername, err := GetUsername(mockController, apiKeys[0], steamID)
	assert.Nil(t, err, fmt.Sprintf("can't get username for user: %s using key: %s", steamID, apiKeys[0]))
	assert.Equal(t, receivedUsername, expectedUsername, "expected to receive username: %s", expectedUsername)
}

func TestGetUsernameWithInvalidFormatSteamID(t *testing.T) {
	mockController := &MockControllerInterface{}
	apiKeys := []string{"test API key"}
	steamID := "invalid format SteamID"

	mockController.On("CallPlayerSummaryAPI", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(validUserSummaryResponse, nil)

	_, err := GetUsername(mockController, apiKeys[0], steamID)
	assert.NotNil(t, err, "didn't throw error for GetUsername call with invalid steamID: ", steamID)
}

func TestCreateDataFolder(t *testing.T) {
	err := CreateUserDataFolder()

	assert.Nil(t, err, "error creating user data folder")

	os.RemoveAll("../testData/")
}

func TestIsValidAPIResponseForSteamID(t *testing.T) {
	isValid := IsValidAPIResponseForSteamId("Internal Server Error")
	assert.False(t, isValid, "failed to catch invalid steamID")

	isValid = IsValidAPIResponseForSteamId("12 for €8.69 on Galahads is a mad deal")
	assert.True(t, isValid, "invalid steamID given for valid response")
}

func TestIsValidResponseForAPIKey(t *testing.T) {
	isValid := IsValidResponseForAPIKey("Forbidden")
	assert.False(t, isValid, "failed to catch invalid steamID")

	isValid = IsValidResponseForAPIKey("8 for €12 on Heineken is not a mad deal")
	assert.True(t, isValid, "invalid steamID given for valid response")
}

func TestExtractSteamIDs(t *testing.T) {
	steamIDs := []string{"76561198090461077", "76561198130544932"}
	IDs, err := ExtractSteamIDs(steamIDs)

	assert.Nil(t, err, "expect no errors when extracting valid IDs")
	assert.ElementsMatch(t, IDs, steamIDs, "expect to receive {\"76561198090461077\", \"76561198130544932\"}")
}

func TestExtarctSteamIDsWithNoneGiven(t *testing.T) {
	steamIDs2 := []string{}
	_, err := ExtractSteamIDs(steamIDs2)

	assert.NotNil(t, err, "no error given for empty steamID slice")
}

func TestIsValidFormatSteamIDWithValidSteamID(t *testing.T) {
	validSteamID := "76561198087169600"

	isValid := IsValidFormatSteamID(validSteamID)
	assert.True(t, isValid, fmt.Sprintf("expect to receive true for steamID: %s", validSteamID))
}

func TestIsValidFormatSteamIDWithInValidSteamID(t *testing.T) {
	invalidSteamID := "eeeeeeeee"

	isValid := IsValidFormatSteamID(invalidSteamID)
	assert.False(t, isValid, fmt.Sprintf("expect to receive false for steamID: %s", invalidSteamID))
}
