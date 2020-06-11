package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

func main() {
	var err error

	// Deferred functions run in reverse order so this will be the last
	// one called, after any tidy up.
	defer func() {
		if err != nil {
			fmt.Println("error encountered:", err)
			os.Exit(1)
		} else {
			fmt.Println("exiting")
		}
	}()

	if len(os.Args) != 2 {
		err = fmt.Errorf("invalid number of arguments")
		return
	}

	fileName := os.Args[1]

	f, err := os.Open(fileName)
	if err != nil {
		err = fmt.Errorf("opening file: %v", err)
		return
	}
	// usually this would be just f.Close() but we are putting some logging
	// to show the process
	defer func() {
		f.Close()
		fmt.Println("closed the file")
	}()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		err = fmt.Errorf("reading file: %v", err)
		return
	}

	fmt.Println("File contents:", string(b))
}
