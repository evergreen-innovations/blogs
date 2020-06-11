package main

import (
	"flag"
	"fmt"
)

func main() {
	withError := flag.Bool("error", false, "exit program with error")
	flag.Parse()

	defer func() {
		fmt.Println("in the deferred function")
	}()

	fmt.Println("in main")

	if *withError {
		panic("Exiting with error")
	}
}
