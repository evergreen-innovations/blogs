package main

import (
	"bufio"
	"errors"
	"flag"
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
			flag.Usage()
			os.Exit(1)
		}
	}()

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s  [options] pattern [file] \n", os.Args[0])
		flag.PrintDefaults()
	}

	var insensitiveFlag bool
	flag.BoolVar(&insensitiveFlag, "i", false, "insensitive match")
	flag.Parse()

	args := flag.Args()

	if len(args) == 0 {
		mainErr = errors.New("pattern must be supplied")
		return
	}

	pat := args[0]

	if insensitiveFlag {
		pat = "(?i)" + pat
	}

	rex, err := regexp.Compile(pat)
	if err != nil {
		mainErr = err
		return
	}

	f := os.Stdin
	if len(args) == 2 {
		var err error
		f, err = os.Open(args[1])
		if err != nil {
			mainErr = err
			return
		}
		defer f.Close()
	}

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
