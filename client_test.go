package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func intInSlice(a int, list []int) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

type XmlRoot struct {
	XMLName  xml.Name  `xml:"root"`
	XMLUsers []XmlUser `xml:"row"`
}

type XmlUser struct {
	XMLName   xml.Name `xml:"row"`
	Id        int      `xml:"id"`
	FirstName string   `xml:"first_name"`
	LastName  string   `xml:"last_name"`
	Age       int      `xml:"age"`
	About     string   `xml:"about"`
	Gender    string   `xml:"gender"`
}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	accessToken := r.Header.Get("AccessToken")
	if len(accessToken) == 0 {
		http.Error(w, "incorrect AccessToken", http.StatusUnauthorized)
	}
	orderField := r.URL.Query().Get("order_field")
	if orderField == "" {
		orderField = "Name"
	}
	availableOrderField := []string{"Id", "Age", "Name"}
	if !stringInSlice(orderField, availableOrderField) {
		jsonResponse, _ := json.Marshal(SearchErrorResponse{Error: "ErrorBadOrderField"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(jsonResponse)
		return
	}
	orderBy, _ := strconv.Atoi(r.URL.Query().Get("order_by"))
	availableOrderBy := []int{OrderByAsc, OrderByAsIs, OrderByDesc}
	if !intInSlice(orderBy, availableOrderBy) {
		jsonResponse, _ := json.Marshal(SearchErrorResponse{Error: ErrorBadOrderField})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(jsonResponse)
		return
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	query := r.URL.Query().Get("query")
	file, _ := os.Open("dataset.xml")
	defer file.Close()
	var xmlRoot XmlRoot
	acceptedUsers := []User{}
	byteData, _ := ioutil.ReadAll(file)
	err := xml.Unmarshal(byteData, &xmlRoot)
	if err != nil {
		panic(err)
	}
	for _, xmlUser := range xmlRoot.XMLUsers {
		name := fmt.Sprintf("%s %s", xmlUser.FirstName, xmlUser.LastName)
		acceptQuery := false
		if query == "" {
			acceptQuery = true
		} else {
			acceptQuery = strings.Contains(name, query) || strings.Contains(xmlUser.About, query)
		}
		if acceptQuery {
			acceptedUsers = append(acceptedUsers, User{
				Id:     xmlUser.Id,
				Name:   name,
				Age:    xmlUser.Age,
				About:  xmlUser.About,
				Gender: xmlUser.Gender,
			})
		}
	}
	if orderBy != 0 {
		sort.Slice(acceptedUsers, func(i, j int) bool {
			switch orderField {
			case "Id":
				partI := acceptedUsers[i].Id
				partJ := acceptedUsers[j].Id
				if orderBy == OrderByAsc {
					return partI < partJ
				} else {
					return partI > partJ
				}
			case "Age":
				partI := acceptedUsers[i].Age
				partJ := acceptedUsers[j].Age
				if orderBy == OrderByAsc {
					return partI < partJ
				} else {
					return partI > partJ
				}
			case "Name":
				partI := acceptedUsers[i].Name
				partJ := acceptedUsers[j].Name
				if orderBy == OrderByAsc {
					return partI < partJ
				} else {
					return partI > partJ
				}
			default:
				panic("order error")
				return false
			}
		})
	}
	var jsonResult []byte
	if offset > len(acceptedUsers) {
		jsonResult, _ = json.Marshal([]User{})
	} else {
		jsonResult, _ = json.Marshal(acceptedUsers[offset:min(limit, len(acceptedUsers))])
	}
	w.Write(jsonResult)
}

func SearchServerInternalServerError(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "server error occurred", http.StatusInternalServerError)
}

func SearchServerBadRequestIncorrectJson(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte{123})
}

func SearchServerOkRequestIncorrectJson(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte{123})
}

func SearchServerTimeOut(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	time.Sleep(2 * time.Second)
	w.Write([]byte{123})
}

func TestAccessTokenError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	clientClient := SearchClient{
		AccessToken: "",
		URL:         ts.URL,
	}
	searchRequest := SearchRequest{}
	response, err := clientClient.FindUsers(searchRequest)
	if err == nil {
		t.Errorf("expected error, got nil, response: %#v", response)
	}
	ts.Close()
}

func TestInternalServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServerInternalServerError))
	clientClient := SearchClient{
		AccessToken: "",
		URL:         ts.URL,
	}
	searchRequest := SearchRequest{}
	response, err := clientClient.FindUsers(searchRequest)
	if err == nil {
		t.Errorf("expected error, got nil, response: %#v", response)
	}
	ts.Close()
}

func TestBadRequestCommonError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	clientClient := SearchClient{
		AccessToken: "secret",
		URL:         ts.URL,
	}
	searchRequest := SearchRequest{OrderField: "incorrect"}
	response, err := clientClient.FindUsers(searchRequest)
	if err == nil {
		t.Errorf("expected error, got nil, response: %#v", response)
	}
	searchRequest = SearchRequest{OrderBy: 999}
	response, err = clientClient.FindUsers(searchRequest)
	if err == nil {
		t.Errorf("expected error, got nil, response: %#v", response)
	}
	ts.Close()
}

func TestBadRequestIncorrectResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServerBadRequestIncorrectJson))
	clientClient := SearchClient{
		AccessToken: "secret",
		URL:         ts.URL,
	}
	searchRequest := SearchRequest{OrderBy: 1}
	response, err := clientClient.FindUsers(searchRequest)
	if err == nil {
		t.Errorf("expected error, got nil, response: %#v", response)
	}
	ts.Close()

	ts = httptest.NewServer(http.HandlerFunc(SearchServerOkRequestIncorrectJson))
	clientClient = SearchClient{
		AccessToken: "secret",
		URL:         ts.URL,
	}
	searchRequest = SearchRequest{OrderBy: 1}
	response, err = clientClient.FindUsers(searchRequest)
	if err == nil {
		t.Errorf("expected error, got nil, response: %#v", response)
	}
	ts.Close()
}

func TestTimeOutError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServerTimeOut))
	clientClient := SearchClient{
		AccessToken: "secret",
		URL:         ts.URL,
	}
	searchRequest := SearchRequest{OrderBy: 1}
	response, err := clientClient.FindUsers(searchRequest)
	if err == nil {
		t.Errorf("expected error, got nil, response: %#v", response)
	}
	ts.Close()
}

func TestUnknownError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	clientClient := SearchClient{
		AccessToken: "secret",
		URL:         ts.URL + "123",
	}
	searchRequest := SearchRequest{OrderBy: 1}
	response, err := clientClient.FindUsers(searchRequest)
	if err == nil {
		t.Errorf("expected error, got nil, response: %#v", response)
	}
	ts.Close()

}

func TestOffsetAndLimitError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	clientClient := SearchClient{
		AccessToken: "secret",
		URL:         ts.URL,
	}
	searchRequest := SearchRequest{OrderBy: 1, Offset: -1}
	response, err := clientClient.FindUsers(searchRequest)
	if err == nil {
		t.Errorf("expected error, got nil, response: %#v", response)
	}
	searchRequest = SearchRequest{OrderBy: 1, Limit: -1}
	response, err = clientClient.FindUsers(searchRequest)
	if err == nil {
		t.Errorf("expected error, got nil, response: %#v", response)
	}
	ts.Close()
}

func TestSuccessCases(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	clientClient := SearchClient{
		AccessToken: "secret",
		URL:         ts.URL,
	}
	searchRequest := SearchRequest{Limit: 99, Offset: 0}
	response, err := clientClient.FindUsers(searchRequest)
	if err != nil {
		t.Errorf("unexpected error, response: %#v", response)
	}
	searchRequest = SearchRequest{Limit: 20, Offset: 10, Query: "Boyd"}
	response, err = clientClient.FindUsers(searchRequest)
	if err != nil {
		t.Errorf("unexpected error, response: %#v", response)
	}
	searchRequest = SearchRequest{Limit: 3, Query: "moll", OrderField:"Age", OrderBy:OrderByAsc}
	response, err = clientClient.FindUsers(searchRequest)
	if err != nil {
		t.Errorf("unexpected error, response: %#v", response)
	}
	if !sort.SliceIsSorted(response.Users, func(i, j int) bool {
		return response.Users[i].Age < response.Users[j].Age
	}) {
		t.Errorf("unexpected error, incorrect sorted: %#v", response)
	}

	searchRequest = SearchRequest{Limit: 3, Query: "ita", OrderField:"Name", OrderBy:OrderByDesc}
	response, err = clientClient.FindUsers(searchRequest)
	if err != nil {
		t.Errorf("unexpected error, response: %#v", response)
	}
	if !sort.SliceIsSorted(response.Users, func(i, j int) bool {
		return response.Users[i].Name > response.Users[j].Name
	}) {
		t.Errorf("unexpected error, incorrect sorted: %#v", response)
	}
	ts.Close()
}
