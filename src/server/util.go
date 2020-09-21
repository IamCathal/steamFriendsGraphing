package main

// // Log an API call to the console with it's details
// func LogCall(method, endpoint, status, startTimeString string, cached bool) {
// 	statusColor := "\033[0m"
// 	cacheString := ""

// 	if cached {
// 		cacheString = "[CACHE] "
// 	}

// 	startTime, err := strconv.ParseInt(startTimeString, 10, 64)
// 	if err != nil {
// 		startTime = -1
// 	}
// 	endTime := time.Now().UnixNano() / int64(time.Millisecond)
// 	delay := endTime - startTime

// 	// If the HTTP status given is 2XX, give it a nice
// 	// green color, otherwise give it a red color
// 	if status[0] == '2' {
// 		statusColor = "\033[32m"
// 	} else {
// 		statusColor = "\033[31m"
// 	}
// 	fmt.Printf("[%s] %s%s %s %s%s%s %dms\n", time.Now().Format("02-Jan-2006 15:04:05"), cacheString, method, endpoint, statusColor, status, "\033[0m", delay)
// }
