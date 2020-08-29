package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"
)

var (
	green = "\033[32m"
	red   = "\033[31m"
)

// Divmod divides a friendslist into stacks of 100 and the remainder
func Divmod(numerator, denominator int) (quotient, remainder int) {
	quotient = numerator / denominator
	remainder = numerator % denominator
	return
}

func GetUsername(apiKey, steamID string) (string, error) {

	if valid := IsValidFormatSteamID(steamID); !valid {
		return "", fmt.Errorf("invalid steamID format")
	}

	// Get the target username from the ID
	targetURL := fmt.Sprintf("http://api.steampowered.com/ISteamUser/GetPlayerSummaries/v0002/?key=%s&steamids=%s",
		apiKey, steamID)
	res, err := http.Get(targetURL)
	if err != nil {
		return "", err
	}
	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		return "", err
	}

	var userStatsObj UserStatsStruct
	json.Unmarshal(body, &userStatsObj)

	if len(userStatsObj.Response.Players) == 0 {
		return "", fmt.Errorf("invalid steamID %s given", steamID)
	}

	return userStatsObj.Response.Players[0].Personaname, nil
}

func PrintUserDetails(apiKey, steamID string) error {

	if valid := IsValidFormatSteamID(steamID); !valid {
	}

	// Get the target username from the ID
	targetURL := fmt.Sprintf("http://api.steampowered.com/ISteamUser/GetPlayerSummaries/v0002/?key=%s&steamids=%s",
		apiKey, steamID)
	res, err := http.Get(targetURL)
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		return err
	}

	var userStatsObj UserStatsStruct
	json.Unmarshal(body, &userStatsObj)

	if len(userStatsObj.Response.Players) == 0 {
		return fmt.Errorf("invalid steamID %s given", steamID)
	}

	fmt.Printf("SteamID: %s\n", userStatsObj.Response.Players[0].Steamid)
	fmt.Printf("Username: %s\n", userStatsObj.Response.Players[0].Personaname)
	fmt.Printf("Profile URL: %s\n", userStatsObj.Response.Players[0].Profileurl)
	fmt.Printf("Time Created: %s\n", time.Unix(int64(userStatsObj.Response.Players[0].Timecreated), 0))
	return nil
}

func CreateUserDataFolder() error {
	// Create the userData folder to hold logs if it doesn't exist
	_, err := os.Stat("../userData/")
	if os.IsNotExist(err) {
		os.Mkdir("../userData/", 0755)
		return nil
	}

	if err != nil {
		return err
	}
	return nil
}

// LogCall logs a to the API with various stats on the request
func LogCall(method, steamID, username, status, statusColor string, startTime int64) {
	endTime := time.Now().UnixNano() / int64(time.Millisecond)
	delay := strconv.FormatInt((endTime - startTime), 10)

	fmt.Printf("%s [%s] %s %s%s%s %vms\n", method, steamID, username,
		statusColor, status, "\033[0m", delay)
}

// CheckErr is a simple function to replace dozen or so if err != nil statements
func CheckErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func IsValidFormatSteamID(steamID string) bool {
	match, _ := regexp.MatchString("([0-9]){17}", steamID)
	if !match {
		return false
	}
	return true
}

func IsValidSteamID(body string) bool {
	match, _ := regexp.MatchString("(Internal Server Error)+", body)
	if match {
		return false
	}
	return true
}

func IsValidAPIKey(body string) bool {
	match, _ := regexp.MatchString("(Forbidden)+", body)
	if match {
		return false
	}
	return true
}
