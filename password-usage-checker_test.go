package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func TestOnlyContainsValidCharacters(t *testing.T) {
	asciiSet := make([]byte, 256)
	for i := range asciiSet {
		asciiSet[i] = byte(i)
	}

	testCharacters := func(characters []byte, assert func([]byte)) {
		for _, ascii := range characters {
			input := []byte{ascii}
			assert(input)
		}
	}

	if containsInvalidCharacters([]byte{}) {
		t.Error("Empty array is not necessarily invalid")
	}

	unacceptableCharacters := append(append([]byte{}, asciiSet[:33]...), asciiSet[127:]...)
	testCharacters(unacceptableCharacters, func(input []byte) {
		if !containsInvalidCharacters(input) {
			t.Errorf("Cannot accept this ASCII character as password input: %d (%s)", input, input)
		}
	})

	acceptableCharacters := asciiSet[33:127]
	testCharacters(acceptableCharacters, func(input []byte) {
		if containsInvalidCharacters(input) {
			t.Error("This is a valid ASCII character:", input)
		}
	})
}

func TestToHash(t *testing.T) {
	actual := toHash([]byte("Password"))
	if "8BE3C943B1609FFFBFC51AAD666D0A04ADF83C9D" != actual {
		t.Error("Expected it not to be:", actual)
	}
}

func TestNumberOfOccurrences(t *testing.T) {
	faultyInput := "sdfds"
	_, err := numberOfOccurrences([]string{faultyInput}, "")
	if err == nil {
		t.Error("Expected failure on ill-formatted hash")
	}

	testNumberOfOccurences := func(matchHash string, assert func(int)) {
		hashes := []string{"8348574395:2"}
		actual, err := numberOfOccurrences(hashes, matchHash)
		if err != nil {
			t.Error("An error?\n", err)
		}
		assert(actual)
	}
	testNumberOfOccurences("8348574395", func(actual int) {
		if actual != 2 {
			t.Error("Expected to find match")
		}
	})
	testNumberOfOccurences("232", func(actual int) {
		if actual != 0 {
			t.Error("Did not expect to find match")
		}
	})
}

func TestFindNumberOfOccurences(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "No response",
			input:    "",
			expected: 0,
		}, {
			name:     "None matching result",
			input:    "HASH:3",
			expected: 0,
		}, {
			name:     "One match",
			input:    "HASH1:20",
			expected: 20,
		}, {
			name:     "Multiple hashes",
			input:    "HASH0:99\r\nHASH1:20\r\nHASH2:3",
			expected: 20,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			prefix := "12345"
			serverMock := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
				if req.RequestURI != "/"+prefix {
					t.Error("Hash not divided as expected:", req.RequestURI)
				}
				_, _ = fmt.Fprint(writer, test.input)
			}))
			defer serverMock.Close()

			actual, e := findNumberOfOccurences(PwndAPI{
				Client:         serverMock.Client(),
				baseURL:        serverMock.URL,
				hashedPassword: prefix + "HASH1"})
			if e != nil {
				t.Fatal("Did not expect this:", e)
			}

			if actual != test.expected {
				t.Error(actual, "!=", test.expected)
			}
		})
	}
}

func BenchmarkNumberOfOccurrences(b *testing.B) {
	inputs := []string{
		"76sdftsd",
		"83485",
		"834896",
		"8348397",
		"83484398",
		"834874399",
	}
	foundHashes := make([]string, len(inputs)-1)
	for i := 1; i < len(inputs); i++ {
		foundHashes[i-1] = fmt.Sprint(inputs[i], ":", i)
	}
	results := make([]int, len(inputs))
	var err error
	for i := 0; i < b.N; i++ {
		for i, hash := range inputs {
			results[i], err = numberOfOccurrences(foundHashes, hash)
			if err != nil {
				b.Error("An error?\n", err)
			}
		}
		_, err = numberOfOccurrences([]string{"badformatting"}, "badform")
		if err == nil {
			b.Error("Error expected on badformatting")
		}
	}
	b.Log("Results:", results)
}

func BenchmarkCompareSplitToRegex(bench *testing.B) {
	benchmark := func(testFunc func(string) (int, error)) func(*testing.B) {
		return func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				for hash, expected := range map[string]int{
					"UNKNOWN":                             0,
					"00CBB2A6F377FACA93FA20D1A7D4D68EF9C": 1,
					"713051F44A50A2FB2EB852CE9963F939CA2": 2,
					"BB0826452CF131AA969A07BA903B78E9CC3": 59,
					"943B1609FFFBFC51AAD666D0A04ADF83C9D": 117316,
					"FF89B26A63C84B485152AF60CFF336B054B": 4,
				} {
					result, err := testFunc(hash)
					if err != nil {
						b.Error("An error?\n", err)
					}
					if result != expected {
						b.Error("Expected result to be", expected, " but was", result)
					}
				}
			}
		}
	}
	fileContent, e := ioutil.ReadFile("benchmarkNumberOfOccurences.txt")
	if e != nil {
		panic(e)
	}
	input := string(fileContent)
	withSplit := func(hash string) (int, error) {
		// Couldn't be bothered to create a file with Windows line endings.
		return numberOfOccurrences(strings.Split(input, "\n"), hash)
	}
	withRegex := func(hash string) (int, error) {
		reg := regexp.MustCompile(fmt.Sprint(hash, ":(\\d+)"))
		matches := reg.FindSubmatch(fileContent)
		if matches == nil {
			return 0, nil
		}
		return strconv.Atoi(string(matches[1]))
	}
	bench.Run("With Split", benchmark(withSplit)) // clear winner!
	bench.Run("With Regex", benchmark(withRegex)) // No bad format checking either
}
