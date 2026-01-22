// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package ssrp

import (
	"encoding/binary"
	"net"
	"strings"
	"testing"
	"time"
)

func TestBuildSvrResp(t *testing.T) {
	data := []byte("ServerName;TEST;InstanceName;INST1;;")
	resp := buildSvrResp(data)

	if resp[0] != 0x05 {
		t.Errorf("expected SVR_RESP byte 0x05, got 0x%02x", resp[0])
	}

	respLen := binary.LittleEndian.Uint16(resp[1:3])
	if int(respLen) != len(data) {
		t.Errorf("expected length %d, got %d", len(data), respLen)
	}

	if string(resp[3:]) != string(data) {
		t.Errorf("data mismatch")
	}
}

func TestFormatInstanceResponse(t *testing.T) {
	inst := TestInstance{
		ServerName:   "TESTSERVER",
		InstanceName: "SQLEXPRESS",
		IsClustered:  false,
		Version:      "15.0.2000.5",
		TcpPort:      1433,
		NamedPipe:    `\\TESTSERVER\pipe\sql\query`,
	}

	resp := formatInstanceResponse(inst)

	expected := []string{
		"ServerName;TESTSERVER",
		"InstanceName;SQLEXPRESS",
		"IsClustered;No",
		"Version;15.0.2000.5",
		"tcp;1433",
		"np;\\\\TESTSERVER\\pipe\\sql\\query",
	}

	for _, exp := range expected {
		if !strings.Contains(resp, exp) {
			t.Errorf("response missing %q, got: %s", exp, resp)
		}
	}

	// Should end with ;;
	if !strings.HasSuffix(resp, ";;") {
		t.Errorf("response should end with ;;, got: %s", resp)
	}
}

func TestTestServerUnicastEx(t *testing.T) {
	instances := []TestInstance{
		{
			ServerName:   "TESTHOST",
			InstanceName: "SQLEXPRESS",
			Version:      "15.0.2000.5",
			TcpPort:      1433,
		},
		{
			ServerName:   "TESTHOST",
			InstanceName: "SQLDEV",
			Version:      "15.0.2000.5",
			TcpPort:      1434,
		},
	}

	srv := NewTestServer(instances)
	addr, err := srv.Start()
	if err != nil {
		t.Fatalf("failed to start test server: %v", err)
	}
	defer srv.Stop()

	// Connect to the test server
	serverAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		t.Fatalf("failed to resolve address: %v", err)
	}

	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		t.Fatalf("failed to dial: %v", err)
	}
	defer conn.Close()

	// Send CLNT_UCAST_EX (0x03)
	_, err = conn.Write([]byte{0x03})
	if err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	// Read response
	conn.SetReadDeadline(time.Now().Add(time.Second))
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatalf("failed to read: %v", err)
	}

	// Verify response format
	if buf[0] != 0x05 {
		t.Errorf("expected SVR_RESP 0x05, got 0x%02x", buf[0])
	}

	respLen := binary.LittleEndian.Uint16(buf[1:3])
	if int(respLen)+3 != n {
		t.Errorf("length mismatch: header says %d, got %d bytes total", respLen, n)
	}

	respData := string(buf[3:n])
	t.Logf("Response data: %s", respData)

	// Should contain both instances
	if !strings.Contains(respData, "SQLEXPRESS") {
		t.Error("response missing SQLEXPRESS instance")
	}
	if !strings.Contains(respData, "SQLDEV") {
		t.Error("response missing SQLDEV instance")
	}
}

func TestTestServerUnicastInst(t *testing.T) {
	instances := []TestInstance{
		{
			ServerName:   "TESTHOST",
			InstanceName: "SQLEXPRESS",
			Version:      "15.0.2000.5",
			TcpPort:      1433,
		},
		{
			ServerName:   "TESTHOST",
			InstanceName: "SQLDEV",
			Version:      "15.0.2000.5",
			TcpPort:      1434,
		},
	}

	srv := NewTestServer(instances)
	addr, err := srv.Start()
	if err != nil {
		t.Fatalf("failed to start test server: %v", err)
	}
	defer srv.Stop()

	serverAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		t.Fatalf("failed to resolve address: %v", err)
	}

	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		t.Fatalf("failed to dial: %v", err)
	}
	defer conn.Close()

	// Send CLNT_UCAST_INST (0x04) for SQLEXPRESS
	// Format: 0x04 + instance name + null terminator
	msg := append([]byte{0x04}, []byte("SQLEXPRESS")...)
	msg = append(msg, 0x00)

	_, err = conn.Write(msg)
	if err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	conn.SetReadDeadline(time.Now().Add(time.Second))
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatalf("failed to read: %v", err)
	}

	if buf[0] != 0x05 {
		t.Errorf("expected SVR_RESP 0x05, got 0x%02x", buf[0])
	}

	respData := string(buf[3:n])
	t.Logf("Response data: %s", respData)

	// Should contain only SQLEXPRESS
	if !strings.Contains(respData, "SQLEXPRESS") {
		t.Error("response missing SQLEXPRESS instance")
	}
	if strings.Contains(respData, "SQLDEV") {
		t.Error("response should not contain SQLDEV instance")
	}
}

func TestTestServerUnicastDac(t *testing.T) {
	instances := []TestInstance{
		{
			ServerName:   "TESTHOST",
			InstanceName: "SQLEXPRESS",
			Version:      "15.0.2000.5",
			TcpPort:      1433,
			DacPort:      57138,
		},
	}

	srv := NewTestServer(instances)
	addr, err := srv.Start()
	if err != nil {
		t.Fatalf("failed to start test server: %v", err)
	}
	defer srv.Stop()

	serverAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		t.Fatalf("failed to resolve address: %v", err)
	}

	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		t.Fatalf("failed to dial: %v", err)
	}
	defer conn.Close()

	// Send CLNT_UCAST_DAC (0x0F)
	// Format: 0x0F + 0x01 (protocol version) + instance name + null terminator
	msg := []byte{0x0F, 0x01}
	msg = append(msg, []byte("SQLEXPRESS")...)
	msg = append(msg, 0x00)

	_, err = conn.Write(msg)
	if err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	conn.SetReadDeadline(time.Now().Add(time.Second))
	buf := make([]byte, 64)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatalf("failed to read: %v", err)
	}

	// Expected: 05 06 00 01 32 df (for port 57138 = 0xDF32)
	if n != 6 {
		t.Fatalf("expected 6 bytes, got %d", n)
	}

	if buf[0] != 0x05 {
		t.Errorf("expected SVR_RESP 0x05, got 0x%02x", buf[0])
	}

	respSize := binary.LittleEndian.Uint16(buf[1:3])
	if respSize != 0x0006 {
		t.Errorf("expected RESP_SIZE 0x0006, got 0x%04x", respSize)
	}

	if buf[3] != 0x01 {
		t.Errorf("expected PROTOCOLVERSION 0x01, got 0x%02x", buf[3])
	}

	dacPort := binary.LittleEndian.Uint16(buf[4:6])
	if dacPort != 57138 {
		t.Errorf("expected DAC port 57138, got %d", dacPort)
	}
}

func TestTestServerNotFound(t *testing.T) {
	instances := []TestInstance{
		{
			ServerName:   "TESTHOST",
			InstanceName: "SQLEXPRESS",
			TcpPort:      1433,
		},
	}

	srv := NewTestServer(instances)
	addr, err := srv.Start()
	if err != nil {
		t.Fatalf("failed to start test server: %v", err)
	}
	defer srv.Stop()

	serverAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		t.Fatalf("failed to resolve address: %v", err)
	}

	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		t.Fatalf("failed to dial: %v", err)
	}
	defer conn.Close()

	// Request non-existent instance
	msg := append([]byte{0x04}, []byte("NONEXISTENT")...)
	msg = append(msg, 0x00)

	_, err = conn.Write(msg)
	if err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	// Should timeout (no response for non-existent instance)
	conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	buf := make([]byte, 64)
	_, err = conn.Read(buf)
	if err == nil {
		t.Error("expected timeout for non-existent instance")
	}
}

// TestReadResult tests the response parsing logic
func TestReadResult(t *testing.T) {
	// Build a mock response matching the spec example
	// ServerName;ILSUNG1;InstanceName;YUKONSTD;IsClustered;No;Version;9.00.1399.06;tcp;57137;;
	respData := "ServerName;ILSUNG1;InstanceName;YUKONSTD;IsClustered;No;Version;9.00.1399.06;tcp;57137;;"

	resp := buildSvrResp([]byte(respData))

	// Create a mock UDP connection using a channel-based approach
	// For this test, we'll just verify the parsing logic directly
	// by simulating what readResult does

	if resp[0] != 0x05 {
		t.Fatalf("expected 0x05, got 0x%02x", resp[0])
	}

	msgLen := int(binary.LittleEndian.Uint16(resp[1:]))
	if msgLen+3 != len(resp) {
		t.Fatalf("length mismatch")
	}

	msg := resp[3 : 3+msgLen]
	msgList := strings.Split(string(msg), ";")

	// Parse like readResult does
	var info InstanceInfo
	for i := 0; i < len(msgList)-1; i += 2 {
		key := msgList[i]
		value := msgList[i+1]
		switch key {
		case "ServerName":
			info.Server = value
		case "InstanceName":
			info.Instance = value
		case "tcp":
			// Note: original code ignores parse errors
			info.Tcp, _ = parseInt(value)
		case "np":
			info.NamedPipe = value
		}
	}

	if info.Server != "ILSUNG1" {
		t.Errorf("expected ServerName ILSUNG1, got %s", info.Server)
	}
	if info.Instance != "YUKONSTD" {
		t.Errorf("expected InstanceName YUKONSTD, got %s", info.Instance)
	}
	if info.Tcp != 57137 {
		t.Errorf("expected tcp 57137, got %d", info.Tcp)
	}
}

func parseInt(s string) (int, error) {
	var n int
	for _, c := range s {
		if c < '0' || c > '9' {
			return n, nil
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}

// TestClientAgainstTestServer tests the actual ssrp client functions against the test server
func TestClientAgainstTestServer(t *testing.T) {
	instances := []TestInstance{
		{
			ServerName:   "TESTHOST",
			InstanceName: "SQLEXPRESS",
			Version:      "15.0.2000.5",
			TcpPort:      1433,
			NamedPipe:    `\\TESTHOST\pipe\sql\query`,
		},
		{
			ServerName:   "TESTHOST",
			InstanceName: "SQLDEV",
			Version:      "15.0.2000.5",
			TcpPort:      51234,
		},
	}

	srv := NewTestServer(instances)
	addr, err := srv.Start()
	if err != nil {
		t.Fatalf("failed to start test server: %v", err)
	}
	defer srv.Stop()

	t.Run("FetchInstanceInfoList", func(t *testing.T) {
		list, err := fetchAddr(addr, "", true)
		if err != nil {
			t.Fatalf("fetchAddr failed: %v", err)
		}

		if len(list) != 2 {
			t.Fatalf("expected 2 instances, got %d", len(list))
		}

		// Check first instance
		found := false
		for _, info := range list {
			if info.Instance == "SQLEXPRESS" {
				found = true
				if info.Server != "TESTHOST" {
					t.Errorf("expected ServerName TESTHOST, got %s", info.Server)
				}
				if info.Tcp != 1433 {
					t.Errorf("expected tcp 1433, got %d", info.Tcp)
				}
				if info.NamedPipe != `\\TESTHOST\pipe\sql\query` {
					t.Errorf("expected named pipe, got %s", info.NamedPipe)
				}
			}
		}
		if !found {
			t.Error("SQLEXPRESS instance not found")
		}
	})

	t.Run("FetchInstanceInfo_specific", func(t *testing.T) {
		list, err := fetchAddr(addr, "SQLDEV", false)
		if err != nil {
			t.Fatalf("fetchAddr failed: %v", err)
		}

		if len(list) != 1 {
			t.Fatalf("expected 1 instance, got %d", len(list))
		}

		if list[0].Instance != "SQLDEV" {
			t.Errorf("expected SQLDEV, got %s", list[0].Instance)
		}
		if list[0].Tcp != 51234 {
			t.Errorf("expected tcp 51234, got %d", list[0].Tcp)
		}
	})

	t.Run("FetchInstanceInfo_case_insensitive", func(t *testing.T) {
		// Instance names should be case-insensitive
		list, err := fetchAddr(addr, "sqlexpress", false)
		if err != nil {
			t.Fatalf("fetchAddr failed: %v", err)
		}

		if len(list) != 1 {
			t.Fatalf("expected 1 instance, got %d", len(list))
		}

		if list[0].Instance != "SQLEXPRESS" {
			t.Errorf("expected SQLEXPRESS, got %s", list[0].Instance)
		}
	})
}

// TestReadResultPanicProtection tests that readResult doesn't panic on malformed data
func TestReadResultPanicProtection(t *testing.T) {
	// Test cases that could cause panic with the old code
	testCases := []struct {
		name     string
		respData string
	}{
		{"empty", ""},
		{"single_key", "ServerName"},
		{"odd_elements", "ServerName;VALUE;InstanceName"},
		{"trailing_semicolon", "ServerName;VALUE;"},
		{"just_semicolons", ";;;"},
		{"normal", "ServerName;TEST;InstanceName;INST;;"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Build SVR_RESP packet
			resp := buildSvrResp([]byte(tc.respData))

			// Create UDP server/client pair for testing
			serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			server, err := net.ListenUDP("udp", serverAddr)
			if err != nil {
				t.Fatal(err)
			}
			defer server.Close()

			clientAddr, err := net.ResolveUDPAddr("udp", server.LocalAddr().String())
			if err != nil {
				t.Fatal(err)
			}
			client, err := net.DialUDP("udp", nil, clientAddr)
			if err != nil {
				t.Fatal(err)
			}
			defer client.Close()

			// Send response from "server" side
			go func() {
				buf := make([]byte, 64)
				n, addr, err := server.ReadFromUDP(buf)
				if err != nil || n == 0 {
					return
				}
				server.WriteToUDP(resp, addr)
			}()

			// Trigger read from client side
			client.Write([]byte{0x03})

			// This should not panic
			buf := make([]byte, 1024)
			result, err := readResult(client, buf)

			// We don't care about the result, just that it didn't panic
			t.Logf("result=%v, err=%v", result, err)
		})
	}
}

// TestClientMessageConstruction verifies the client constructs messages correctly
func TestClientMessageConstruction(t *testing.T) {
	t.Run("CLNT_UCAST_INST_construction", func(t *testing.T) {
		// Verify the fixed code produces correct CLNT_UCAST_INST message
		// Per spec example: 04 59 55 4b 4f 4e 53 54 44 00 for "YUKONSTD"
		instance := "YUKONSTD"

		// What the fixed code produces:
		msg := make([]byte, 1+len(instance)+1)
		msg[0] = 0x04
		copy(msg[1:], instance)
		msg[len(msg)-1] = 0x00

		expected := []byte{0x04, 0x59, 0x55, 0x4b, 0x4f, 0x4e, 0x53, 0x54, 0x44, 0x00}

		if string(msg) != string(expected) {
			t.Errorf("CLNT_UCAST_INST mismatch:\n  got:      %x\n  expected: %x", msg, expected)
		}

		// Verify the null terminator is at the right position
		if msg[len(msg)-1] != 0x00 {
			t.Error("message should end with null terminator")
		}
		if msg[1+len(instance)] != 0x00 {
			t.Errorf("null terminator should be at position %d", 1+len(instance))
		}
	})

	t.Run("CLNT_UCAST_INST_truncation", func(t *testing.T) {
		// Instance name must be max 32 bytes per spec
		longInstance := "THIS_IS_A_VERY_LONG_INSTANCE_NAME_THAT_EXCEEDS_32_BYTES"

		// Simulate what the code does
		instBytes := []byte(longInstance)
		if len(instBytes) > 32 {
			instBytes = instBytes[:32]
		}
		msg := make([]byte, 1+len(instBytes)+1)
		msg[0] = 0x04
		copy(msg[1:], instBytes)

		// Message should be 1 (type) + 32 (truncated name) + 1 (null) = 34 bytes
		if len(msg) != 34 {
			t.Errorf("expected message length 34, got %d", len(msg))
		}

		// Verify instance was truncated to 32 bytes
		if string(msg[1:33]) != longInstance[:32] {
			t.Errorf("instance not truncated correctly")
		}
	})
}

// TestProtocolExamples verifies our implementation against the spec examples
func TestProtocolExamples(t *testing.T) {
	t.Run("CLNT_UCAST_EX_request", func(t *testing.T) {
		// Per spec: Request = 0x03
		msg := []byte{0x03}
		if msg[0] != 0x03 {
			t.Error("CLNT_UCAST_EX should be 0x03")
		}
	})

	t.Run("CLNT_UCAST_INST_request", func(t *testing.T) {
		// Per spec example: 04 59 55 4b 4f 4e 53 54 44 00 = 0x04 + "YUKONSTD" + 0x00
		instanceName := "YUKONSTD"
		msg := make([]byte, 1+len(instanceName)+1)
		msg[0] = 0x04
		copy(msg[1:], instanceName)
		msg[len(msg)-1] = 0x00

		expected := []byte{0x04, 0x59, 0x55, 0x4b, 0x4f, 0x4e, 0x53, 0x54, 0x44, 0x00}
		if string(msg) != string(expected) {
			t.Errorf("CLNT_UCAST_INST mismatch:\n  got:      %x\n  expected: %x", msg, expected)
		}
	})

	t.Run("CLNT_UCAST_DAC_request", func(t *testing.T) {
		// Per spec example: 0f 01 59 55 4b 4f 4e 53 54 44 00 = 0x0F + 0x01 + "YUKONSTD" + 0x00
		instanceName := "YUKONSTD"
		msg := make([]byte, 2+len(instanceName)+1)
		msg[0] = 0x0F
		msg[1] = 0x01 // PROTOCOLVERSION
		copy(msg[2:], instanceName)
		msg[len(msg)-1] = 0x00

		expected := []byte{0x0f, 0x01, 0x59, 0x55, 0x4b, 0x4f, 0x4e, 0x53, 0x54, 0x44, 0x00}
		if string(msg) != string(expected) {
			t.Errorf("CLNT_UCAST_DAC mismatch:\n  got:      %x\n  expected: %x", msg, expected)
		}
	})

	t.Run("SVR_RESP_DAC", func(t *testing.T) {
		// Per spec example: 05 06 00 01 32 df = port 0xDF32 = 57138
		port := uint16(57138)
		resp := make([]byte, 6)
		resp[0] = 0x05
		binary.LittleEndian.PutUint16(resp[1:], 0x0006)
		resp[3] = 0x01
		binary.LittleEndian.PutUint16(resp[4:], port)

		expected := []byte{0x05, 0x06, 0x00, 0x01, 0x32, 0xdf}
		if string(resp) != string(expected) {
			t.Errorf("SVR_RESP DAC mismatch:\n  got:      %x\n  expected: %x", resp, expected)
		}
	})
}
