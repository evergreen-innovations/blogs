package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
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

	errs := make(chan error)

	go func() {
		// Simple long-running process
		for i := 0; i < 10; i++ {
			fmt.Println("in the for loop, iteration", i)
			time.Sleep(time.Second)
		}

		// Indicate normal end to the program
		errs <- nil
	}()

	// Trap any signals to exit gracefully
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("signal trapped: %v", <-c)
	}()

	// Block execution until any errors are encountered.
	// Deferred functions will be run afterwards.
	err = <-errs

}
