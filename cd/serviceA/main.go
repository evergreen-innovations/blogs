package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	serverUrl string = "http://localhost:9000/post"
)

// Service struct
type Service struct {
	ServiceName string `json:"serviceName"`
	Value       int    `json:"value"`
}

func main() {
	var mainErr error

	// Deferred functions run in reverse order so this will be the last
	// one called, after any tidy up.
	defer func() {
		if mainErr != nil {
			log.Println("error encountered:", mainErr)
			os.Exit(1)
		} else {
			log.Println("exiting")
		}
	}()

	flag.Parse()

	errs := make(chan error)

	// Go-routine to send mock values to Server B
	go func() {
		rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
		ticker := time.NewTicker(500 * time.Millisecond)
		for range ticker.C {
			// Generate random integer value between 0 and 10.
			value := rnd.Intn(10)

			// converts it into a string
			body := &Service{
				ServiceName: "serviceA",
				Value:       value,
			}
			payloadBuf := new(bytes.Buffer)
			err := json.NewEncoder(payloadBuf).Encode(body)
			if err != nil {
				errs <- fmt.Errorf("error encoding json body: %v", err)
				return
			}

			// Prints the integer value generated
			fmt.Printf("sending value %v\n", body)

			// Sends the post request the url specified
			req, err := http.NewRequest("POST", serverUrl, payloadBuf)
			if err != nil {
				errs <- fmt.Errorf("opening file: %v", err)
				return
			}
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				errs <- fmt.Errorf("error connecting to http client: %v", err)
				return
			}

			fmt.Println("response Status:", resp.Status)
			fmt.Println("response Headers:", resp.Header)
			respBody, _ := ioutil.ReadAll(resp.Body)
			fmt.Println("response Body:", string(respBody))

		}

		errs <- fmt.Errorf("ticker loop closed")
	}()

	// Trap any signals to exit gracefully
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("signal trapped: %v", <-c)
	}()

	// Block execution until any errors are encountered.
	// Deferred functions will be run afterwards.
	mainErr = <-errs
}
