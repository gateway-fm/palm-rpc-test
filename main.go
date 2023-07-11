package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/nsf/jsondiff"
)

var (
	client = &http.Client{}
)

func main() {
	host1 := flag.String("host1", "http://127.0.0.1:8545", "address of first RPC host")
	host2 := flag.String("host2", "http://127.0.0.1:8546", "address of second RPC host")
	flag.Parse()

	inputFiles, err := os.ReadDir("./input")
	if err != nil {
		fmt.Printf("Error reading input directory: %v\n", err)
		return
	}

	compareOptions := jsondiff.DefaultConsoleOptions()

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

		diff, report := jsondiff.Compare(res1, res2, &compareOptions)
		if diff == jsondiff.FullMatch {
			fmt.Println("Files match")
		} else {
			fmt.Println("!! Files do not match !!")
			fmt.Println(report)
		}
	}
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
