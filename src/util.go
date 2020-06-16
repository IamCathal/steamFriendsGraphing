package main

import (
	"fmt"
	"log"
	"strconv"
)

// Divmod divices a friendslist into stacks of 100 and the remainder
func Divmod(numerator, denominator int) (quotient, remainder int) {
	quotient = numerator / denominator
	remainder = numerator % denominator
	return
}

// LogCall logs a HTTP call to console with HTTP status and round-trip time
func LogCall(method, steamID, username, status, statusColor string, roundTripTime int64) {
	delay := strconv.FormatInt(roundTripTime, 10)
	fmt.Printf("%s [%s] %s %s%s%s %vms\n", method, steamID, username, statusColor, status, "\033[0m", delay)
}

// CheckErr is a simple function to replace dozen or so if err != nil statements
func CheckErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
