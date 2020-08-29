package main

import (
	"fmt"
	"log"
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

func CreateUserDataFolder() error {
	// Create the userData folder to hold logs if it doesn't exist
	_, err := os.Stat("../userData/")
	if os.IsNotExist(err) {
		os.Mkdir("../userData/", 0755)
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
