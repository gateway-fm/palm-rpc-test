package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/nsf/jsondiff"
)

var (
	client = &http.Client{}
)

func main() {
	host1 := flag.String("host1", "http://127.0.0.1:8545", "address of first RPC host")
	host2 := flag.String("host2", "http://127.0.0.1:8546", "address of second RPC host")
	consoleOut := flag.Bool("console", false, "output results to console as processing happens")
	flag.Parse()

	inputFiles, err := os.ReadDir("./input")
	if err != nil {
		fmt.Printf("Error reading input directory: %v\n", err)
		return
	}

	consoleOptions := jsondiff.DefaultConsoleOptions()
	markdownOptions := jsondiff.DefaultJSONOptions()
	markdownOptions.SkipMatches = true
	var markdownOutput string

	for _, inputFile := range inputFiles {
		fmt.Printf("Processing %s\n", inputFile.Name())

		path := fmt.Sprintf("./input/%s", inputFile.Name())
		file, err := os.OpenFile(path, os.O_RDONLY, 0644)
		fileBytes, err := io.ReadAll(file)
		if err != nil {
			fmt.Printf("Error reading input file: %v\n", err)
			return
		}

		if err != nil {
			fmt.Printf("Error reading input file: %v\n", err)
			return
		}

		res1, err := getResponse(*host1, fileBytes)
		if err != nil {
			fmt.Printf("Error getting response from host1: %v\n", err)
			return
		}

		res2, err := getResponse(*host2, fileBytes)
		if err != nil {
			fmt.Printf("Error getting response from host2: %v\n", err)
			return
		}

		// first marshall json to an RpcError struct to see if we got an error - it might be that the node
		// does not support the call in which case comparison isn't useful
		var possibleError PossibleError
		err = json.Unmarshal(res1, &possibleError)
		if err != nil {
			fmt.Printf("Error unmarshalling response from host1: %v\n", err)
			return
		}

		if strings.Contains(possibleError.Error.Message, "does not exist") {
			fmt.Println("Host1 does not support this call")
			continue
		}

		err = json.Unmarshal(res2, &possibleError)
		if err != nil {
			fmt.Printf("Error unmarshalling response from host2: %v\n", err)
			return
		}

		if strings.Contains(possibleError.Error.Message, "does not exist") {
			fmt.Println("Host2 does not support this call")
			continue
		}

		diff, report := jsondiff.Compare(res1, res2, &consoleOptions)
		if diff == jsondiff.FullMatch {
			fmt.Println("Files match")
		} else {
			if *consoleOut {
				fmt.Println(report)
			}
			fmt.Println("!!! Files do not match")

			_, report = jsondiff.Compare(res1, res2, &markdownOptions)
			markdownOutput += fmt.Sprintf("# File: %s\n", inputFile.Name())
			markdownOutput += "```json\n"
			markdownOutput += fmt.Sprintf("%s\n", report)
			markdownOutput += "```\n\n\n"
		}
	}

	_ = os.WriteFile("./output/output.md", []byte(markdownOutput), 0644)
}

func getResponse(host string, contents []byte) ([]byte, error) {
	br := bytes.NewReader(contents)
	req, err := http.NewRequest(http.MethodPost, host, br)
	if err != nil {
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return b, nil
}

type RpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type PossibleError struct {
	Error RpcError `json:"error"`
}
