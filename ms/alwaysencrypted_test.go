// Copyright 2026 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package ms

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/kardianos/rdb"
	"github.com/kardianos/rdb/internal/uconv"
	"github.com/kardianos/rdb/ms/aecrypt"
	"github.com/kardianos/rdb/table"
)

func TestPlainStream(t *testing.T) {
	nvarchar := &SQLColumn{Column: rdb.Column{Name: "N"}, code: typeNVarChar, info: typeInfoLookup[typeNVarChar]}
	nvarcharMax := &SQLColumn{Column: rdb.Column{Name: "M", Unlimit: true}, code: typeNVarChar, info: typeInfoLookup[typeNVarChar]}
	intN := func(width int) *SQLColumn {
		return &SQLColumn{Column: rdb.Column{Name: "I", Length: width}, code: typeIntN, info: typeInfoLookup[typeIntN]}
	}
	eight := []byte{0xFE, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF} // -2

	list := []struct {
		Name    string
		Column  *SQLColumn
		Plain   []byte
		Null    bool
		Want    string // Hex.
		WantErr string
	}{
		{Name: "nvarchar", Column: nvarchar, Plain: []byte{'a', 0, 'b', 0}, Want: "04006100" + "6200"},
		{Name: "nvarchar empty", Column: nvarchar, Plain: []byte{}, Want: "0000"},
		{Name: "nvarchar null", Column: nvarchar, Null: true, Want: "ffff"},
		{Name: "nvarchar odd length", Column: nvarchar, Plain: []byte{'a'}, WantErr: "not UTF-16"},
		{Name: "nvarchar(max)", Column: nvarcharMax, Plain: []byte{'a', 0}, Want: "0200000000000000" + "02000000" + "6100" + "00000000"},
		{Name: "nvarchar(max) empty", Column: nvarcharMax, Plain: []byte{}, Want: "0000000000000000" + "00000000"},
		{Name: "nvarchar(max) null", Column: nvarcharMax, Null: true, Want: "ffffffffffffffff"},
		{Name: "int keeps 4 of 8 bytes", Column: intN(4), Plain: eight, Want: "04" + "feffffff"},
		{Name: "bigint", Column: intN(8), Plain: eight, Want: "08" + "feffffffffffffff"},
		{Name: "tinyint", Column: intN(1), Plain: []byte{7, 0, 0, 0, 0, 0, 0, 0}, Want: "0107"},
		{Name: "int null", Column: intN(4), Null: true, Want: "00"},
		{Name: "int short plaintext", Column: intN(4), Plain: []byte{1, 2}, WantErr: "decrypted 2 bytes for a 4 byte integer"},
	}
	for _, tc := range list {
		t.Run(tc.Name, func(t *testing.T) {
			got, err := plainStream(tc.Column, tc.Plain, tc.Null)
			if len(tc.WantErr) > 0 {
				if err == nil || !strings.Contains(err.Error(), tc.WantErr) {
					t.Fatalf("got error %v, want one containing %q", err, tc.WantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if hex.EncodeToString(got) != tc.Want {
				t.Fatalf("got %x, want %s", got, tc.Want)
			}
		})
	}
}

func TestAEPlaintext(t *testing.T) {
	param := func(name string, tp rdb.Type, length int) *rdb.Param {
		return &rdb.Param{Name: name, Type: tp, Length: length}
	}
	list := []struct {
		Name     string
		Param    *rdb.Param
		Value    interface{}
		Want     string // Hex.
		WantNull bool
		WantErr  string
	}{
		{Name: "nvarchar string", Param: param("s", rdb.TypeVarChar, 10), Value: "ab", Want: "61006200"},
		{Name: "nvarchar bytes", Param: param("s", rdb.TypeVarChar, 10), Value: []byte("ab"), Want: "61006200"},
		{Name: "nvarchar non-ASCII", Param: param("s", rdb.TypeVarChar, 10), Value: "é", Want: "e900"},
		{Name: "nvarchar too long", Param: param("s", rdb.TypeVarChar, 1), Value: "ab", WantErr: "too long"},
		{Name: "nvarchar(max)", Param: param("s", rdb.TypeVarChar, 0), Value: "ab", Want: "61006200"},
		{Name: "nchar pads with spaces", Param: param("s", rdb.TypeChar, 4), Value: "ab", Want: "6100620020002000"},
		{Name: "nchar full", Param: param("s", rdb.TypeChar, 2), Value: "ab", Want: "61006200"},
		{Name: "varbinary", Param: param("b", rdb.TypeBinary, 4), Value: []byte{1, 2}, Want: "0102"},
		{Name: "int", Param: param("i", rdb.TypeInt32, 0), Value: int32(-2), Want: "feffffffffffffff"},
		{Name: "int from Go int", Param: param("i", rdb.TypeInt32, 0), Value: 3, Want: "0300000000000000"},
		{Name: "bigint", Param: param("i", rdb.TypeInt64, 0), Value: int64(1) << 40, Want: "0000000000010000"},
		{Name: "null", Param: param("s", rdb.TypeVarChar, 10), Value: nil, WantNull: true},
		{Name: "varchar unsupported", Param: param("s", rdb.TypeAnsiVarChar, 10), Value: "ab", WantErr: "encrypting varchar values is not supported"},
		{Name: "date unsupported", Param: param("d", rdb.TypeDate, 0), Value: "2026-01-02", WantErr: "encrypting date values is not supported"},
		{Name: "int from string", Param: param("i", rdb.TypeInt32, 0), Value: "3", WantErr: "encrypting string as int"},
	}
	for _, tc := range list {
		t.Run(tc.Name, func(t *testing.T) {
			ti, err := getParamTypeInfo(protoVer74, tc.Param.Type)
			if err != nil {
				t.Fatal(err)
			}
			got, null, err := aePlaintext(ti, tc.Param, tc.Value)
			if len(tc.WantErr) > 0 {
				if err == nil || !strings.Contains(err.Error(), tc.WantErr) {
					t.Fatalf("got error %v, want one containing %q", err, tc.WantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if null != tc.WantNull {
				t.Fatalf("null %t, want %t", null, tc.WantNull)
			}
			if hex.EncodeToString(got) != tc.Want {
				t.Fatalf("got %x, want %s", got, tc.Want)
			}
		})
	}
}

// TestDockerAlwaysEncrypted runs every Always Encrypted scenario against one
// database. The server only stores key metadata, so the test makes up a
// column master key and wraps its own column encryption key.
func TestDockerAlwaysEncrypted(t *testing.T) {
	env := setupDockerEnv(t)
	ctx := context.Background()
	plainMaster := env.db.Normal()

	const dbName = "rdb_always_encrypted"
	exec := func(t *testing.T, pool *rdb.ConnPool, sql string, params ...rdb.Param) error {
		t.Helper()
		_, err := pool.Query(ctx, &rdb.Command{SQL: sql, Arity: rdb.Zero}, params...)
		return err
	}
	if err := exec(t, plainMaster, "if db_id(N'"+dbName+"') is not null begin alter database "+dbName+" set single_user with rollback immediate; drop database "+dbName+"; end; create database "+dbName+";"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		exec(t, plainMaster, "alter database "+dbName+" set single_user with rollback immediate; drop database "+dbName+";")
	})

	cmk, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	newKey := func(t *testing.T) (*aecrypt.Key, []byte) {
		t.Helper()
		root := make([]byte, aecrypt.KeySize)
		if _, err := rand.Read(root); err != nil {
			t.Fatal(err)
		}
		wrapped, err := aecrypt.Wrap(root, "CurrentUser/My/0123456789ABCDEF0123456789ABCDEF01234567", cmk)
		if err != nil {
			t.Fatal(err)
		}
		k, err := aecrypt.NewKey(root, wrapped)
		if err != nil {
			t.Fatal(err)
		}
		return k, wrapped
	}
	key, wrapped := newKey(t)
	otherKey, _ := newKey(t)

	config := func(keys ...string) *rdb.Config {
		return &rdb.Config{
			DriverName:         "ms",
			Hostname:           env.host,
			Port:               env.port,
			Username:           "sa",
			Password:           saPassword,
			Database:           dbName,
			PoolInitCapacity:   1,
			PoolMaxCapacity:    2,
			DialTimeout:        5 * time.Second,
			InsecureSkipVerify: true,
			ColumnKeys:         keys,
		}
	}
	openPool := func(t *testing.T, keys ...string) *rdb.ConnPool {
		t.Helper()
		pool, err := rdb.Open(config(keys...))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(pool.Close)
		return pool
	}
	plain := openPool(t)
	ae := openPool(t, key.Text())

	const enc = "column_encryption_key = CEK_rdb, algorithm = 'AEAD_AES_256_CBC_HMAC_SHA_256'"
	setup := `
create column master key CMK_rdb with (key_store_provider_name = N'MSSQL_CERTIFICATE_STORE', key_path = N'CurrentUser/My/0123456789ABCDEF0123456789ABCDEF01234567');
create column encryption key CEK_rdb with values (column_master_key = CMK_rdb, algorithm = 'RSA_OAEP', encrypted_value = 0x` + hex.EncodeToString(wrapped) + `);
`
	if err := exec(t, plain, setup); err != nil {
		t.Fatal(err)
	}
	schema := `
create table dbo.Sample (
	ID int not null primary key,
	Name nvarchar(60) collate Latin1_General_BIN2 encrypted with (encryption_type = deterministic, ` + enc + `) null,
	Note nvarchar(max) encrypted with (encryption_type = randomized, ` + enc + `) null,
	Num int encrypted with (encryption_type = deterministic, ` + enc + `) null,
	Raw varbinary(50) encrypted with (encryption_type = randomized, ` + enc + `) null,
	Fixed nchar(10) collate Latin1_General_BIN2 encrypted with (encryption_type = deterministic, ` + enc + `) null,
	Plain nvarchar(20) null
);
create table dbo.Dated (
	ID int not null,
	At datetime2 encrypted with (encryption_type = randomized, ` + enc + `) null
);
`
	if err := exec(t, plain, schema); err != nil {
		t.Fatal(err)
	}

	type sample struct {
		ID    int32
		Name  interface{} // nil for null.
		Note  interface{}
		Num   interface{}
		Raw   interface{}
		Fixed interface{}
		Plain string
	}
	longNote := strings.Repeat("0123456789", 900) // Over 8000 bytes of UTF-16: PLP.
	rows := []sample{
		{ID: 1, Name: "251017004HLA", Note: longNote, Num: int32(42), Raw: []byte{1, 2, 3}, Fixed: "12345", Plain: "p1"},
		{ID: 2, Name: "251017004dHLA", Note: "", Num: int32(-7), Plain: "p2"},
		{ID: 3, Plain: "p3"},
	}
	insertParams := func(s sample) []rdb.Param {
		return []rdb.Param{
			{Name: "ID", Type: rdb.TypeInt32, Value: s.ID},
			{Name: "Name", Type: rdb.TypeVarChar, Length: 60, Value: s.Name},
			{Name: "Note", Type: rdb.TypeVarChar, Value: s.Note},
			{Name: "Num", Type: rdb.TypeInt32, Value: s.Num},
			{Name: "Raw", Type: rdb.TypeBinary, Length: 50, Value: s.Raw},
			{Name: "Fixed", Type: rdb.TypeChar, Length: 10, Value: s.Fixed},
			{Name: "Plain", Type: rdb.TypeVarChar, Length: 20, Value: s.Plain},
		}
	}
	const insertSQL = "insert into dbo.Sample (ID, Name, Note, Num, Raw, Fixed, Plain) values (@ID, @Name, @Note, @Num, @Raw, @Fixed, @Plain);"

	fill := func(t *testing.T, pool *rdb.ConnPool, sql string, params ...rdb.Param) (*table.Buffer, error) {
		t.Helper()
		res, err := pool.Query(ctx, &rdb.Command{SQL: sql}, params...)
		if err != nil {
			return nil, err
		}
		return table.Fill(ctx, res)
	}
	text := func(v interface{}) interface{} {
		if b, ok := v.([]byte); ok {
			return string(b)
		}
		return v
	}

	t.Run("insert with encrypted parameters", func(t *testing.T) {
		for _, s := range rows {
			if err := exec(t, ae, insertSQL, insertParams(s)...); err != nil {
				t.Fatalf("ID %d: %v", s.ID, err)
			}
		}
	})
	t.Run("select decrypts", func(t *testing.T) {
		buf, err := fill(t, ae, "select ID, Name, Note, Num, Raw, Fixed, Plain from dbo.Sample order by ID;")
		if err != nil {
			t.Fatal(err)
		}
		if len(buf.Row) != len(rows) {
			t.Fatalf("got %d rows, want %d", len(buf.Row), len(rows))
		}
		for i, want := range rows {
			r := buf.Row[i]
			got := sample{
				ID:    r.Get("ID").(int32),
				Name:  text(r.Get("Name")),
				Note:  text(r.Get("Note")),
				Num:   r.Get("Num"),
				Raw:   r.Get("Raw"),
				Fixed: text(r.Get("Fixed")),
				Plain: text(r.Get("Plain")).(string),
			}
			if want.Fixed != nil {
				want.Fixed = fmt.Sprintf("%-10s", want.Fixed) // Stored padded.
			}
			if fmt.Sprint(got) != fmt.Sprint(want) {
				t.Errorf("row %d:\ngot  %.200v\nwant %.200v", i, got, want)
			}
		}
	})
	t.Run("where on a deterministic column", func(t *testing.T) {
		list := []struct {
			Name   string
			SQL    string
			Params []rdb.Param
			Want   string
		}{
			{
				Name:   "equal",
				SQL:    "select ID from dbo.Sample where Name = @Name order by ID;",
				Params: []rdb.Param{{Name: "Name", Type: rdb.TypeVarChar, Length: 60, Value: "251017004dHLA"}},
				Want:   "[2]",
			},
			{
				Name: "in",
				SQL:  "select ID from dbo.Sample where Name in (@A, @B) order by ID;",
				Params: []rdb.Param{
					{Name: "A", Type: rdb.TypeVarChar, Length: 60, Value: "251017004HLA"},
					{Name: "B", Type: rdb.TypeVarChar, Length: 60, Value: "251017004dHLA"},
				},
				Want: "[1 2]",
			},
			{
				Name:   "integer",
				SQL:    "select ID from dbo.Sample where Num = @Num order by ID;",
				Params: []rdb.Param{{Name: "Num", Type: rdb.TypeInt32, Value: int32(42)}},
				Want:   "[1]",
			},
			{
				Name: "with a plaintext parameter",
				SQL:  "select ID from dbo.Sample where Name = @Name and Plain = @Plain order by ID;",
				Params: []rdb.Param{
					{Name: "Name", Type: rdb.TypeVarChar, Length: 60, Value: "251017004HLA"},
					{Name: "Plain", Type: rdb.TypeVarChar, Length: 20, Value: "p1"},
				},
				Want: "[1]",
			},
			{
				Name:   "nchar padded to match",
				SQL:    "select ID from dbo.Sample where Fixed = @Fixed order by ID;",
				Params: []rdb.Param{{Name: "Fixed", Type: rdb.TypeChar, Length: 10, Value: "12345"}},
				Want:   "[1]",
			},
			{
				Name:   "parameter declared nvarchar(max)",
				SQL:    "select ID from dbo.Sample where Name = @Name order by ID;",
				Params: []rdb.Param{{Name: "Name", Type: rdb.TypeVarChar, Value: "251017004HLA"}},
				Want:   "[1]",
			},
			{
				Name:   "no match",
				SQL:    "select ID from dbo.Sample where Name = @Name order by ID;",
				Params: []rdb.Param{{Name: "Name", Type: rdb.TypeVarChar, Length: 60, Value: "251017004hla"}},
				Want:   "[]",
			},
		}
		for _, tc := range list {
			t.Run(tc.Name, func(t *testing.T) {
				buf, err := fill(t, ae, tc.SQL, tc.Params...)
				if err != nil {
					t.Fatal(err)
				}
				ids := []int32{}
				for _, r := range buf.Row {
					ids = append(ids, r.Get("ID").(int32))
				}
				if got := fmt.Sprint(ids); got != tc.Want {
					t.Fatalf("got %s, want %s", got, tc.Want)
				}
			})
		}
	})
	t.Run("plain connection sees the ciphertext", func(t *testing.T) {
		buf, err := fill(t, plain, "select Name, Note from dbo.Sample where ID = 1;")
		if err != nil {
			t.Fatal(err)
		}
		name := buf.Row[0].Get("Name").([]byte)
		want, err := key.Encrypt(uconv.Encode.FromString("251017004HLA"), true)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(name, want) {
			t.Fatalf("stored Name %x, want deterministic %x", name, want)
		}
		note, err := key.Decrypt(buf.Row[0].Get("Note").([]byte))
		if err != nil {
			t.Fatal(err)
		}
		if got := uconv.Decode.ToString(note); got != longNote {
			t.Fatalf("Note decrypts to %.40q..., want %.40q...", got, longNote)
		}
	})
	t.Run("errors", func(t *testing.T) {
		list := []struct {
			Name    string
			Keys    []string
			SQL     string
			Params  []rdb.Param
			WantErr string
		}{
			{Name: "other key", Keys: []string{otherKey.Text()}, SQL: "select Name from dbo.Sample;", WantErr: "no column_key matches"},
			{Name: "other key for a parameter", Keys: []string{otherKey.Text()}, SQL: "select ID from dbo.Sample where Name = @Name;", Params: []rdb.Param{{Name: "Name", Type: rdb.TypeVarChar, Length: 60, Value: "x"}}, WantErr: "parameter @Name: no column_key matches"},
			{Name: "damaged key text", Keys: []string{"AAAA" + key.Text()}, SQL: "select 1;", WantErr: "column_key[0]"},
			{Name: "unsupported column type", Keys: []string{key.Text()}, SQL: "select At from dbo.Dated;", WantErr: `encrypted column "At": decrypting DateTime2N values is not supported`},
			// Storing needs the column's exact type; comparing does not (see
			// "parameter declared nvarchar(max)" above).
			{Name: "insert declared longer than the column", Keys: []string{key.Text()}, SQL: "insert into dbo.Sample (ID, Name) values (9, @Name);", Params: []rdb.Param{{Name: "Name", Type: rdb.TypeVarChar, Value: "x"}}, WantErr: "Operand type clash: nvarchar(max) encrypted with"},
		}
		for _, tc := range list {
			t.Run(tc.Name, func(t *testing.T) {
				pool, err := rdb.Open(config(tc.Keys...))
				if err == nil {
					defer pool.Close()
					_, err = fill(t, pool, tc.SQL, tc.Params...)
				}
				if err == nil || !strings.Contains(err.Error(), tc.WantErr) {
					t.Fatalf("got error %v, want one containing %q", err, tc.WantErr)
				}
			})
		}
	})
	t.Run("pool recovers after an error", func(t *testing.T) {
		_, err := fill(t, ae, "select At from dbo.Dated;")
		if err == nil {
			t.Fatal("no error")
		}
		buf, err := fill(t, ae, "select Name from dbo.Sample where ID = 1;")
		if err != nil {
			t.Fatal(err)
		}
		if got := text(buf.Row[0].Get("Name")); got != "251017004HLA" {
			t.Fatalf("got %v", got)
		}
	})
}
