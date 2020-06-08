// Package modbus provides simple client and server functionality over a TCP connection.
package modbus

import (

	"github.com/evergreen-innovations/blogs/modbus/internal/conversions"

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

	return conversions.Float32FromBytes(result), nil
}

// Close closes the client
func (c *Client) Close() error {
	return c.handler.Close()
}
