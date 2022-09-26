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
`go server_log_reader.go -logFilePath server_log.csv -metricsQuery uploadsGreaterThan[50]`

The utility reads the file a single time and creates data structures that are then used to
return the responses to your queries. You will need to specify your query when you use
the utility. If values are not provided for your arguments then

Only supports the following queries (`lessThan` and `greaterThan` query values need to be in `kB`):
* `usersAccessed`
* `uploadsGreaterThan[value]`
* `uploadsLessThan[value]`
* `downloadsGreaterThan[value]`
* `downloadsLessThan[value]`
* `uploadsByUser[userName]OnDate[MM DD YYYY]`
* `downloadsByUser[userName]OnDate[MM DD YYYY]`

Tested using shell scripting techniques with the following commands:
`cat, grep, wc, sort`


## Things that would have been handled with more time

* The counts still are not quite right. Would figure out the specifics with more time.
* Error handling of a plethora of edge cases.
   * Verifying data in file is matching correct formats
* Add >= and <= query support
* Add case insensitive query string support
* Having string messages span multiple lines instead of a super long line message
* Better metricsQuery string checking for early termination before processing
* Better naming for the project and utility script. I like `file_server_csv_log_reader.go` better
* Better data structures for data processing and query result output. This solution
works but is not the most efficient and doesn't meet the final two query types
* Adding support for more types of query options
* Full unit testing of the utility using the go `testing` package
* More expressive command line output
