package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/evergreen-innovations/blogs/modbus"
)

// Defining register values for the demo
const (
	FrequencyAddr uint16 = 16384
	PhaseV1Addr   uint16 = 16386
	PhaseV2Addr   uint16 = 16388
	PhaseV3Addr   uint16 = 16390
	CurrentI1Addr uint16 = 16402
	CurrentI2Addr uint16 = 16404
	CurrentI3Addr uint16 = 16406
)

// Register stores the name and address of a register
type Register struct {
	Name    string
	Address uint16
}

var registers = []Register{
	{"Frequency", FrequencyAddr},
	{"PhaseV1", PhaseV1Addr},
	{"PhaseV2", PhaseV2Addr},
	{"PhaseV3", PhaseV3Addr},
	{"CurrentI1", CurrentI1Addr},
	{"CurrentI2", CurrentI2Addr},
	{"CurrentI3", CurrentI3Addr},
}

const (
	defaultHost string = "0.0.0.0"
	defaultPort string = ":1503"
)

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

	// Set up the commandline options
	host := flag.String("host", defaultHost, "host for the modbus server")
	port := flag.String("port", defaultPort, "port for the modbus server")
	flag.Parse()

	// Open the modbus server
	addr := fmt.Sprintf("%s%s", *host, *port)
	s, err := modbus.NewServer(addr)
	if err != nil {
		mainErr = fmt.Errorf("creating server: %v", err)
		return
	}
	defer s.Close()

	fmt.Println("Modbus server for power meter running at address", addr)

	// Channel to capture any errors from the go-routines
	// that make up the program.
	errs := make(chan error)

	// Go-routine for writing to the registers
	go func() {
		rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
		ticker := time.NewTicker(500 * time.Millisecond)
		for range ticker.C {
			// Loop over the register address values from map and write the values
			for range ticker.C {
				// Loop over the register address values from map and write the values
				for _, r := range registers {
					value := uint16(rnd.Int())
					fmt.Printf("writing to %v[%v] value: %v\n", r.Name, r.Address, value)
					s.WriteRegister(r.Address, value)
				}
			}
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
