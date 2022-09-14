package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"
)

func main() {
	// Deferred functions called in FILO order meaning
	// this will be the last function to run when we return.
	var mainErr error
	defer func() {
		if mainErr != nil {
			fmt.Printf("%s: %s\n", os.Args[0], mainErr)
			os.Exit(1)
		}
	}()

	args := os.Args[1:]

	if len(args) == 0 {
		mainErr = errors.New("pattern must be supplied")
		return
	}

	pat := args[0]
	rex, err := regexp.Compile(pat)
	if err != nil {
		mainErr = err
		return
	}

	if len(args) != 2 {
		mainErr = errors.New("file must be supplied")
		return
	}

	f, err := os.Open(args[1])
	if err != nil {
		mainErr = err
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		l := scanner.Text()

		if rex.MatchString(l) {
			fmt.Println(l)
		}
	}

	if err := scanner.Err(); err != nil {
		mainErr = err
		return
	}
}
