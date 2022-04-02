# hw4_test_coverage

This is a combined task on how to send requests, receive responses, work with parameters, headers, and also write tests.

The task is not difficult, the main amount of work is writing different conditions and tests to satisfy these conditions.

We have some kind of search service:
* SearchClient - a structure with the FindUsers method, which sends a request to an external system and returns the result, slightly transforming it
* SearchServer is a kind of external system. Directly searches for data in the `dataset.xml` file. In production, it would run as a separate web service.

Required:
* Write a SearchServer function in the `client_test.go` file that you will run through the test server (`httptest.NewServer`, use case in `http/server_test.go`)
* Cover the FindUsers method with tests so that the coverage is as large as possible, namely 100%. Write tests in `client_test.go`.
* It is also required to generate an html-report with coverage. See an example of building test coverage and a report in the corresponding lecture (I remind you that for this your code must be inside GOPATH).

Additionally:
* Data for work is in the file `dataset.xml`
* * The `query` parameter searches the fields `Name` and `About`
* The `order_field` parameter works on the fields `Id`, `Age`, `Name`, if empty - then return by `Name`, if something else - SearchServer swears by an error. `Name` is first_name + last_name from xml.
* If `query` is empty, then we do only sorting, i.e. return all records
* The code must be written in the client_test.go file. Your tests will be there, as well as the SearchServer function.
* See `xml/*` for how to work with XML
* Run as `go test -cover`
* Build coverage: `go test -coverprofile=cover.out && go tool cover -html=cover.out -o cover.html`. To build coverage, your code must be inside GOPATH
