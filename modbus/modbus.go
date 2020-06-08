// Package modbus provides simple client and server functionality over a TCP connection.
package modbus

import (
<<<<<<< HEAD
	"encoding/binary"
	"fmt"
	"log"
	"time"
=======
	"github.com/evergreen-innovations/modbus/internal/conversions"
>>>>>>> fb3fac00a55c6bf3148e97054e9fcc4296e7fb5d

	"github.com/goburrow/modbus"
	"github.com/tbrandon/mbserver"
)

// Server is modbus server
type Server struct {
	s *mbserver.Server
}

// NewServer creates a new modbus server which listens at the given address
func NewServer(addr string) (*Server, error) {
	s := mbserver.NewServer()
	if err := s.ListenTCP(addr); err != nil {
		return nil, err
	}

	return &Server{s: s}, nil
}

// WriteRegister writes a value to the given address
func (s *Server) WriteRegister(address uint16, value uint16) {
	s.s.HoldingRegisters[address] = value
}

// Close closes the server
func (s *Server) Close() {
	s.s.Close()
}

// Client is a modbus client
type Client struct {
	handler *modbus.TCPClientHandler
	client  modbus.Client
}

// NewClient starts a modbus client listening at the given address
func NewClient(addr string) (*Client, error) {
	handler := modbus.NewTCPClientHandler(addr)
	handler.Timeout = 10 * time.Second
	if err := handler.Connect(); err != nil {
		return nil, err
	}
	client := modbus.NewClient(handler)

	return &Client{handler: handler, client: client}, nil
}

// ReadRegister reads from a specified register
func (c *Client) ReadRegister(address uint16) (float32, error) {
	result, err := c.client.ReadHoldingRegisters(address, 1) // read 2 bytes
	if err != nil {
		return 0.0, err
	}

<<<<<<< HEAD
	float := Float32frombytes(result)

	// fmt.Printf("\n Byte Array %v\n", result)
	fmt.Printf("\n Read from register %d (%s), value %f ", address, regname, float)

=======
	return conversions.Float32FromBytes(result), nil
>>>>>>> fb3fac00a55c6bf3148e97054e9fcc4296e7fb5d
}

// Close closes the client
func (c *Client) Close() error {
	return c.handler.Close()
}

//Modbus conversions

//Float32frombytes - convert bytes to float 32 value
func Float32frombytes(bytes []byte) float32 {
	bits := binary.BigEndian.Uint16(bytes)
	return float32(bits)
}
