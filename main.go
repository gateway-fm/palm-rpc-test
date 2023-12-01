package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/nsf/jsondiff"
)

var (
	client = &http.Client{}
)

var (
	consoleOptions  = jsondiff.DefaultConsoleOptions()
	markdownOptions = jsondiff.DefaultJSONOptions()
	consoleOut      = false
)

func main() {
	host1 := flag.String("host1", "http://127.0.0.1:8545", "address of first RPC host")
	host2 := flag.String("host2", "http://127.0.0.1:8546", "address of second RPC host")
	folder := flag.String("folder", "./input", "folder containing input files")
	flag.BoolVar(&consoleOut, "console", false, "output results to console as processing happens")
	flag.Parse()

	markdownOptions.SkipMatches = true

	inputFiles, err := os.ReadDir(*folder)
	if err != nil {
		fmt.Printf("Error reading input directory: %v\n", err)
		return
	}

	var markdownOutput string

	for _, inputFile := range inputFiles {
		filename := inputFile.Name()
		linkedFile := strings.ReplaceAll(filename, ".json", ".md")
		fmt.Printf("Processing %s\n", filename)

		path := fmt.Sprintf("%s/%s", *folder, filename)
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

		// if we have a file in the expected folder then check that rather than the other node
		// it could be specific to the new client
		expectedFile, err := os.OpenFile(fmt.Sprintf("%s-expected/%s", *folder, filename), os.O_RDONLY, 0644)
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				fmt.Printf("Error reading expected file: %v\n", err)
				return
			}
		}
		if err == nil {
			fmt.Printf("Found expected file for %s, comparing...\n", filename)
			// the file exists so just compare that
			expectedBytes, err := io.ReadAll(expectedFile)
			if err != nil {
				fmt.Printf("Error reading expected file: %v\n", err)
				return
			}
			markdownOutput += diffTheFiles(res1, expectedBytes, filename, linkedFile)
			writeLinkedFileOutput(res1, expectedBytes, fmt.Sprintf("./output/%s", linkedFile))
			continue
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

		markdownOutput += diffTheFiles(res1, res2, filename, linkedFile)

		// now write the outputs to a file that we can link to in the markdown
		writeLinkedFileOutput(res1, res2, fmt.Sprintf("./output/%s", linkedFile))

		time.Sleep(100 * time.Millisecond)
	}

	_ = os.WriteFile("./output/output.md", []byte(markdownOutput), 0644)
}

func getResponse(host string, contents []byte) ([]byte, error) {
	br := bytes.NewReader(contents)
	req, err := http.NewRequest(http.MethodPost, host, br)
	req.Header.Add("Content-Type", "application/json")
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

func diffTheFiles(res1, res2 []byte, fileName, linkedFile string) string {
	output := ""
	diff, report := jsondiff.Compare(res1, res2, &consoleOptions)
	if diff == jsondiff.FullMatch {
		fmt.Println("Files match")
	} else {
		if consoleOut {
			fmt.Println(report)
		}
		fmt.Println("!!! Files do not match")

		_, report = jsondiff.Compare(res1, res2, &markdownOptions)
		output += fmt.Sprintf("# File: [%s](%s)\n", fileName, linkedFile)
		output += "```json\n"
		output += fmt.Sprintf("%s\n", report)
		output += "```\n\n\n"
	}

	return output
}

func writeLinkedFileOutput(ours, theirs []byte, filename string) {
	out := "# Ours\n"
	out += "```json\n"
	out += fmt.Sprintf("%s\n", ours)
	out += "```\n\n\n"
	out += "# Theirs\n"
	out += "```json\n"
	out += fmt.Sprintf("%s\n", theirs)
	out += "```\n\n\n"
	err := os.WriteFile(filename, []byte(out), 0644)
	if err != nil {
		fmt.Printf("Error writing linked file: %v\n", err)
	}
}

type RpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type PossibleError struct {
	Error RpcError `json:"error"`
}
