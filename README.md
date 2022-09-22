# server_log_reader
A GoLang command line utility that allows the user to query csv server_logs for an
upload/download server.

Go version used to write this utility is: `go1.9.3 darwin/amd64`
This utility relies only on the Go standard library and has zero external dependencies

## User Guide

The utility is written as a single GoLang file that will be run with Go and takes in two
command line arguments. The arguments are flag typed so you must specify the flag before
you provide the value for that flag. The following command shows an example for the flags
with valid values:
`go server_log_reader.go -logFilePath server_log.csv -metricsQuery #usersAccessed`

The utility reads the file a single time and creates data structures that are then used to
return the responses to your queries. You will need to specify your query when you use
the utility. If values are not provided for your arguments then

Only supports the following queries (> and < query values need to be in kB):
* `#usersAccessed`
* `#uploads>[value]`
* `#uploads<[value]`
* `#downloads>[value]`
* `#downloads<[value]`

The following queries were unsupported due to lack of time:
* `#userUploads[date{MM/DD/YYYY}]`
* `#userDownloads[date{MM/DD/YYYY}]`


## Things that would have been handled with more time

* Much more additional manual testing. Large program and didn't have a ton of interest in testing as the amount of
		time spent was larger than recommended
* Fix bugs after manual testing
* Fix line number bug for logging
* Error handling of a plethora of edge cases. Especially possibilities of index failures with the > and < queries
		 and the correct math checks on splice values using the working index
* Finishing up the date queries
* Add >= and <= query support
* Create additional helper methods for ease of reading/reuse
* Having string messages span multiple lines instead of a super long line message
* Better metricsQuery string checking for early termination before processing
* Better naming for the project and utility script. I like `file_server_csv_log_reader.go` better
* Better data structures for data processing and query result output. This solution
works but is not the most efficient and doesn't meet the final two query types
* Adding support for more types of query options
* Full unit testing of the utility using the go `testing` package
* More expressive command line output
