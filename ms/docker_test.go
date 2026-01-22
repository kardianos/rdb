//go:build linux && amd64

// This test file contains Docker-based integration tests for the MS SQL Server driver.
// It requires Docker to be available and will pull/start an MSSQL container.
// Tests include TLS connections and comprehensive data type coverage.

package ms

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/kardianos/rdb"
	"github.com/kardianos/rdb/ms/internal/testcert"
	"github.com/kardianos/rdb/must"
)

const (
	dockerImage   = "mcr.microsoft.com/mssql/server:2025-latest"
	containerName = "rdb-mssql-test"
	saPassword    = "TestP@ssw0rd!"
	mssqlPort     = 11433 // Non-standard port to avoid conflicts
)

// dockerTestEnv holds the Docker test environment state.
type dockerTestEnv struct {
	containerID string
	host        string
	port        int
	db          must.ConnPool
	dbTLS       must.ConnPool
	skipCleanup bool
}

var (
	dockerEnv       *dockerTestEnv
	dockerEnvOnce   sync.Once
	dockerMu        sync.Mutex
	dockerChecked   bool
	dockerAvailable bool
)

// checkDockerAvailable verifies Docker is installed and running.
func checkDockerAvailable(t *testing.T) bool {
	t.Helper()
	if testing.Short() {
		t.Skip("short: docker test")
	}
	dockerMu.Lock()
	defer dockerMu.Unlock()

	if dockerChecked {
		return dockerAvailable
	}
	defer func() {
		dockerChecked = true
	}()

	if runtime.GOOS != "linux" || runtime.GOARCH != "amd64" {
		t.Skip("Docker tests only run on linux/amd64")
		return false
	}

	cmd := exec.Command("docker", "version")
	if err := cmd.Run(); err != nil {
		t.Skipf("Docker not available: %v", err)
		return false
	}

	dockerAvailable = true
	return true
}

// setupDockerEnv creates the Docker test environment.
// The container is started once and reused across all docker tests.
// Cleanup is handled by TestMain via dockerCleanupFunc.
func setupDockerEnv(t *testing.T) *dockerTestEnv {
	t.Helper()
	if testing.Short() {
		t.Skip("short: docker test")
	}

	if !checkDockerAvailable(t) {
		return nil
	}

	var dockerEnvErr error
	dockerEnvOnce.Do(func() {
		dockerEnvErr = initDockerEnv(t)
		if dockerEnvErr == nil {
			// Register cleanup to run when TestMain exits.
			dockerCleanupFunc = func() {
				cleanupDockerEnv()
			}
		}
	})

	if dockerEnvErr != nil {
		t.Fatalf("docker env init: %v", dockerEnvErr)
	}

	return dockerEnv
}

// initDockerEnv initializes the Docker test environment (called once via sync.Once).
func initDockerEnv(t *testing.T) error {
	env := &dockerTestEnv{
		host: "127.0.0.1",
		port: mssqlPort,
	}

	// Stop and remove any existing container
	exec.Command("docker", "stop", containerName).Run()
	exec.Command("docker", "rm", containerName).Run()

	// Pull the image
	fmt.Println("Docker: Pulling MSSQL image...")
	pullCmd := exec.Command("docker", "pull", dockerImage)
	pullCmd.Stdout = os.Stdout
	pullCmd.Stderr = os.Stderr
	if err := pullCmd.Run(); err != nil {
		return fmt.Errorf("docker pull: %w", err)
	}

	// Start the container without custom TLS config
	// SQL Server will generate its own self-signed certificate
	fmt.Println("Docker: Starting MSSQL container...")
	runArgs := []string{
		"run", "-d",
		"--name", containerName,
		"-e", "ACCEPT_EULA=Y",
		"-e", "MSSQL_SA_PASSWORD=" + saPassword,
		"-e", "MSSQL_PID=Developer",
		"-p", fmt.Sprintf("%d:1433", mssqlPort),
		dockerImage,
	}

	var stdout, stderr bytes.Buffer
	runCmd := exec.Command("docker", runArgs...)
	runCmd.Stdout = &stdout
	runCmd.Stderr = &stderr
	if err := runCmd.Run(); err != nil {
		return fmt.Errorf("docker run: %v\nstderr: %s", err, stderr.String())
	}
	env.containerID = strings.TrimSpace(stdout.String())
	fmt.Printf("Docker: Container started: %s\n", env.containerID[:12])

	// Wait for SQL Server to be ready by monitoring logs
	fmt.Println("Docker: Waiting for SQL Server to start...")
	if err := waitForMSSQLReady(env.containerID, 90*time.Second); err != nil {
		cleanupDockerEnv()
		return fmt.Errorf("wait for MSSQL: %w", err)
	}

	// Create connection pool
	// TLS is used for login encryption even in non-secure mode,
	// so we need InsecureSkipVerify for self-signed certs
	config := &rdb.Config{
		DriverName:         "ms",
		Hostname:           env.host,
		Port:               env.port,
		Username:           "sa",
		Password:           saPassword,
		Database:           "master",
		PoolInitCapacity:   1,
		PoolMaxCapacity:    5,
		DialTimeout:        5 * time.Second,
		InsecureSkipVerify: true,                      // Accept SQL Server's self-signed cert for login
		ResetQuery:         "SET TEXTSIZE 2147483647", // Prevent truncation of large varchar(max) data
	}

	env.db = must.Open(config)

	// Test the connection
	ctx := context.Background()
	if err := env.db.Normal().Ping(ctx); err != nil {
		cleanupDockerEnv()
		return fmt.Errorf("ping: %w", err)
	}

	// Create TLS connection pool using SQL Server's self-signed certificate
	tlsConfig := &rdb.Config{
		DriverName:         "ms",
		Hostname:           env.host,
		Port:               env.port,
		Username:           "sa",
		Password:           saPassword,
		Database:           "master",
		PoolInitCapacity:   1,
		PoolMaxCapacity:    5,
		DialTimeout:        5 * time.Second,
		Secure:             true,
		InsecureSkipVerify: true, // Accept SQL Server's self-signed cert
	}

	// TLS connection - may fail if not configured properly
	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Docker: TLS connection failed (may not be configured): %v\n", r)
			}
		}()
		env.dbTLS = must.Open(tlsConfig)
	}()

	dockerEnv = env
	return nil
}

// cleanupDockerEnv stops and removes the Docker container.
func cleanupDockerEnv() {
	if dockerEnv == nil {
		return
	}

	if dockerEnv.db.Valid() {
		dockerEnv.db.Close()
	}
	if dockerEnv.dbTLS.Valid() {
		dockerEnv.dbTLS.Close()
	}

	if dockerEnv.containerID != "" && !dockerEnv.skipCleanup {
		fmt.Println("Docker: Stopping container...")
		exec.Command("docker", "stop", containerName).Run()
		exec.Command("docker", "rm", containerName).Run()
	}

	dockerEnv = nil
}

// waitForMSSQLReady polls docker logs for the SQL Server ready message.
func waitForMSSQLReady(containerID string, timeout time.Duration) error {
	// SQL Server prints this message when ready
	readyMsg := "SQL Server is now ready for client connections"

	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		// Get current container logs
		cmd := exec.Command("docker", "logs", containerID)
		output, err := cmd.CombinedOutput()
		if err != nil {
			time.Sleep(time.Second)
			continue
		}
		os.Stdout.Write(output)

		// Check if ready message is in the output
		if bytes.Contains(output, []byte(readyMsg)) {
			// Give SQL Server a moment after the message
			time.Sleep(time.Millisecond * 200)
			return nil
		}

		time.Sleep(time.Second)
	}

	return fmt.Errorf("timeout waiting for SQL Server ready message")
}

// TestDockerDecimalPrecision tests various decimal precisions.
func TestDockerDecimalPrecision(t *testing.T) {
	env := setupDockerEnv(t)
	if env == nil {
		return
	}

	// Table-driven tests for decimal precision
	// Per TDS spec, valid lengths are 0x05, 0x09, 0x0D, 0x11 (5,9,13,17 bytes)
	// which correspond to precision ranges:
	// 1-9: 5 bytes, 10-19: 9 bytes, 20-28: 13 bytes, 29-38: 17 bytes
	tests := []struct {
		name      string
		precision int
		scale     int
		value     string
		wantStr   string
	}{
		// Precision 1-9 (5 byte storage)
		{name: "p1_s0", precision: 1, scale: 0, value: "9", wantStr: "9"},
		{name: "p5_s2", precision: 5, scale: 2, value: "123.45", wantStr: "123.45"},
		{name: "p9_s4", precision: 9, scale: 4, value: "12345.6789", wantStr: "12345.6789"},

		// Precision 10-19 (9 byte storage)
		{name: "p10_s0", precision: 10, scale: 0, value: "1234567890", wantStr: "1234567890"},
		{name: "p15_s5", precision: 15, scale: 5, value: "1234567890.12345", wantStr: "1234567890.12345"},
		{name: "p19_s4", precision: 19, scale: 4, value: "123456789012345.6789", wantStr: "123456789012345.6789"},

		// Precision 20-28 (13 byte storage)
		{name: "p20_s0", precision: 20, scale: 0, value: "12345678901234567890", wantStr: "12345678901234567890"},
		{name: "p25_s5", precision: 25, scale: 5, value: "12345678901234567890.12345", wantStr: "12345678901234567890.12345"},
		{name: "p28_s4", precision: 28, scale: 4, value: "123456789012345678901234.5678", wantStr: "123456789012345678901234.5678"},

		// Precision 29-38 (17 byte storage)
		{name: "p29_s0", precision: 29, scale: 0, value: "12345678901234567890123456789", wantStr: "12345678901234567890123456789"},
		{name: "p35_s5", precision: 35, scale: 5, value: "123456789012345678901234567890.12345", wantStr: "123456789012345678901234567890.12345"},
		{name: "p38_s6", precision: 38, scale: 6, value: "12345678901234567890123456789012.345678", wantStr: "12345678901234567890123456789012.345678"},

		// Edge cases
		{name: "negative", precision: 10, scale: 2, value: "-12345678.90", wantStr: "-12345678.90"},
		{name: "zero", precision: 10, scale: 2, value: "0.00", wantStr: "0.00"},
		{name: "small_frac", precision: 10, scale: 8, value: "1.23456789", wantStr: "1.23456789"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Parse the value
			rat := new(big.Rat)
			if _, ok := rat.SetString(tc.value); !ok {
				t.Fatalf("invalid test value: %s", tc.value)
			}

			cmd := &rdb.Command{
				SQL: fmt.Sprintf(`
					DECLARE @v decimal(%d,%d) = @input;
					SELECT v = @v, s = CONVERT(varchar(100), @v);
				`, tc.precision, tc.scale),
				Arity: rdb.OneMust,
			}

			ctx := context.Background()
			res := env.db.Query(ctx, cmd,
				rdb.Param{Name: "input", Type: rdb.TypeDecimal, Precision: tc.precision, Scale: tc.scale, Value: rat},
			)
			defer res.Close()

			res.Scan()
			gotRat := res.Get("v")
			gotStr := res.Get("s")

			// Compare string representation
			if s, ok := gotStr.(string); ok {
				if s != tc.wantStr {
					t.Errorf("string mismatch: got %q, want %q", s, tc.wantStr)
				}
			} else if b, ok := gotStr.([]byte); ok {
				if string(b) != tc.wantStr {
					t.Errorf("string mismatch: got %q, want %q", string(b), tc.wantStr)
				}
			}

			// Verify the rat value round-trips correctly
			// Parse expected value fresh (the input rat may be modified by encoder)
			if r, ok := gotRat.(*big.Rat); ok {
				expectedRat := new(big.Rat)
				expectedRat.SetString(tc.value)
				gotStr := r.FloatString(tc.scale)
				wantStr := expectedRat.FloatString(tc.scale)
				if gotStr != wantStr {
					t.Errorf("rat value mismatch: got %v, want %v", gotStr, wantStr)
				}
			} else {
				t.Errorf("unexpected type for decimal: %T", gotRat)
			}
		})
	}
}

// TestDockerTLSConnection tests TLS encrypted connections.
func TestDockerTLSConnection(t *testing.T) {
	env := setupDockerEnv(t)
	if env == nil {
		return
	}

	if !env.dbTLS.Valid() {
		t.Skip("TLS connection not available")
	}

	// Test basic query over TLS
	cmd := &rdb.Command{
		SQL:   `SELECT encrypted = encrypt_option FROM sys.dm_exec_connections WHERE session_id = @@SPID`,
		Arity: rdb.OneMust,
	}

	ctx := context.Background()
	res := env.dbTLS.Query(ctx, cmd)
	defer res.Close()

	res.Scan()
	encrypted := res.Get("encrypted")

	t.Logf("Connection encryption: %v", encrypted)

	// The connection should report TRUE for encryption
	if s, ok := encrypted.(string); ok {
		if s != "TRUE" {
			t.Logf("Warning: Connection may not be encrypted: %s", s)
		}
	}
}

// TestDockerTDS8 tests TDS 8.0 protocol connections.
func TestDockerTDS8(t *testing.T) {
	env := setupDockerEnv(t)
	if env == nil {
		return
	}

	t.Run("tds8_only", func(t *testing.T) {
		// Try TDS 8.0 only (no fallback).
		// This may fail if the server doesn't have proper TLS cert for TDS 8.0.
		config := &rdb.Config{
			DriverName:         "ms",
			Hostname:           env.host,
			Port:               env.port,
			Username:           "sa",
			Password:           saPassword,
			Database:           "master",
			PoolInitCapacity:   1,
			PoolMaxCapacity:    1,
			DialTimeout:        5 * time.Second,
			InsecureSkipVerify: true,
			KV:                 map[string]interface{}{"tds8": "only"},
		}

		pool, err := rdb.Open(config)
		if err != nil {
			t.Fatalf("TDS 8.0 only mode not available (expected on unconfigured servers): %v", err)
		}
		defer pool.Close()

		ctx := context.Background()
		if err := pool.Ping(ctx); err != nil {
			t.Fatalf("TDS 8.0 ping failed: %v", err)
		}

		t.Log("TDS 8.0 only mode succeeded")
	})

	t.Run("tds8_auto_fallback", func(t *testing.T) {
		// Test auto-detection with fallback to TDS 7.x.
		config := &rdb.Config{
			DriverName:         "ms",
			Hostname:           env.host,
			Port:               env.port,
			Username:           "sa",
			Password:           saPassword,
			Database:           "master",
			PoolInitCapacity:   1,
			PoolMaxCapacity:    1,
			DialTimeout:        5 * time.Second,
			Secure:             true,
			InsecureSkipVerify: true,
		}

		pool, err := rdb.Open(config)
		if err != nil {
			t.Fatalf("Failed to open connection with auto TDS 8.0 fallback: %v", err)
		}
		defer pool.Close()

		ctx := context.Background()
		if err := pool.Ping(ctx); err != nil {
			t.Fatalf("Failed to ping: %v", err)
		}

		// Verify encryption is active.
		cmd := &rdb.Command{
			SQL:   `SELECT encrypt_option FROM sys.dm_exec_connections WHERE session_id = @@SPID`,
			Arity: rdb.OneMust,
		}
		res, err := pool.Query(ctx, cmd)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		defer res.Close()
		res.Scan()
		encrypted := res.Getx(0)
		t.Logf("Connection encryption (with fallback): %v", encrypted)
	})

	t.Run("tds8_disable", func(t *testing.T) {
		// Test explicit TDS 8.0 disable (force TDS 7.x).
		config := &rdb.Config{
			DriverName:         "ms",
			Hostname:           env.host,
			Port:               env.port,
			Username:           "sa",
			Password:           saPassword,
			Database:           "master",
			PoolInitCapacity:   1,
			PoolMaxCapacity:    1,
			DialTimeout:        5 * time.Second,
			Secure:             true,
			InsecureSkipVerify: true,
			KV:                 map[string]interface{}{"tds8": "disable"},
		}

		pool, err := rdb.Open(config)
		if err != nil {
			t.Fatalf("Failed to open connection with TDS 8.0 disabled: %v", err)
		}
		defer pool.Close()

		ctx := context.Background()
		if err := pool.Ping(ctx); err != nil {
			t.Fatalf("Failed to ping: %v", err)
		}

		t.Log("TDS 7.x connection (TDS 8.0 disabled) succeeded")
	})
}

// TestDockerIntegerTypes tests all integer type variants.
func TestDockerIntegerTypes(t *testing.T) {
	env := setupDockerEnv(t)
	if env == nil {
		return
	}

	tests := []struct {
		name     string
		sqlType  string
		rdbType  rdb.Type
		value    interface{}
		wantType string
	}{
		// TinyInt (1 byte, unsigned 0-255)
		{name: "tinyint_0", sqlType: "tinyint", rdbType: rdb.TypeInt8, value: byte(0), wantType: "int8"},
		{name: "tinyint_255", sqlType: "tinyint", rdbType: rdb.TypeInt8, value: byte(255), wantType: "int8"},
		{name: "tinyint_127", sqlType: "tinyint", rdbType: rdb.TypeInt8, value: byte(127), wantType: "int8"},

		// SmallInt (2 bytes, signed)
		{name: "smallint_min", sqlType: "smallint", rdbType: rdb.TypeInt16, value: int16(-32768), wantType: "int16"},
		{name: "smallint_max", sqlType: "smallint", rdbType: rdb.TypeInt16, value: int16(32767), wantType: "int16"},
		{name: "smallint_0", sqlType: "smallint", rdbType: rdb.TypeInt16, value: int16(0), wantType: "int16"},

		// Int (4 bytes, signed)
		{name: "int_min", sqlType: "int", rdbType: rdb.TypeInt32, value: int32(-2147483648), wantType: "int32"},
		{name: "int_max", sqlType: "int", rdbType: rdb.TypeInt32, value: int32(2147483647), wantType: "int32"},
		{name: "int_0", sqlType: "int", rdbType: rdb.TypeInt32, value: int32(0), wantType: "int32"},

		// BigInt (8 bytes, signed)
		{name: "bigint_min", sqlType: "bigint", rdbType: rdb.TypeInt64, value: int64(-9223372036854775808), wantType: "int64"},
		{name: "bigint_max", sqlType: "bigint", rdbType: rdb.TypeInt64, value: int64(9223372036854775807), wantType: "int64"},
		{name: "bigint_0", sqlType: "bigint", rdbType: rdb.TypeInt64, value: int64(0), wantType: "int64"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &rdb.Command{
				SQL:   fmt.Sprintf(`SELECT v = CAST(@input AS %s)`, tc.sqlType),
				Arity: rdb.OneMust,
			}

			ctx := context.Background()
			res := env.db.Query(ctx, cmd,
				rdb.Param{Name: "input", Type: tc.rdbType, Value: tc.value},
			)
			defer res.Close()

			res.Scan()
			got := res.Getx(0)

			// Check type name
			gotType := fmt.Sprintf("%T", got)
			if gotType != tc.wantType {
				t.Errorf("type mismatch: got %s, want %s", gotType, tc.wantType)
			}
		})
	}
}

// TestDockerFloatTypes tests float type variants.
func TestDockerFloatTypes(t *testing.T) {
	env := setupDockerEnv(t)
	if env == nil {
		return
	}

	tests := []struct {
		name    string
		sqlType string
		rdbType rdb.Type
		value   float64
	}{
		{name: "real_positive", sqlType: "real", rdbType: rdb.TypeFloat32, value: 3.14159},
		{name: "real_negative", sqlType: "real", rdbType: rdb.TypeFloat32, value: -3.14159},
		{name: "real_zero", sqlType: "real", rdbType: rdb.TypeFloat32, value: 0.0},
		{name: "float_positive", sqlType: "float", rdbType: rdb.TypeFloat64, value: 3.141592653589793},
		{name: "float_negative", sqlType: "float", rdbType: rdb.TypeFloat64, value: -3.141592653589793},
		{name: "float_large", sqlType: "float", rdbType: rdb.TypeFloat64, value: 1.7976931348623157e+308},
		{name: "float_small", sqlType: "float", rdbType: rdb.TypeFloat64, value: 2.2250738585072014e-308},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &rdb.Command{
				SQL:   fmt.Sprintf(`SELECT v = CAST(@input AS %s)`, tc.sqlType),
				Arity: rdb.OneMust,
			}

			var inputValue interface{}
			if tc.rdbType == rdb.TypeFloat32 {
				inputValue = float32(tc.value)
			} else {
				inputValue = tc.value
			}

			ctx := context.Background()
			res := env.db.Query(ctx, cmd,
				rdb.Param{Name: "input", Type: tc.rdbType, Value: inputValue},
			)
			defer res.Close()

			res.Scan()
			got := res.Getx(0)

			var gotFloat float64
			switch v := got.(type) {
			case float32:
				gotFloat = float64(v)
			case float64:
				gotFloat = v
			default:
				t.Fatalf("unexpected type: %T", got)
			}

			// For float32, we lose precision
			if tc.rdbType == rdb.TypeFloat32 {
				if float32(gotFloat) != float32(tc.value) {
					t.Errorf("value mismatch: got %v, want %v", float32(gotFloat), float32(tc.value))
				}
			} else {
				if gotFloat != tc.value {
					t.Errorf("value mismatch: got %v, want %v", gotFloat, tc.value)
				}
			}
		})
	}
}

// TestDockerDateTimeTypes tests date/time type variants per TDS 7.3+ spec.
func TestDockerDateTimeTypes(t *testing.T) {
	env := setupDockerEnv(t)
	if env == nil {
		return
	}

	loc := time.UTC

	tests := []struct {
		name    string
		rdbType rdb.Type
		value   time.Time
	}{
		// Date only (TDS DATENTYPE 0x28)
		{name: "date_min", rdbType: rdb.TypeDate, value: time.Date(1, 1, 1, 0, 0, 0, 0, loc)},
		{name: "date_max", rdbType: rdb.TypeDate, value: time.Date(9999, 12, 31, 0, 0, 0, 0, loc)},
		{name: "date_now", rdbType: rdb.TypeDate, value: time.Now().Truncate(24 * time.Hour).In(loc)},

		// Time only (TDS TIMENTYPE 0x29)
		{name: "time_midnight", rdbType: rdb.TypeTime, value: time.Date(1, 1, 1, 0, 0, 0, 0, loc)},
		{name: "time_noon", rdbType: rdb.TypeTime, value: time.Date(1, 1, 1, 12, 0, 0, 0, loc)},
		{name: "time_precise", rdbType: rdb.TypeTime, value: time.Date(1, 1, 1, 23, 59, 59, 999999900, loc)},

		// DateTime2 (TDS DATETIME2NTYPE 0x2A)
		{name: "datetime2_min", rdbType: rdb.TypeTimestamp, value: time.Date(1, 1, 1, 0, 0, 0, 0, loc)},
		{name: "datetime2_now", rdbType: rdb.TypeTimestamp, value: time.Now().In(loc).Truncate(100 * time.Nanosecond)},

		// DateTimeOffset (TDS DATETIMEOFFSETNTYPE 0x2B)
		{name: "datetimeoffset_utc", rdbType: rdb.TypeTimestampz, value: time.Now().UTC().Truncate(100 * time.Nanosecond)},
		{name: "datetimeoffset_local", rdbType: rdb.TypeTimestampz, value: time.Now().Truncate(100 * time.Nanosecond)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &rdb.Command{
				SQL:   `SELECT v = @input`,
				Arity: rdb.OneMust,
			}

			ctx := context.Background()
			res := env.db.Query(ctx, cmd,
				rdb.Param{Name: "input", Type: tc.rdbType, Value: tc.value},
			)
			defer res.Close()

			res.Scan()
			got := res.Getx(0)

			switch v := got.(type) {
			case time.Time:
				// For date-only comparisons, truncate to date
				switch tc.rdbType {
				case rdb.TypeDate:
					gotDate := v.Truncate(24 * time.Hour)
					wantDate := tc.value.Truncate(24 * time.Hour)
					if !gotDate.Equal(wantDate) {
						t.Errorf("date mismatch: got %v, want %v", gotDate, wantDate)
					}
				case rdb.TypeTime:
					// For time-only, compare just the time portion
					gotTime := v.Sub(time.Date(v.Year(), v.Month(), v.Day(), 0, 0, 0, 0, v.Location()))
					wantTime := tc.value.Sub(time.Date(tc.value.Year(), tc.value.Month(), tc.value.Day(), 0, 0, 0, 0, tc.value.Location()))
					// Allow 1 microsecond tolerance for time precision
					diff := gotTime - wantTime
					if diff < 0 {
						diff = -diff
					}
					if diff > time.Microsecond {
						t.Errorf("time mismatch: got %v, want %v (diff: %v)", gotTime, wantTime, diff)
					}
				default:
					// For datetime types, allow small tolerance
					diff := v.Sub(tc.value)
					if diff < 0 {
						diff = -diff
					}
					if diff > time.Microsecond {
						t.Errorf("datetime mismatch: got %v, want %v (diff: %v)", v, tc.value, diff)
					}
				}
			case time.Duration:
				// Time type returns as Duration
				wantDur := tc.value.Sub(time.Date(tc.value.Year(), tc.value.Month(), tc.value.Day(), 0, 0, 0, 0, tc.value.Location()))
				diff := v - wantDur
				if diff < 0 {
					diff = -diff
				}
				if diff > time.Microsecond {
					t.Errorf("duration mismatch: got %v, want %v", v, wantDur)
				}
			default:
				t.Errorf("unexpected type: %T", got)
			}
		})
	}
}

// TestDockerNullValues tests NULL handling for all types.
func TestDockerNullValues(t *testing.T) {
	env := setupDockerEnv(t)
	if env == nil {
		return
	}

	types := []struct {
		name    string
		rdbType rdb.Type
		prec    int
		scale   int
		length  int
	}{
		{name: "tinyint", rdbType: rdb.TypeInt8},
		{name: "smallint", rdbType: rdb.TypeInt16},
		{name: "int", rdbType: rdb.TypeInt32},
		{name: "bigint", rdbType: rdb.TypeInt64},
		{name: "real", rdbType: rdb.TypeFloat32},
		{name: "float", rdbType: rdb.TypeFloat64},
		{name: "decimal", rdbType: rdb.TypeDecimal, prec: 18, scale: 4},
		{name: "bit", rdbType: rdb.TypeBool},
		{name: "date", rdbType: rdb.TypeDate},
		{name: "time", rdbType: rdb.TypeTime},
		{name: "datetime2", rdbType: rdb.TypeTimestamp},
		{name: "datetimeoffset", rdbType: rdb.TypeTimestampz},
		{name: "varchar", rdbType: rdb.TypeAnsiVarChar, length: 100},
		{name: "nvarchar", rdbType: rdb.TypeVarChar, length: 100},
		{name: "varbinary", rdbType: rdb.TypeBinary, length: 100},
	}

	for _, tc := range types {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &rdb.Command{
				SQL:   `SELECT v = @input`,
				Arity: rdb.OneMust,
			}

			param := rdb.Param{
				Name:      "input",
				Type:      tc.rdbType,
				Value:     nil,
				Null:      true,
				Precision: tc.prec,
				Scale:     tc.scale,
				Length:    tc.length,
			}

			ctx := context.Background()
			res := env.db.Query(ctx, cmd, param)
			defer res.Close()

			res.Scan()
			got := res.Getx(0)

			if got != nil {
				t.Errorf("expected nil, got %v (%T)", got, got)
			}
		})
	}
}

// TestDockerUnicode tests Unicode string handling (nvarchar).
func TestDockerUnicode(t *testing.T) {
	env := setupDockerEnv(t)
	if env == nil {
		return
	}

	tests := []struct {
		name  string
		value string
	}{
		{name: "ascii", value: "Hello, World!"},
		{name: "latin1", value: "Héllo, Wörld!"},
		{name: "cyrillic", value: "Привет мир"},
		{name: "chinese", value: "你好世界"},
		{name: "japanese", value: "こんにちは世界"},
		{name: "korean", value: "안녕하세요"},
		{name: "arabic", value: "مرحبا بالعالم"},
		{name: "emoji", value: "Hello 👋 World 🌍"},
		{name: "mixed", value: "Hello Привет 你好 🌍"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &rdb.Command{
				SQL:   `SELECT v = @input`,
				Arity: rdb.OneMust,
			}

			ctx := context.Background()
			res := env.db.Query(ctx, cmd,
				rdb.Param{Name: "input", Type: rdb.TypeVarChar, Value: tc.value, Length: 1000},
			)
			defer res.Close()

			res.Scan()
			got := res.Getx(0)

			var gotStr string
			switch v := got.(type) {
			case string:
				gotStr = v
			case []byte:
				gotStr = string(v)
			default:
				t.Fatalf("unexpected type: %T", got)
			}

			if gotStr != tc.value {
				t.Errorf("value mismatch:\ngot:  %q\nwant: %q", gotStr, tc.value)
			}
		})
	}
}

// TestDockerLargeData tests large data handling (varchar(max), varbinary(max)).
func TestDockerLargeData(t *testing.T) {
	env := setupDockerEnv(t)
	if env == nil {
		return
	}

	// Generate test data of various sizes
	sizes := []int{
		100,     // Small
		8000,    // Max non-LOB varchar
		8001,    // Just over, triggers MAX handling
		65536,   // 64KB
		1048576, // 1MB
	}

	for _, size := range sizes {
		t.Run(fmt.Sprintf("varchar_%d", size), func(t *testing.T) {
			// Generate test string
			data := strings.Repeat("A", size)

			cmd := &rdb.Command{
				SQL:   `SELECT v = @input, len = LEN(@input)`,
				Arity: rdb.OneMust,
			}

			ctx := context.Background()
			res := env.db.Query(ctx, cmd,
				rdb.Param{Name: "input", Type: rdb.TypeVarChar, Value: data, Length: 0}, // 0 = max
			)
			defer res.Close()

			res.Scan()
			got := res.Get("v")
			gotLen := res.Get("len")

			var gotStr string
			switch v := got.(type) {
			case string:
				gotStr = v
			case []byte:
				gotStr = string(v)
			default:
				t.Fatalf("unexpected type: %T", got)
			}

			if len(gotStr) != size {
				t.Errorf("length mismatch: got %d, want %d", len(gotStr), size)
			}

			if gotStr != data {
				t.Errorf("data mismatch at size %d", size)
			}

			t.Logf("Size %d: reported len=%v", size, gotLen)
		})
	}
}

// TestDockerTLSCertChain tests TLS certificate chain verification.
// This test:
// 1. Generates a CA and server certificate in memory
// 2. Starts a Docker container with the custom certificate
// 3. Connects with InsecureSkipVerify to verify basic connectivity
// 4. Connects with the CA in RootCAs to verify cert chain validation
// 5. Verifies that connection fails with wrong CA
func TestDockerTLSCertChain(t *testing.T) {
	if !checkDockerAvailable(t) {
		return
	}

	const (
		proto74 = 0x74000004
		proto80 = 0x8000000
	)

	// Use a different container name and port to avoid conflicts
	// Use SQL Server 2025 for TDS 8.0 (strict encryption) support
	const (
		tlsContainerName = "rdb-mssql-tls-test"
		tlsPort          = 11434
		tlsDockerImage   = dockerImage
	)

	// Clean up any existing container
	exec.Command("docker", "stop", tlsContainerName).Run()
	exec.Command("docker", "rm", tlsContainerName).Run()

	// Generate CA and server certificate
	t.Log("Generating CA and server certificate...")
	ca, err := testcert.GenerateCA("Test CA", 24*time.Hour)
	if err != nil {
		t.Fatalf("generate CA: %v", err)
	}

	serverCert, err := ca.GenerateServerCert(
		"localhost",
		[]string{"localhost"},
		[]net.IP{net.ParseIP("127.0.0.1")},
		24*time.Hour,
	)
	if err != nil {
		t.Fatalf("generate server cert: %v", err)
	}

	// Create temp directory for certificates and config
	tmpDir := t.TempDir()
	// Make temp dir accessible by mssql user (uid 10001) in container
	os.Chmod(tmpDir, 0755)

	// Create certs subdirectory for copying to container
	certsDir := filepath.Join(tmpDir, "certs")
	if err := os.Mkdir(certsDir, 0755); err != nil {
		t.Fatalf("mkdir certs: %v", err)
	}

	// Write certificate and key files with world-readable permissions
	// (mssql user uid 10001 needs to read them)
	certPath := filepath.Join(certsDir, "server.pem")
	keyPath := filepath.Join(certsDir, "server.key")

	if err := os.WriteFile(certPath, serverCert.CertPEM, 0644); err != nil {
		t.Fatalf("write cert: %v", err)
	}
	if err := os.WriteFile(keyPath, serverCert.KeyPEM, 0644); err != nil {
		t.Fatalf("write key: %v", err)
	}

	// Create mssql.conf for TLS configuration
	// SQL Server 2025 with forceencryption supports TDS 8.0
	mssqlConf := `[network]
tlscert = /certs/server.pem
tlskey = /certs/server.key
tlsprotocols = 1.2
forceencryption = 1
`
	confPath := filepath.Join(tmpDir, "mssql.conf")
	if err := os.WriteFile(confPath, []byte(mssqlConf), 0644); err != nil {
		t.Fatalf("write mssql.conf: %v", err)
	}

	// Pull the image first
	t.Logf("Pulling MSSQL Docker image %s...", tlsDockerImage)
	pullCmd := exec.Command("docker", "pull", tlsDockerImage)
	pullCmd.Stdout = os.Stdout
	pullCmd.Stderr = os.Stderr
	if err := pullCmd.Run(); err != nil {
		t.Fatalf("docker pull: %v", err)
	}

	// Start container with mssql.conf and certs mounted
	t.Log("Starting MSSQL container...")
	runArgs := []string{
		"run", "-d",
		"--name", tlsContainerName,
		"-e", "ACCEPT_EULA=Y",
		"-e", "MSSQL_SA_PASSWORD=" + saPassword,
		"-e", "MSSQL_PID=Developer",
		"-p", fmt.Sprintf("%d:1433", tlsPort),
		"-v", confPath + ":/var/opt/mssql/mssql.conf",
		"-v", certsDir + ":/certs",
		tlsDockerImage,
	}

	var stdout, stderr bytes.Buffer
	runCmd := exec.Command("docker", runArgs...)
	runCmd.Stdout = &stdout
	runCmd.Stderr = &stderr
	if err := runCmd.Run(); err != nil {
		t.Fatalf("docker run: %v\nstderr: %s", err, stderr.String())
	}
	containerID := strings.TrimSpace(stdout.String())
	t.Logf("Container started: %s", containerID[:12])

	// Ensure cleanup
	defer func() {
		t.Log("Stopping TLS test container...")
		exec.Command("docker", "stop", tlsContainerName).Run()
		exec.Command("docker", "rm", tlsContainerName).Run()
	}()

	// Wait for SQL Server to be ready (no restart needed)
	t.Log("Waiting for SQL Server to start with TLS...")
	if err := waitForMSSQLReady(containerID, 90*time.Second); err != nil {
		t.Fatalf("wait for MSSQL: %v", err)
	}

	runCheck := func(t *testing.T, expectConnect bool, wantProto int64, config *rdb.Config) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*6)
		defer cancel()

		cp, err := rdb.Open(config)
		if err != nil {
			t.Fatal(err)
		}
		err = cp.Ping(ctx)
		switch expectConnect {
		case false:
			if err != nil {
				t.Logf("expected failure: %v", err)
				return
			}
			t.Fatalf("ping incorrectly worked: %v", err)
		case true:
			if err != nil {
				t.Fatalf("ping failed: %v", err)
			}
			res, err := cp.Query(ctx, &rdb.Command{
				SQL: `
SELECT session_id, protocol_type, protocol_version, encrypt_option
FROM sys.dm_exec_connections
WHERE session_id = @@SPID;
`,
			})
			if err != nil {
				t.Fatalf("failed to query exec_connections: %v", err)
			}
			var sessionID, protocolVersion int64
			var protocolType, encryptOption string
			err = res.Prep("session_id", &sessionID).Prep("protocol_type", &protocolType).Prep("protocol_version", &protocolVersion).Prep("encrypt_option", &encryptOption).Scan()
			if err != nil {
				t.Fatalf("failed to scan query: %v", err)
			}
			t.Logf("SID=%d ProtoType=%s ProtoVer=0x%x Encrypt=%s", sessionID, protocolType, protocolVersion, encryptOption)
			if wantProto > 0 {
				if wantProto != protocolVersion {
					t.Fatalf("wanted protocol 0x%x, got 0x%x", wantProto, protocolVersion)
				}
			}
		}
	}

	// Test 1: Connect with InsecureSkipVerify (should work)
	t.Run("insecure_skip_verify", func(t *testing.T) {
		runCheck(t, true, proto74, &rdb.Config{
			DriverName:         "ms",
			Hostname:           "127.0.0.1",
			Port:               tlsPort,
			Username:           "sa",
			Password:           saPassword,
			Database:           "master",
			PoolInitCapacity:   1,
			PoolMaxCapacity:    1,
			DialTimeout:        10 * time.Second,
			InsecureSkipVerify: true,
		})
	})

	// Test 2: Connect with correct CA in RootCAs (should work)
	t.Run("valid_ca_chain", func(t *testing.T) {
		runCheck(t, true, proto80, &rdb.Config{
			DriverName:         "ms",
			Hostname:           "localhost", // Must match cert's CN/SAN
			Port:               tlsPort,
			Username:           "sa",
			Password:           saPassword,
			Database:           "master",
			PoolInitCapacity:   1,
			PoolMaxCapacity:    1,
			DialTimeout:        10 * time.Second,
			Secure:             true,
			InsecureSkipVerify: false,
			RootCAs:            ca.CertPool(),
		})
	})

	// Test 3: Connect with wrong CA (should fail)
	t.Run("invalid_ca_chain", func(t *testing.T) {
		// Generate a different CA that didn't sign the server cert
		wrongCA, err := testcert.GenerateCA("Wrong CA", 24*time.Hour)
		if err != nil {
			t.Fatalf("generate wrong CA: %v", err)
		}

		runCheck(t, false, 0, &rdb.Config{
			DriverName:         "ms",
			Hostname:           "localhost",
			Port:               tlsPort,
			Username:           "sa",
			Password:           saPassword,
			Database:           "master",
			PoolInitCapacity:   1,
			PoolMaxCapacity:    1,
			DialTimeout:        10 * time.Second,
			Secure:             true,
			InsecureSkipVerify: false,
			RootCAs:            wrongCA.CertPool(), // Wrong CA
		})
	})

	// Test 4: Connect without any CA (should fail when Secure=true)
	t.Run("no_ca_secure", func(t *testing.T) {
		runCheck(t, false, 0, &rdb.Config{
			DriverName:         "ms",
			Hostname:           "localhost",
			Port:               tlsPort,
			Username:           "sa",
			Password:           saPassword,
			Database:           "master",
			PoolInitCapacity:   1,
			PoolMaxCapacity:    1,
			DialTimeout:        10 * time.Second,
			Secure:             true,
			InsecureSkipVerify: false,
			// No RootCAs - will use system pool which won't have our CA
		})
	})

	// Test 5: Connect with TDS 8.0 (TLS-first with ALPN)
	t.Run("tds8_with_cert", func(t *testing.T) {
		runCheck(t, true, proto80, &rdb.Config{
			DriverName:       "ms",
			Hostname:         "localhost",
			Port:             tlsPort,
			Username:         "sa",
			Password:         saPassword,
			Database:         "master",
			PoolInitCapacity: 1,
			PoolMaxCapacity:  1,
			DialTimeout:      10 * time.Second,
			Secure:           true,
			RootCAs:          ca.CertPool(),
			KV:               map[string]interface{}{"tds8": "only"},
		})
	})
}
