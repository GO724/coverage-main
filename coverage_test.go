package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

// тут писать код тестов

// error client.go  101-131
const (
	SERVER_AUTHORIZED     = ""
	SERVER_UNAUTHORIZED   = "bad AccessToken"
	SERVER_INTERNAL_ERROR = "SearchServer fatal error | unknown error"
	SERVER_TIMEOUT        = "timeout for"
	SERVER_BAD_JSON       = "cant unpack error json: | cant unpack result json:"
	SERVER_BAD_ORDERFIELD = "OrderFeld"
	SERVER_BAD_REQUEST    = "unknown bad request error:"
)

// ***************************** TestSearchServer_Auth *****************************

// Authorization test

func TestSearchServer_Auth(t *testing.T) {

	// test server
	ts := httptest.NewServer(http.HandlerFunc(SearchServer)) // server.go
	defer ts.Close()

	// type SearchClient struct in client.go
	client := SearchClient{
		// AccessToken: "",
		// URL:         "",
	}

	// type SearchRequest struct in client.go
	request := SearchRequest{
		// Limit:      0,
		// Offset:     0,
		// Query:      "",
		// OrderField: "",
		// OrderBy:    0,
	}

	// test cases
	casesAuth := map[string]struct {
		token       string
		expectedErr string
	}{
		"valid token": {
			token:       "valid_token",
			expectedErr: SERVER_AUTHORIZED,
		},
		"invalid token": {
			token:       "invalid_token",
			expectedErr: SERVER_UNAUTHORIZED,
		},
		"empty token": {
			token:       "",
			expectedErr: SERVER_UNAUTHORIZED,
		},
	}

	for name, tc := range casesAuth {
		t.Run(name, func(t *testing.T) {
			client.AccessToken = tc.token
			client.URL = ts.URL
			_, err := client.FindUsers(request)
			if !checkError(err, tc.expectedErr) {
				t.Errorf("token(%s) expected %s, got %v", tc.token, tc.expectedErr, err)
			}
		})
	}
}

// ***************************** TestSearchServer_Pagination *****************************

// Pagination test (server)
func TestSearchServer_Pagination(t *testing.T) {

	// test server
	ts := httptest.NewServer(http.HandlerFunc(SearchServer)) // server.go
	defer ts.Close()

	// type SearchClient struct in client.go
	client := SearchClient{
		AccessToken: "valid_token",
		URL:         ts.URL,
	}

	testPagination := []struct {
		name             string         // test name
		mokDataFile      string         // MOK data file
		client           SearchClient   // client param
		request          SearchRequest  // request param
		expectedResponse SearchResponse // response data
		expectedError    error          // response error
	}{
		{
			name:        "limit < 0",
			mokDataFile: "TestClientFindUsers.xml",
			request: SearchRequest{
				Limit:      -1,
				Offset:     0,
				Query:      "",
				OrderField: "",
				OrderBy:    0,
			},
			expectedResponse: SearchResponse{
				Users:    []User{},
				NextPage: false,
			},
			expectedError: fmt.Errorf("limit must be > 0"),
		},
		{
			name:        "offset < 0",
			mokDataFile: "TestClientFindUsers.xml",
			request: SearchRequest{
				Limit:      2,
				Offset:     -1,
				Query:      "",
				OrderField: "",
				OrderBy:    0,
			},
			expectedResponse: SearchResponse{
				Users:    []User{},
				NextPage: false,
			},
			expectedError: fmt.Errorf("offset must be > 0"),
		},
		{
			name:        "Limit < 0",
			mokDataFile: "TestClientFindUsers.xml",
			request: SearchRequest{
				Limit:      -1,
				Offset:     0,
				Query:      "",
				OrderField: "",
				OrderBy:    0,
			},
			expectedResponse: SearchResponse{
				Users:    []User{},
				NextPage: false,
			},
			expectedError: fmt.Errorf("limit must be > 0"),
		},
	}

	saveDataset := DataFile

	for testID, tc := range testPagination {

		DataFile = tc.mokDataFile

		t.Run(tc.name, func(t *testing.T) {
			response, err := client.FindUsers(tc.request)
			if errors.Is(err, tc.expectedError) {
				t.Errorf("Wrong result[%d]\nExpected: %+v\nGot: %+v", testID, tc.expectedError, err)
			}
			if response != nil && !reflect.DeepEqual(*response, tc.expectedResponse) {
				t.Errorf("Wrong result\nExpected: %+v\nGot: %+v", tc.expectedResponse, *response)
			}
		})
	}

	DataFile = saveDataset

}

// ***************************** TestSearchServer_Sort *****************************

// Sort test
// Параметр `order_field` работает по полям `Id`, `Age`, `Name`, если пустой - то возвращаем по `Name`, если что-то другое - SearchServer ругается ошибкой.
// Параметр `order_by` задает направление сортировки (по полю переданному в `order_field`) или ее отсутствие (OrderByAsIs)
// Если `query` пустой, то делаем только сортировку, т.е. возвращаем все записи

func TestSearchServer_Sort(t *testing.T) {

	// test server
	ts := httptest.NewServer(http.HandlerFunc(SearchServer)) // server.go
	defer ts.Close()

	// type SearchClient struct in client.go
	client := SearchClient{
		AccessToken: "valid_token",
		URL:         ts.URL,
	}

	// test cases
	casesQuery := map[string]struct {
		mokDataFile           string
		request               SearchRequest
		expectedResponce      SearchResponse
		expectedResponceError string
	}{
		"sort by Id": {
			mokDataFile: "dataset.xml",
			// type SearchRequest struct in client.go
			request: SearchRequest{
				Limit:      2,
				Offset:     0,
				Query:      "",
				OrderField: "Id",
				OrderBy:    -1,
			},
			// type expectedResponce struct in client.go
			expectedResponce: SearchResponse{
				Users:    getExpectedUsers("dataset.xml", 34, 33),
				NextPage: true,
			},
			expectedResponceError: "",
		},
		"sort by Age": {
			mokDataFile: "dataset.xml",
			// type SearchRequest struct in client.go
			request: SearchRequest{
				Limit:      2,
				Offset:     0,
				Query:      "Jennings",
				OrderField: "Age",
				OrderBy:    -1,
			},
			// type expectedResponce struct in client.go
			expectedResponce: SearchResponse{
				Users:    getExpectedUsers("dataset.xml", 6),
				NextPage: false,
			},
			expectedResponceError: "",
		},
		"sort by Name": {
			mokDataFile: "dataset.xml",
			// type SearchRequest struct in client.go
			request: SearchRequest{
				Limit:      2,
				Offset:     0,
				Query:      "Jennings",
				OrderField: "Name",
				OrderBy:    -1,
			},
			// type expectedResponce struct in client.go
			expectedResponce: SearchResponse{
				Users:    getExpectedUsers("dataset.xml", 6),
				NextPage: false,
			},
			expectedResponceError: "",
		},
		"invalid sort params": {
			mokDataFile: "dataset.xml",
			// type SearchRequest struct in client.go
			request: SearchRequest{
				Limit:      2,
				Offset:     0,
				Query:      "Jennings",
				OrderField: "About",
				OrderBy:    1,
			},
			// type expectedResponce struct in client.go
			expectedResponce: SearchResponse{
				Users:    getExpectedUsers("dataset.xml", 0, 1),
				NextPage: true,
			},
			expectedResponceError: SERVER_BAD_ORDERFIELD,
		},
	}

	var response *SearchResponse
	var err error

	saveDataset := DataFile

	for name, tc := range casesQuery {

		DataFile = tc.mokDataFile

		t.Run(name, func(t *testing.T) {
			response, err = client.FindUsers(tc.request)

			if !checkError(err, tc.expectedResponceError) {
				t.Errorf("client.FindUsers unexpected error : %+v", err)
			}
			if response != nil && !reflect.DeepEqual(*response, tc.expectedResponce) {
				t.Errorf("Wrong result\nExpected: %+v\nGot: %+v", tc.expectedResponce, *response)
			}
		})
	}

	DataFile = saveDataset
}

// ***************************** TestSearchServer_Query *****************************

// Query test
// Параметр `query` ищет по полям `Name` и `About`
// Если `query` пустой, то делаем только сортировку, т.е. возвращаем все записи

func TestSearchServer_Query(t *testing.T) {

	// test server
	ts := httptest.NewServer(http.HandlerFunc(SearchServer)) // server.go
	defer ts.Close()

	// type SearchClient struct in client.go
	client := SearchClient{
		AccessToken: "valid_token",
		URL:         ts.URL,
	}

	// test cases
	casesQuery := map[string]struct {
		mokDataFile      string
		request          SearchRequest
		expectedResponce SearchResponse
	}{
		"name query": {
			mokDataFile: "dataset.xml",
			// type SearchRequest struct in client.go
			request: SearchRequest{
				Limit:      2,
				Offset:     0,
				Query:      "Jennings",
				OrderField: "",
				OrderBy:    0,
			},
			// type expectedResponce struct in client.go
			expectedResponce: SearchResponse{
				Users:    getExpectedUsers("dataset.xml", 6),
				NextPage: false,
			},
		},
		"empty query": {
			mokDataFile: "dataset.xml",
			// type SearchRequest struct in client.go
			request: SearchRequest{
				Limit:      2,
				Offset:     0,
				Query:      "",
				OrderField: "",
				OrderBy:    0,
			},
			// type expectedResponce struct in client.go
			expectedResponce: SearchResponse{
				Users:    getExpectedUsers("dataset.xml", 0, 1),
				NextPage: true,
			},
		},
	}

	var response *SearchResponse
	var err error

	saveDataset := DataFile

	for name, tc := range casesQuery {

		DataFile = tc.mokDataFile

		t.Run(name, func(t *testing.T) {
			response, err = client.FindUsers(tc.request)
			if err != nil {
				t.Errorf("client.FindUsers unexpected error : %+v", err)
			}
			if response != nil && !reflect.DeepEqual(*response, tc.expectedResponce) {
				t.Errorf("Wrong result\nExpected: %+v\nGot: %+v", tc.expectedResponce, *response)
			}
		})
	}

	DataFile = saveDataset
}

func getExpectedUsers(dataFile string, validID ...int) []User {
	mokDataSet, err := loadDataset(dataFile)
	if err != nil {
		return []User{}
	}
	if len(validID) == 0 {
		return mokDataSet
	}
	expectedUsers := make([]User, 0, 100)
	for i := range validID {
		expectedUsers = append(expectedUsers, mokDataSet[validID[i]])
	}
	return expectedUsers
}

func checkError(err error, s string) bool {
	if err == nil && len(s) == 0 {
		return true
	}
	if strings.Contains(fmt.Sprintf("%v", err), s) {
		return true
	}

	return false
}
