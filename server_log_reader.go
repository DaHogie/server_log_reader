package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
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

// Helper method for pulling the search value from the query string
// Returns -1 if it was unable to pull the value correctly
func getSearchValueFromQueryString(queryString string) int {
	startIndex := strings.Index(queryString, "[")
	if startIndex == -1 {
			return -1
	}
	newString := queryString[startIndex+1:]
	endIndex := strings.Index(newString, "]")
	if endIndex == -1 {
			return -1
	}
	searchValueString := newString[:endIndex]

	// Convert the searchValue to an integer
	searchValue, err := strconv.Atoi(searchValueString)
	// Error handling the string to int conversion, and print a valuable message for the user
	check(err, "Failed to find the search value in the queryString. Make sure that the value is an integer and try again.")

	return searchValue
}


// Helper method for determining where the workingIndex for our math on the > and < queries is.
// Takes in the value of the fileSize list, the queryString for determining the search value, and the boolean that determines if we're heading up or down the list
// to find the amount of values we're looking for.
func findWorkingIndex(intList []int, searchValue int, wantLess bool) int {
	// Create return value holder
	workingIndex := -1

	// Find the working index and check for duplicates if/when it's found. If the workingIndex is -1
	// at the end then we should have nothing in the list.
	for index, value := range intList {
		if value >= searchValue {
			workingIndex = index
			if !wantLess {
				for {
					if workingIndex < len(intList) && workingIndex+1 < len(intList) && intList[workingIndex+1] == searchValue {
						workingIndex += 1
					} else {
						break
					}
				}
			} else {
				break
			}
		}
	}

	return workingIndex
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
	lineNumber := 0

	// Create data structures for in memory storage for handling the queries
	uniqueUsers := map[string]int{}
	uploadSizes := []int{}
	downloadSizes := []int{}

	// Give user information about utilities status
	fmt.Println("File has been opened and processing is beginning...")

	for {
		// Receive a line from the file and check to see if an error occurs during each line read.
		// If the err is io.EOF then we jump out of our infinite loop.
		// If there was an error that wasn't io.EOF then we print a message explaining file
		// reading failed at line #x while in process and panic exit.
		csvRecord, err := csvLogFileReader.Read()

		// Skip the first line of the file as it just contains the column names
		// Increment the lineNumber otherwise for logging
		if lineNumber == 0 {
			lineNumber += 1
			continue
		}
		if err == io.EOF {
			break
		}
		check(err, fmt.Sprintf("There was an error reading the file during processing at line #%d. File could not be processed. Please check the contents for errors in formatting and try again", lineNumber))

		// Use map as a makeshift set type by only inputting unique keys. We'll use the length of the map to determine the number of users that have accessed the server.
		// O(# unique users) memory. O(1) lookup for the answer to the query
		currentUserRecord := csvRecord[1]
		_, keyPresent := uniqueUsers[currentUserRecord]
		if !keyPresent {
			uniqueUsers[currentUserRecord] = 1
		}

		// Brute forcing the > and < queries. Going to have O(N) memory usage, O(NlogN)+O(N) lookup which simplifies to O(NlogN).
		// The sorting speed is longer than the lookup speed so that's what wins out in the speed determination.
		// Using a slice to store the values.
		// Sorting the values from least to greatest. sorting speed is O(NlogN)
		// The lookup will be O(N) because we're going to find out where in the slice the
		// separation point is with iteration, and we'll use math on the current index of the lookup and the length of the slice to
		// find our answers. There will be one slice for uploads and one slice for downloads.
		// Going to need to include the duplicates as well in the math. Otherwise we may output the wrong return values.
		// It's going to be that we find an index of the number we're looking for, or the position it would have been placed, and handle
		// it from there. If it's there and there are duplicates, we find the index on the side of the duplicates that is towards the direction we
		// are querying for and do the math. If it's not there then we just take the position it should have been and take the index towards
		// the direction we are querying for. Left for <, right for >.
		// Would have liked to have a fancier solution here but ran out of time.

		// Temp variables for easy reading later on
		currentFileAction := csvRecord[2]
		// CSV read the file sizes as strings so converting them to integers
		currentFileSize, err := strconv.Atoi(csvRecord[3])
		// Error handling the string to int conversion and printing the line number
		check(err, fmt.Sprintf("Failed converting the file size string to int when processing the record on line #%d", lineNumber))

		// Use the right slice depending on the file action.
		if currentFileAction == "upload" {
			uploadSizes = append(uploadSizes, currentFileSize)
		} else if currentFileAction == "download" {
			downloadSizes = append(downloadSizes, currentFileSize)
		}

		// Sort the slices. Default is ascending, which is what we were looking for anyway.
		sort.Ints(uploadSizes)
		sort.Ints(downloadSizes)

		// Increment lineNumber for logging
		lineNumber += 1

	}

	// Inform the user that we've read and processed the file.
	fmt.Printf("Finished reading and processing file. There were %d records total. Starting query lookup...\n", lineNumber)

	// Create boolean that will determine whether or not we want the amount of file sizes above or below our search value
	wantLess := true

	// Get search value from the string. Only works if we are given a correctly formatted metricsQuery.
	// Needs better error handling.
	searchValue := getSearchValueFromQueryString(*metricsQuery)

	// Handle query string parsing, lookups, and inform the user of the results.
	// If the query string was improperly formatted then send the error to the user.
	if strings.Contains(*metricsQuery, "usersAccessed") {
		fmt.Printf("The number of unique users that have accessed the system is: %d\n", len(uniqueUsers))
	} else if strings.Contains(*metricsQuery, "uploadsGreaterThan[") {
		workingIndex := findWorkingIndex(uploadSizes, searchValue, !wantLess)
		fmt.Printf("The number of uploads > %dkB is: %d\n", searchValue, len(uploadSizes[workingIndex:]))
	} else if strings.Contains(*metricsQuery, "uploadsLessThan[") {
		workingIndex := findWorkingIndex(uploadSizes, searchValue, wantLess)
		fmt.Printf("The number of uploads < %dkB is: %d\n", searchValue, len(uploadSizes[:workingIndex]))
	} else if strings.Contains(*metricsQuery, "downloadsGreaterThan[") {
		workingIndex := findWorkingIndex(downloadSizes, searchValue, !wantLess)
		fmt.Printf("The number of downloads > %dkB is: %d\n", searchValue, len(downloadSizes[workingIndex:]))
	} else if strings.Contains(*metricsQuery, "downloadsLessThan[") {
		workingIndex := findWorkingIndex(downloadSizes, searchValue, wantLess)
		fmt.Printf("The number of download < %dkB is: %d\n", searchValue, len(uploadSizes[:workingIndex]))
	}

}
