package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"
)

func TestGETPasswordChecker(t *testing.T) {
	pageNotFound := "404 page not found\n"
	tests := []struct {
		name   string
		path   string
		status int
		body   string
	}{
		{
			name:   "No path",
			path:   "",
			status: http.StatusNotFound,
			body:   pageNotFound,
		},
		{
			name:   "Root",
			path:   "/",
			status: http.StatusNotFound,
			body:   pageNotFound,
		},
		{
			name:   "Some path",
			path:   "/some/path",
			status: http.StatusNotFound,
			body:   pageNotFound,
		},
		{
			name:   "Check password",
			path:   "/password-checker/password",
			status: http.StatusOK,
			body:   "{\"Occurrences\":3645804}",
		},
		{
			name:   "Check Password explicit no debug",
			path:   "/password-checker/Password?debug=false",
			status: http.StatusOK,
			body:   "{\"Occurrences\":117316}",
		},
		{
			name:   "Check password with debug (no value)",
			path:   "/password-checker/password?debug",
			status: http.StatusOK,
			body:   "{\"Occurrences\":3645804,\"Hash\":\"5BAA61E4C9B93F3F0682250B6CF8331B7EE68FD8\"}",
		},
		{
			name:   "Check Password with debug",
			path:   "/password-checker/Password?debug=true",
			status: http.StatusOK,
			body:   "{\"Occurrences\":117316,\"Hash\":\"8BE3C943B1609FFFBFC51AAD666D0A04ADF83C9D\"}",
		},
	}
	portNr := 6542
	go StartServer(true, portNr, log.New(os.Stdout, "", 0))
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			url := fmt.Sprint("http://localhost:", portNr, test.path)

			doGETRequest := func() {
				getResp, getErr := http.Get(url)
				if getErr != nil {
					t.Fatal("Something went horribly wrong", getErr)
				}
				defer getResp.Body.Close()
				if getResp.StatusCode != test.status {
					t.Error("Expected", test.status, "for GET method, but got:", getResp.StatusCode)
				}
				getBody := readBody(getResp)
				if getBody != test.body {
					t.Error("Unexpected response", getBody)
				}
			}

			doPOSTRequest := func() {
				postResp, postErr := http.Post(url, " text/plain; charset=utf-8", nil)
				if postErr != nil {
					t.Fatal("Something went horribly wrong", postErr)
				}
				defer postResp.Body.Close()
				if postResp.StatusCode != http.StatusNotFound {
					t.Error("Expected POST method not found, but got:", postResp.StatusCode)
				}
				postBody := readBody(postResp)
				if postBody != pageNotFound {
					t.Error("Unexpected response", postBody)
				}
			}

			doGETRequest()
			doPOSTRequest()
		})
	}
}
