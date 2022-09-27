package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

// Error checking method to reduce if statements for errors in the main function
func check(e error, message string) {
	if e != nil {
		fmt.Println(message)
		panic(e)
	}
}

// Helper method for pulling the KB search value from the query strings
// Modified the following algorithm from stack overflow:
// https://stackoverflow.com/questions/26916952/go-retrieve-a-string-from-between-two-characters-or-other-strings
// Returns an empty string if no value could be determined.
func getKBValueFromQueryString(queryString string) string {
	startIndex := strings.Index(queryString, "[")
	if startIndex == -1 {
			return ""
	}
	newString := queryString[startIndex+1:]
	endIndex := strings.Index(newString, "]")
	if endIndex == -1 {
			return ""
	}
	searchValueString := newString[:endIndex]

	return searchValueString
}

// Helper method for pulling the user and date values from the query string
// Modified the algorithm from the above method
// Returns empty string and the current system time if no values could be determined
func getUserAndDateValuesFromQueryString(queryString string) (string, time.Time) {
	userNameStartIndex := strings.Index(queryString, "r[")
	if userNameStartIndex == -1 {
			return "", time.Now()
	}
	updatedQueryString := queryString[userNameStartIndex+2:]
	userNameEndIndex := strings.Index(updatedQueryString, "]O")
	if userNameEndIndex == -1 {
			return "", time.Now()
	}
	userName := updatedQueryString[:userNameEndIndex]

	dateStartIndex := strings.Index(queryString, "e[")
	if dateStartIndex == -1 {
		return "", time.Now()
	}
	updatedDateQueryString := queryString[dateStartIndex+2:]
	dateEndIndex := strings.Index(updatedDateQueryString, "]")
	if dateEndIndex == -1 {
		return "", time.Now()
	}
	dateString := updatedDateQueryString[:dateEndIndex]

	fmt.Printf("Found the following values in the user action by date query string:\n user: %s, date: %s\n", userName, dateString)

	// Converting date string to go time package Time type
	const shortForm = "01 02 2006"
	date, err := time.Parse(shortForm, dateString)
	check(err, "Error parsing date from query string. Please check your input and try again")


	return userName, date
}


func main () {

	// Setup flags for command line parsing
	logFilePath := flag.String("logFilePath", "./server_log.csv", "the path of the server_log.csv file that you want metrics for")
	metricsQuery := flag.String("metricsQuery", "usersAccessed", "the query string used to pull metrics from the server_log.csv file. Must be the following formats: usersAccessed uploadsGreaterThan[value] uploadsLessThan[value] downloadsGreaterThan[value] downloadsLessThan[value]")

	if !strings.Contains(*metricsQuery, "usersAccessed") &&
		 (!strings.Contains(*metricsQuery, "uploadsGreaterThan") ||
		 !strings.Contains(*metricsQuery, "uploadsLessThan") ||
		 !strings.Contains(*metricsQuery, "downloadsGreaterThan") ||
		 !strings.Contains(*metricsQuery, "downloadsLessThan")) &&
		 (!strings.Contains(*metricsQuery, "uploadsByUser") ||
		 !strings.Contains(*metricsQuery, "downloadsByUser")) {

		fmt.Printf("metricsQuery argument %s is not an accepted query format. Please view the utilities usage and try again.\n", *metricsQuery)
		os.Exit(1)
	}

	// Actually pull in the values from the command line to the variables defined above
	flag.Parse()

	// Attempt to open the file path passed to the utility. If it errors out then send that message to the command prompt and panic exit.
	logFileOpened, err := os.Open(*logFilePath)
	check(err, "CSV file could not be opened. Please verify the file path you've provided and try again.")

	// Defer the closing of the file to when the main method finishes processing
	defer logFileOpened.Close()

	// Read in the values of the csv file using csv.Reader
	csvLogFileReader := csv.NewReader(logFileOpened)

	// Initialize line counter for better logging to user
	lineNumber := 1

	// Create data structures for unique users query
	uniqueUsers := map[string]int{}

	// Create boolean that will determine whether or not we want the amount of file sizes above or below our search value
	wantLess := true

	// Create boolean that determines the type of file action we want to find the aggregate for
	wantUploads := true

	// Creates boolean for choosing user file actions per date
	wantUserActionsOnDate := false

	// Create counters for the amounts of uploads/downloads above/below a certain size
	uploads := 0
	downloads := 0
	userUploadsOnDate := 0
	userDownloadsOnDate := 0

	// Create data storage for the username and the date we're doing a lookup for
	userNameSearchValue := ""
	fileActionDateSearchValue := time.Now()

	// Create data storage for the KB search value
	kBSearchValue := 0
	kBSearchValueString := ""

	// Get search value from the string. Only works if we are given a correctly formatted metricsQuery.
	// Needs better error handling.
	// Selects the values based on if we're choosing UserByDate or by kB size.
	if strings.Contains(*metricsQuery, "OnDate") {
		wantUserActionsOnDate = true
		userNameSearchValue, fileActionDateSearchValue = getUserAndDateValuesFromQueryString(*metricsQuery)
	} else if strings.Contains(*metricsQuery, "GreaterThan") || strings.Contains(*metricsQuery, "LessThan") {
		kBSearchValueString = getKBValueFromQueryString(*metricsQuery)
		// Convert search value to integer if we are doing an uploads/downloads query
		kBSearchValue, err = strconv.Atoi(kBSearchValueString)
		// Check for error in value conversion and inform user of the occurrence if there is one
		check(err, "Error in retrieving query search value for the > and < query. Please verify your input and try again.")
	}

	// Decision tree for setting boolean values before processing
	if strings.Contains(*metricsQuery, "downloads") {
		wantUploads = false
	}
	if strings.Contains(*metricsQuery, "GreaterThan") {
		wantLess = false
	}

	// Give user information about utilities status
	fmt.Println("File has been opened and processing is beginning...")

	for {
		// Receive a line from the file and check to see if an error occurs during each line read.
		// If the err is io.EOF then we jump out of our infinite loop.
		// If there was an error that wasn't io.EOF then we print a message explaining file
		// reading failed at line #x while in process and panic exit.
		csvRecord, err := csvLogFileReader.Read()

		// Skip the first line of the file as it just contains the column names
		// Increment the lineNumber before skipping to ensure we don't end up in an infinite loop and can continue processing
		if lineNumber == 1 {
			lineNumber += 1
			continue
		}

		// Infinite loop termination condition
		if err == io.EOF {
			break
		}
		// Error message for any other error besides io.EOF
		check(err, fmt.Sprintf("There was an error reading the file during processing at line #%d. File could not be processed. Please check the contents for errors in formatting and try again", lineNumber))

		// Use map as a makeshift set type by only inputting unique keys. We'll use the length of the map to determine the number of users that have accessed the server.
		// O(# unique users) memory. O(1) lookup for the answer to the query
		currentUserRecord := csvRecord[1]
		_, keyPresent := uniqueUsers[currentUserRecord]
		if !keyPresent {
			uniqueUsers[currentUserRecord] = 1
		}

		// Every query will take O(N) lookup time. We will read the file a single time and answer the query
		// that was passed in at runtime. Providing the result based on what was requested.
		// We'll just make the counts of the values we're looking for and return the answer at the end.

		// Temp variables for easy reading later on
		// Date string being parsed into go time.Time type
		const unixDateForm = "Mon Jan 02 15:04:05 UTC 2006"
		currentActionDateString := csvRecord[0]
		currentActionDate, err := time.Parse(unixDateForm, currentActionDateString)
		check(err, fmt.Sprintf("Error converting date value on line #%d from csv to Go's time.Time format. Please check your input and try again.", lineNumber))
		// File action
		currentFileAction := csvRecord[2]
		// CSV package reads the file sizes as strings so converting them to integers
		currentFileSize, err := strconv.Atoi(csvRecord[3])
		// Error handling the string to int conversion and printing the line number
		check(err, fmt.Sprintf("Failed converting the file size string to int when processing the record on line #%d", lineNumber))

		// Update the aggregates correctly based on the decision booleans
		if wantUserActionsOnDate && currentUserRecord == userNameSearchValue && wantUploads && currentFileAction == "upload" &&
			currentActionDate.Year() == fileActionDateSearchValue.Year() &&
			currentActionDate.Month() == fileActionDateSearchValue.Month() &&
			currentActionDate.Day() == fileActionDateSearchValue.Day() {

			userUploadsOnDate += 1
		} else if wantUserActionsOnDate && currentUserRecord == userNameSearchValue && !wantUploads && currentFileAction == "download" &&
			currentActionDate.Year() == fileActionDateSearchValue.Year() &&
			currentActionDate.Month() == fileActionDateSearchValue.Month() &&
			currentActionDate.Day() == fileActionDateSearchValue.Day() {

			userDownloadsOnDate += 1
		} else if wantUploads && wantLess && currentFileAction == "upload" && currentFileSize < kBSearchValue {
			uploads += 1
		} else if wantUploads && !wantLess && currentFileAction == "upload" && currentFileSize > kBSearchValue {
			uploads += 1
		} else if !wantUploads && wantLess && currentFileAction == "download" && currentFileSize < kBSearchValue {
			downloads += 1
		} else if !wantUploads && !wantLess && currentFileAction == "download" && currentFileSize > kBSearchValue {
			downloads += 1
		}

		// Increment lineNumber for logging
		lineNumber += 1

	}

	// Inform the user that we've read and processed the file and tell them how many records there were. We subtract two because the
	// first line of the file is just the column names and we are over incrementing the lineNumber variable by one simply because of its
	// increment placement in the processing loop.
	fmt.Printf("Finished reading and processing file. There were %d records total. The result of your query will be provided below...\n\n", lineNumber-2)

	// Decision tree for informing user of the results of their query
	if strings.Contains(*metricsQuery, "usersAccessed") {
		fmt.Printf("The number of unique users that have accessed the system is: %d\n", len(uniqueUsers))
	} else if wantUserActionsOnDate && wantUploads {
		fmt.Printf("The number of uploads from user (%s) on date (%s) is: %d\n", userNameSearchValue, fileActionDateSearchValue.Format("01-02-2006"), userUploadsOnDate)
	} else if wantUserActionsOnDate && !wantUploads {
		fmt.Printf("The number of downloads from user (%s) on date (%s) is: %d\n", userNameSearchValue, fileActionDateSearchValue.Format("01-02-2006"), userDownloadsOnDate)
	} else if wantUploads && wantLess {
		fmt.Printf("The number of uploads less than %dkB is: %d\n", kBSearchValue, uploads)
	} else if wantUploads && !wantLess {
		fmt.Printf("The number of uploads greater than %dkB is: %d\n", kBSearchValue, uploads)
	} else if !wantUploads && wantLess {
		fmt.Printf("The number of downloads less than %dkB is: %d\n", kBSearchValue, downloads)
	} else if !wantUploads && !wantLess {
		fmt.Printf("The number of downloads greater than %dkB is: %d\n", kBSearchValue, downloads)
	}

}
