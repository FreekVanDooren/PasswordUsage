package main

import (
	"crypto/sha1"
	"flag"
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

type PwndAPI struct {
	Client         *http.Client
	baseURL        string
	hashedPassword string
}

func main() {
	debug := flag.Bool("debug", false, "Shows additional information about user input")
	port := flag.Int("port", 6543, "Port number for server to listen to. Default is 6543")
	flag.Parse()

	serverLoggger, serverFile := setupLogging("server.log")
	defer serverFile.Close()
	go StartServer(*debug, *port, serverLoggger)

	cmdLogger, cmdFile := setupLogging("cmd.log")
	defer cmdFile.Close()
	go CheckPasswordUsage("Hi there!", *debug, cmdLogger)

	select {} // making sure program doesn't terminate
}

func setupLogging(fileName string) (logger *log.Logger, logFile *os.File) {
	if dir, wdError := os.Getwd(); wdError == nil {
		logDirectory := filepath.Join(dir, "logs")
		if createDirError := os.MkdirAll(logDirectory, os.ModePerm); createDirError == nil {
			logFilePath := filepath.Join(logDirectory, fileName)
			log.Println("Logging files in:", logFilePath)

			if logFile, openFileError := os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm); openFileError == nil {
				return log.New(logFile, "Password Checker: ", log.LstdFlags), logFile
			} else {
				panic(openFileError)
			}
		} else {
			panic(createDirError)
		}
	} else {
		panic(wdError)
	}
}

func CheckPasswordUsage(welcomeText string, debug bool, logger *log.Logger) {
	fmt.Println(welcomeText)
	fmt.Println()
	fmt.Printf("Please enter the password you'd like to check:")

	password, _ := terminal.ReadPassword(int(syscall.Stdin)) // doesn't work in Goland
	fmt.Println()
	hashedPassword := toHash(password)
	if debug {
		fmt.Println("This is the bytes entered:", password)
		fmt.Println("Which as string is:", string(password))
		fmt.Println("And it's hashed as:", hashedPassword)
	}
	fmt.Println()
	occurrences, humanReadable := validateAndCheck(password, hashedPassword)
	logger.Printf("{bytes: %v, ascii: %s, hash: %v, occurrences: %v}", password, password, hashedPassword, occurrences)
	fmt.Println(humanReadable)
	CheckPasswordUsage("Try again?", debug, logger)
}

func validateAndCheck(password []byte, hashedPassword string) (occurrences int, humanReadable string) {
	if len(password) == 0 {
		humanReadable = "Did you accidentally hit enter?\n"
		return
	}
	if containsInvalidCharacters(password) {
		humanReadable = "You naughty! That password has a special key.\n"
		return
	}
	fmt.Println("Checking database...")
	var err error
	if occurrences, err = defaultPasswordRequest(hashedPassword); err == nil {
		if occurrences > 0 {
			humanReadable = fmt.Sprint("Unfortunately, it has been used ", occurrences, " times")
		} else {
			humanReadable = "Luckily, it could not be found"
		}
		return
	} else {
		panic(err)
	}
}

func containsInvalidCharacters(bytePw []byte) bool {
	const upperlimit = byte(126) // ~
	const lowerlimit = byte(33)  // !
	for _, b := range bytePw {
		if b > upperlimit || b < lowerlimit {
			return true
		}
	}
	return false
}

func defaultPasswordRequest(hashedPassword string) (int, error) {
	return findNumberOfOccurences(
		PwndAPI{
			Client:         http.DefaultClient,
			baseURL:        "https://api.pwnedpasswords.com/range",
			hashedPassword: hashedPassword,
		})
}

func toHash(password []byte) string {
	return fmt.Sprintf("%X", sha1.Sum(password))
}

func findNumberOfOccurences(api PwndAPI) (int, error) {
	response, e := api.Client.Get(fmt.Sprint(api.baseURL, "/", api.hashedPassword[0:5]))
	if e != nil {
		return -2, e
	}
	defer response.Body.Close()
	body := readBody(response)
	return numberOfOccurrences(strings.Split(body, "\r\n"), api.hashedPassword[5:])
}

func numberOfOccurrences(hashes []string, matchHash string) (int, error) {
	for _, hash := range hashes {
		if strings.HasPrefix(hash, matchHash) {
			separatorIndex := strings.Index(hash, ":")
			if separatorIndex >= 0 {
				return strconv.Atoi(hash[separatorIndex+1:])
			} else {
				return -1, fmt.Errorf("Unexpected format of PWNed response:\n%s", hashes)
			}
		}
	}
	return 0, nil
}

func readBody(response *http.Response) string {
	bytes, _ := ioutil.ReadAll(response.Body)
	return string(bytes)
}
