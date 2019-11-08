package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"path/filepath"
	"strings"

	"errors"

	"github.com/xeipuuv/gojsonschema"
)

const defaultLogSchema = "logschema.json"

func main() {
	schemaFlag := flag.String("schema", "", "the schema file for parsing log entries")
	debugFlag := flag.Bool("debug", false, "if set, it will output debug information")

	if len(os.Args) == 1 || os.Args[1] == "help" {
		flag.Usage()
		os.Exit(0)
	}

	flag.Parse()

	args := []string{}
	for _, arg := range os.Args[1:] {
		if arg[:2] != "--" {
			args = append(args, arg)
		}
	}

	var scanner *bufio.Scanner
	if len(args) == 0 {
		scanner = bufio.NewScanner(os.Stdin)
	} else {
		file, err := os.Open(args[0])
		if err != nil {
			fmt.Printf("failed to read the source: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()

		scanner = bufio.NewScanner(file)
	}

	schema, err := parseConfig(*schemaFlag)
	if err != nil {
		fmt.Printf("failed to read configuration: %v\n", err)
		os.Exit(1)
	}

	timeout := int64(5) // number of seconds
	tckr := time.NewTicker(time.Duration(timeout) * time.Second)
	defer tckr.Stop()
	lastLineTime := time.Now()

	schemaLoader := gojsonschema.NewStringLoader(schema)

	for {
		select {
		case <-tckr.C:
			if time.Since(lastLineTime).Seconds() > float64(timeout) {
				fmt.Println("No more logs to analyze")
				os.Exit(0)
			}
		default:
			if scanner.Scan() {
				res, err := gojsonschema.Validate(
					schemaLoader,
					gojsonschema.NewStringLoader(scanner.Text()),
				)
				if err != nil {
					fmt.Printf("failed to parse line: %q, err: %v\n", scanner.Text(), err)
					os.Exit(1)
				}

				if !res.Valid() {
					for _, err := range res.Errors() {
						fmt.Printf("failed to validate line: %q, err: %v\n", scanner.Text(), err.String())
					}
					os.Exit(1)
				}
				lastLineTime = time.Now()
				if *debugFlag {
					log.Printf("Successfuly validated:\n%q\n\n", scanner.Text())
				}
			}

			if err := scanner.Err(); err != nil {
				fmt.Printf("failed to read a line: %v\n", err)
				os.Exit(1)
			}
		}
	}
}

func parseConfig(fileName string) (string, error) {
	var err error

	if fileName == "" {
		fallbackFilename, _ := filepath.Abs(defaultLogSchema)
		if _, err = os.Stat(fallbackFilename); os.IsNotExist(err) {
			return "", errors.New("missing config file")
		}
		fileName = fallbackFilename
	} else {
		if !strings.HasPrefix(fileName, "/") {
			fileName, err = filepath.Abs(fileName)
			if err != nil {
				return "", fmt.Errorf("failed to determine config file location: %v", err)
			}
		}

		if _, err = os.Stat(fileName); os.IsNotExist(err) {
			return "", fmt.Errorf("file %s does not exist", fileName)
		}
	}

	f, err := os.Open(fileName)
	if err != nil {
		return "", fmt.Errorf("failed to open the config file: %v", err)
	}

	fContent, err := ioutil.ReadAll(f)
	if err != nil {
		return "", fmt.Errorf("failed to read the config file: %v", err)
	}

	return string(fContent), nil
}
