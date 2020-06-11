package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	withError := flag.Bool("error", false, "exit program with error")
	flag.Parse()

	defer func() {
		fmt.Println("in the deferred function")
	}()

	fmt.Println("in main")

	if *withError {
		fmt.Println("exiting with error")
		os.Exit(1)
	}
}
