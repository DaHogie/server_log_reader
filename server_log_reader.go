package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// Error checking method to reduce if statements for errors in the main function
func check(e error, message string) {
	if e != nil {
		fmt.Println(message)
		panic(e)
	}
}

// Helper method for pulling the search value from the query strings
// Modified the following algorithm from stack overflow:
// https://stackoverflow.com/questions/26916952/go-retrieve-a-string-from-between-two-characters-or-other-strings
// Returns an empty string if no value could be determined.
func getSearchValueFromQueryString(queryString string) string {
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

func main () {

	// Setup flags for command line parsing
	logFilePath := flag.String("logFilePath", "./server_log.csv", "the path of the server_log.csv file that you want metrics for")
	metricsQuery := flag.String("metricsQuery", "usersAccessed", "the query string used to pull metrics from the server_log.csv file. Must be the following formats: usersAccessed uploadsGreaterThan[value] uploadsLessThan[value] downloadsGreaterThan[value] downloadsLessThan[value]")

	if !strings.Contains(*metricsQuery, "usersAccessed") &&
		 (!strings.Contains(*metricsQuery, "uploadsGreaterThan[") ||
		 !strings.Contains(*metricsQuery, "uploadsLessThan") ||
		 !strings.Contains(*metricsQuery, "downloadsGreaterThan[") ||
		 !strings.Contains(*metricsQuery, "downloadsLessThan[")) {

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

	// Create counters for the amounts of uploads/downloads above/below a certain size
	uploads := 0
	downloads := 0

	// Get search value from the string. Only works if we are given a correctly formatted metricsQuery.
	// Needs better error handling.
	searchValueString := getSearchValueFromQueryString(*metricsQuery)

	// Convert search value to integer if we are doing an uploads/downloads query
	searchValue, err := strconv.Atoi(searchValueString)
	// Check for error in value conversion and inform user of the occurrence if there is one
	check(err, "Error in retrieving query search value for the > and < query. Please verify your input and try again.")

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
		currentFileAction := csvRecord[2]
		// CSV package reads the file sizes as strings so converting them to integers
		currentFileSize, err := strconv.Atoi(csvRecord[3])
		// Error handling the string to int conversion and printing the line number
		check(err, fmt.Sprintf("Failed converting the file size string to int when processing the record on line #%d", lineNumber))

		// Update the aggregates correctly based on the decision booleans
		if wantUploads && wantLess && currentFileAction == "upload" && currentFileSize < searchValue {
			uploads += 1
		} else if wantUploads && !wantLess && currentFileAction == "upload" && currentFileSize > searchValue {
			uploads += 1
		} else if !wantUploads && wantLess && currentFileAction == "download" && currentFileSize < searchValue {
			downloads += 1
		} else if !wantUploads && !wantLess && currentFileAction == "download" && currentFileSize > searchValue {
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
	} else if wantUploads && wantLess {
		fmt.Printf("The number of uploads less than %dkB is: %d\n", searchValue, uploads)
	} else if wantUploads && !wantLess {
		fmt.Printf("The number of uploads greater than %dkB is: %d\n", searchValue, uploads)
	} else if !wantUploads && wantLess {
		fmt.Printf("The number of downloads less than %dkB is: %d\n", searchValue, downloads)
	} else if !wantUploads && !wantLess {
		fmt.Printf("The number of downloads greater than %dkB is: %d\n", searchValue, downloads)
	}

}
