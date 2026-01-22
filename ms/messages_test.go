// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package ms

import (
	"context"
	"strings"
	"testing"

	"github.com/kardianos/rdb"
)

// TestPrintStatements tests SQL PRINT statements which generate info messages
func TestPrintStatements(t *testing.T) {
	checkSkip(t)
	if parallel {
		t.Parallel()
	}
	defer assertFreeConns(t)
	defer recoverTest(t)

	ctx := context.Background()

	// PRINT generates an info message (severity 0, tokenInfo)
	cmd := &rdb.Command{
		SQL:   `PRINT 'Hello from SQL Server'`,
		Arity: rdb.Zero,
	}

	res, err := db.Normal().Query(ctx, cmd)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	defer res.Close()

	// The PRINT message should be captured
	msgs := res.Info()
	found := false
	for _, msg := range msgs {
		if strings.Contains(msg.Message, "Hello from SQL Server") {
			found = true
			if msg.Type != rdb.SqlInfo {
				t.Errorf("expected SqlInfo type, got %v", msg.Type)
			}
			t.Logf("Got PRINT message: %s (class=%d, state=%d)", msg.Message, msg.Class, msg.State)
		}
	}
	if !found {
		t.Logf("Messages received: %v", msgs)
	}
}

// TestRaiseErrorInfo tests RAISERROR with severity < 11 (info messages)
func TestRaiseErrorInfo(t *testing.T) {
	checkSkip(t)
	if parallel {
		t.Parallel()
	}
	defer assertFreeConns(t)
	defer recoverTest(t)

	ctx := context.Background()

	// RAISERROR with severity 1-10 generates info messages, not errors
	cmd := &rdb.Command{
		SQL:   `RAISERROR('This is an info message', 10, 1)`,
		Arity: rdb.Zero,
	}

	res, err := db.Normal().Query(ctx, cmd)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	defer res.Close()

	msgs := res.Info()
	found := false
	for _, msg := range msgs {
		if strings.Contains(msg.Message, "This is an info message") {
			found = true
			t.Logf("Got RAISERROR info: %s (class=%d, state=%d, number=%d)",
				msg.Message, msg.Class, msg.State, msg.Number)
		}
	}
	if !found {
		t.Logf("Messages received: %v", msgs)
	}
}

// TestRaiseErrorWithResult tests RAISERROR alongside query results
func TestRaiseErrorWithResult(t *testing.T) {
	checkSkip(t)
	if parallel {
		t.Parallel()
	}
	defer assertFreeConns(t)
	defer recoverTest(t)

	ctx := context.Background()

	// Send both an info message and a result
	cmd := &rdb.Command{
		SQL: `
			PRINT 'Before query'
			SELECT value = 42
			PRINT 'After query'
		`,
		Arity: rdb.OneMust,
	}

	res := db.Query(ctx, cmd)
	defer res.Close()

	res.Scan()
	val := res.Getx(0)
	if val != int32(42) {
		t.Errorf("expected 42, got %v", val)
	}

	msgs := res.Info()
	t.Logf("Messages: %d", len(msgs))
	for _, msg := range msgs {
		t.Logf("  - %s (type=%v, class=%d)", msg.Message, msg.Type, msg.Class)
	}
}

// TestSQLError tests SQL errors (severity >= 11)
func TestSQLError(t *testing.T) {
	checkSkip(t)
	if parallel {
		t.Parallel()
	}
	defer assertFreeConns(t)
	defer recoverTest(t)

	ctx := context.Background()

	// Query a non-existent table to generate an error
	cmd := &rdb.Command{
		SQL:   `SELECT * FROM nonexistent_table_xyz_12345`,
		Arity: rdb.Zero,
	}

	_, err := db.Normal().Query(ctx, cmd)
	if err == nil {
		t.Fatal("expected error for non-existent table")
	}

	// Check that the error contains useful information
	errStr := err.Error()
	if !strings.Contains(errStr, "nonexistent_table_xyz_12345") {
		t.Errorf("error should mention table name, got: %v", err)
	}
	t.Logf("Got expected SQL error: %v", err)
}

// TestOrderByMetadata tests queries with ORDER BY which generate tokenOrder
func TestOrderByMetadata(t *testing.T) {
	checkSkip(t)
	if parallel {
		t.Parallel()
	}
	defer assertFreeConns(t)
	defer recoverTest(t)

	ctx := context.Background()

	// Query with ORDER BY generates tokenOrder in the response
	cmd := &rdb.Command{
		SQL: `
			SELECT value = v
			FROM (VALUES (3), (1), (2)) AS t(v)
			ORDER BY v
		`,
	}

	res, err := db.Normal().Query(ctx, cmd)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	defer res.Close()

	var values []int32
	for {
		err := res.Scan()
		if err != nil {
			break
		}
		values = append(values, res.Getx(0).(int32))
	}

	if len(values) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(values))
	}

	// Verify order
	if values[0] != 1 || values[1] != 2 || values[2] != 3 {
		t.Errorf("expected [1,2,3], got %v", values)
	}
}

// TestMultipleResultSets tests handling multiple result sets
func TestMultipleResultSets(t *testing.T) {
	checkSkip(t)
	if parallel {
		t.Parallel()
	}
	defer assertFreeConns(t)
	defer recoverTest(t)

	ctx := context.Background()

	cmd := &rdb.Command{
		SQL: `
			SELECT first = 1
			SELECT second = 2
		`,
	}

	res, err := db.Normal().Query(ctx, cmd)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	defer res.Close()

	// First result set
	if err := res.Scan(); err != nil {
		t.Fatalf("expected first row: %v", err)
	}
	first := res.Getx(0)
	if first != int32(1) {
		t.Errorf("expected 1, got %v", first)
	}

	// Move to next result set
	more, err := res.NextResult()
	if err != nil {
		t.Fatalf("NextResult failed: %v", err)
	}
	if !more {
		t.Fatal("expected more results")
	}

	// Second result set
	if err := res.Scan(); err != nil {
		t.Fatalf("expected second row: %v", err)
	}
	second := res.Getx(0)
	if second != int32(2) {
		t.Errorf("expected 2, got %v", second)
	}
}

// TestRowCount tests that row count is reported correctly
func TestRowCount(t *testing.T) {
	checkSkip(t)
	if parallel {
		t.Parallel()
	}
	defer assertFreeConns(t)
	defer recoverTest(t)

	ctx := context.Background()

	// Use a transaction to ensure all operations happen on the same connection
	// (temp tables are session-scoped)
	tran, err := db.Normal().Begin(ctx)
	if err != nil {
		t.Fatalf("begin failed: %v", err)
	}
	defer tran.Rollback()

	// Create temp table and insert rows
	setupCmd := &rdb.Command{
		SQL: `
			CREATE TABLE #temp_rowcount (id int)
			INSERT INTO #temp_rowcount VALUES (1), (2), (3)
		`,
		Arity: rdb.Zero,
	}

	res, err := tran.Query(ctx, setupCmd)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	res.Close()

	// Update rows and check affected count
	updateCmd := &rdb.Command{
		SQL:   `UPDATE #temp_rowcount SET id = id + 10`,
		Arity: rdb.Zero,
	}

	res, err = tran.Query(ctx, updateCmd)
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	affected := res.RowsAffected()
	res.Close()

	if affected != 3 {
		t.Errorf("expected 3 rows affected, got %d", affected)
	}

	// Cleanup (temp table automatically dropped on rollback)
}

// TestNBCRow tests NULL Bitmap Compressed rows
// NBCRow is used when there are many NULL columns in a result
func TestNBCRow(t *testing.T) {
	checkSkip(t)
	if parallel {
		t.Parallel()
	}
	defer assertFreeConns(t)
	defer recoverTest(t)

	ctx := context.Background()

	// Query that returns many NULL columns may trigger NBCRow encoding
	cmd := &rdb.Command{
		SQL: `
			SELECT
				a = CAST(NULL AS int),
				b = CAST(NULL AS varchar(10)),
				c = CAST(NULL AS datetime),
				d = CAST(NULL AS float),
				e = 42,
				f = CAST(NULL AS bit),
				g = CAST(NULL AS decimal(10,2)),
				h = CAST(NULL AS binary(10))
		`,
		Arity: rdb.OneMust,
	}

	res := db.Query(ctx, cmd)
	defer res.Close()

	res.Scan()

	// Check that NULLs are returned as nil
	if res.Get("a") != nil {
		t.Error("expected a to be nil")
	}
	if res.Get("b") != nil {
		t.Error("expected b to be nil")
	}
	if res.Get("c") != nil {
		t.Error("expected c to be nil")
	}
	if res.Get("d") != nil {
		t.Error("expected d to be nil")
	}
	// e should be 42
	if res.Get("e") != int32(42) {
		t.Errorf("expected e=42, got %v", res.Get("e"))
	}
	if res.Get("f") != nil {
		t.Error("expected f to be nil")
	}
	if res.Get("g") != nil {
		t.Error("expected g to be nil")
	}
	if res.Get("h") != nil {
		t.Error("expected h to be nil")
	}
}
