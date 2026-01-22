// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package ssrp

import (
	"encoding/binary"
	"net"
	"strings"
	"sync"
)

// TestInstance represents a SQL Server instance for testing.
type TestInstance struct {
	ServerName   string
	InstanceName string
	IsClustered  bool
	Version      string
	TcpPort      int
	NamedPipe    string
	DacPort      int // DAC port for this instance
}

// TestServer is a mock SSRP server for testing.
type TestServer struct {
	conn      *net.UDPConn
	instances []TestInstance
	mu        sync.Mutex
	running   bool
	done      chan struct{}
}

// NewTestServer creates a new test SSRP server.
func NewTestServer(instances []TestInstance) *TestServer {
	return &TestServer{
		instances: instances,
		done:      make(chan struct{}),
	}
}

// Start begins listening on a random UDP port.
// Returns the address to connect to.
func (s *TestServer) Start() (string, error) {
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		return "", err
	}
	s.conn, err = net.ListenUDP("udp", addr)
	if err != nil {
		return "", err
	}

	s.mu.Lock()
	s.running = true
	s.mu.Unlock()

	go s.serve()

	// Return the address without the port (caller will use port 1434 convention)
	localAddr := s.conn.LocalAddr().(*net.UDPAddr)
	return localAddr.String(), nil
}

// Port returns the port the server is listening on.
func (s *TestServer) Port() int {
	if s.conn == nil {
		return 0
	}
	return s.conn.LocalAddr().(*net.UDPAddr).Port
}

// Stop shuts down the test server.
func (s *TestServer) Stop() {
	s.mu.Lock()
	s.running = false
	s.mu.Unlock()

	if s.conn != nil {
		s.conn.Close()
	}
	close(s.done)
}

func (s *TestServer) serve() {
	buf := make([]byte, 1024)
	for {
		s.mu.Lock()
		running := s.running
		s.mu.Unlock()
		if !running {
			return
		}

		n, remoteAddr, err := s.conn.ReadFromUDP(buf)
		if err != nil {
			continue
		}
		if n < 1 {
			continue
		}

		var resp []byte
		switch buf[0] {
		case 0x02: // CLNT_BCAST_EX
			resp = s.handleBroadcastEx()
		case 0x03: // CLNT_UCAST_EX
			resp = s.handleUnicastEx()
		case 0x04: // CLNT_UCAST_INST
			if n < 2 {
				continue
			}
			instanceName := extractNullTerminatedString(buf[1:n])
			resp = s.handleUnicastInst(instanceName)
		case 0x0F: // CLNT_UCAST_DAC
			if n < 3 {
				continue
			}
			// protocolVersion := buf[1] // Should be 0x01
			instanceName := extractNullTerminatedString(buf[2:n])
			resp = s.handleUnicastDac(instanceName)
		default:
			continue
		}

		if resp != nil {
			s.conn.WriteToUDP(resp, remoteAddr)
		}
	}
}

func extractNullTerminatedString(data []byte) string {
	for i, b := range data {
		if b == 0 {
			return string(data[:i])
		}
	}
	return string(data)
}

// handleBroadcastEx handles CLNT_BCAST_EX (0x02) - same as CLNT_UCAST_EX
func (s *TestServer) handleBroadcastEx() []byte {
	return s.handleUnicastEx()
}

// handleUnicastEx handles CLNT_UCAST_EX (0x03) - returns all instances
func (s *TestServer) handleUnicastEx() []byte {
	var data strings.Builder
	for _, inst := range s.instances {
		data.WriteString(formatInstanceResponse(inst))
	}
	return buildSvrResp([]byte(data.String()))
}

// handleUnicastInst handles CLNT_UCAST_INST (0x04) - returns specific instance
func (s *TestServer) handleUnicastInst(instanceName string) []byte {
	for _, inst := range s.instances {
		if strings.EqualFold(inst.InstanceName, instanceName) {
			return buildSvrResp([]byte(formatInstanceResponse(inst)))
		}
	}
	return nil // No response if instance not found
}

// handleUnicastDac handles CLNT_UCAST_DAC (0x0F) - returns DAC port
func (s *TestServer) handleUnicastDac(instanceName string) []byte {
	for _, inst := range s.instances {
		if strings.EqualFold(inst.InstanceName, instanceName) {
			// SVR_RESP (DAC) format:
			// 1 byte: SVR_RESP (0x05)
			// 2 bytes: RESP_SIZE (0x0006)
			// 1 byte: PROTOCOLVERSION (0x01)
			// 2 bytes: TCP_DAC_PORT (little-endian)
			resp := make([]byte, 6)
			resp[0] = 0x05
			binary.LittleEndian.PutUint16(resp[1:], 0x0006)
			resp[3] = 0x01
			binary.LittleEndian.PutUint16(resp[4:], uint16(inst.DacPort))
			return resp
		}
	}
	return nil
}

// formatInstanceResponse formats a single instance for SVR_RESP RESP_DATA.
// Format: ServerName;VALUE;InstanceName;VALUE;IsClustered;Yes/No;Version;VALUE;tcp;PORT;np;PIPE;;
func formatInstanceResponse(inst TestInstance) string {
	var b strings.Builder
	b.WriteString("ServerName;")
	b.WriteString(inst.ServerName)
	b.WriteString(";InstanceName;")
	b.WriteString(inst.InstanceName)
	b.WriteString(";IsClustered;")
	if inst.IsClustered {
		b.WriteString("Yes")
	} else {
		b.WriteString("No")
	}
	b.WriteString(";Version;")
	b.WriteString(inst.Version)
	if inst.TcpPort > 0 {
		b.WriteString(";tcp;")
		b.WriteString(intToStr(inst.TcpPort))
	}
	if inst.NamedPipe != "" {
		b.WriteString(";np;")
		b.WriteString(inst.NamedPipe)
	}
	b.WriteString(";;")
	return b.String()
}

func intToStr(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}

// buildSvrResp builds an SVR_RESP packet.
// Format: 1 byte (0x05) + 2 bytes length (little-endian) + data
func buildSvrResp(data []byte) []byte {
	resp := make([]byte, 3+len(data))
	resp[0] = 0x05
	binary.LittleEndian.PutUint16(resp[1:], uint16(len(data)))
	copy(resp[3:], data)
	return resp
}
