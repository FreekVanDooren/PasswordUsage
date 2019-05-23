package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

const PATH = "/password-checker/"

func StartServer(debugMode bool, portNr int, logger *log.Logger) {
	logger.Print("Started with debugmode <", debugMode, ">")
	http.HandleFunc(PATH, CheckPassword(debugMode, logger))
	err := http.ListenAndServe(fmt.Sprint(":", portNr), nil)
	if err != nil {
		panic(err)
	}
}

func CheckPassword(debugMode bool, logger *log.Logger) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		respondWithError := func(statusCode int, err error) {
			http.Error(w, err.Error(), statusCode)
		}
		isDebugInfoSuppliable := func() bool {
			if debugMode {
				query := r.URL.Query()
				debugParameter := query.Get("debug")
				if query["debug"] != nil && debugParameter == "" {
					return true
				}
				userRequested, _ := strconv.ParseBool(debugParameter)
				return userRequested
			}
			return false
		}
		if r.Method == http.MethodPost {
			http.NotFound(w, r)
			return
		}

		password := strings.TrimPrefix(r.URL.Path, PATH)
		hashedPassword := toHash([]byte(password))
		debugInfoSuppliable := isDebugInfoSuppliable()
		logger.Print("Password <", password, "> hashed as <", hashedPassword, "> debug info returned <", debugInfoSuppliable, ">")

		if occurrences, pwndErr := defaultPasswordRequest(hashedPassword); pwndErr == nil {
			responseHash := ""
			if debugInfoSuppliable {
				responseHash = hashedPassword
			}
			if response, jsonError := json.Marshal(
				struct {
					Occurrences int
					Hash        string `json:",omitempty"`
				}{
					Occurrences: occurrences,
					Hash:        responseHash,
				}); jsonError == nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(response)
			} else {
				respondWithError(http.StatusFailedDependency, jsonError)
			}
		} else {
			respondWithError(http.StatusUnprocessableEntity, pwndErr)
		}
	}
}
