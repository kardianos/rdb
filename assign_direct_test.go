// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package rdb

import (
	"bytes"
	"io"
	"testing"
)

func TestDirectAssignInt(t *testing.T) {
	var i32 int32
	handled, err := DirectAssignInt(&i32, 42, false, nil)
	if !handled || err != nil || i32 != 42 {
		t.Fatalf("int32: handled=%v err=%v val=%d", handled, err, i32)
	}

	var i int
	handled, err = DirectAssignInt(&i, -7, false, nil)
	if !handled || err != nil || i != -7 {
		t.Fatalf("int: handled=%v err=%v val=%d", handled, err, i)
	}

	var n Nullable
	handled, err = DirectAssignInt(&n, 9, false, nil)
	if !handled || err != nil || n.Null || n.Value.(int64) != 9 {
		t.Fatalf("Nullable: handled=%v err=%v val=%#v", handled, err, n)
	}

	handled, err = DirectAssignInt(&i32, 0, true, nil)
	if !handled || err != ErrScanNull {
		t.Fatalf("null into *int32: handled=%v err=%v", handled, err)
	}

	handled, err = DirectAssignInt(&n, 0, true, nil)
	if !handled || err != nil || !n.Null {
		t.Fatalf("null into Nullable: handled=%v err=%v n=%#v", handled, err, n)
	}

	var s string
	handled, err = DirectAssignInt(&s, 1, false, nil)
	if handled || err != nil {
		t.Fatalf("wrong type should not handle: handled=%v err=%v", handled, err)
	}
}

func TestDirectAssignOpt(t *testing.T) {
	var o Opt[int32]
	handled, err := DirectAssignInt(&o, 7, false, nil)
	if !handled || err != nil || !o.Valid || o.V != 7 {
		t.Fatalf("Opt set: handled=%v err=%v o=%+v", handled, err, o)
	}
	handled, err = DirectAssignInt(&o, 0, true, nil)
	if !handled || err != nil || o.Valid || o.V != 0 {
		t.Fatalf("Opt null: handled=%v err=%v o=%+v", handled, err, o)
	}

	var os Opt[string]
	handled, err = DirectAssignBytes(&os, []byte("hi"), false, true, nil)
	if !handled || err != nil || !os.Valid || os.V != "hi" {
		t.Fatalf("Opt string: handled=%v err=%v o=%+v", handled, err, os)
	}
	handled, err = DirectAssignBytes(&os, nil, true, true, nil)
	if !handled || err != nil || os.Valid {
		t.Fatalf("Opt string null: handled=%v err=%v o=%+v", handled, err, os)
	}
}

func TestDirectAssignNullFlag(t *testing.T) {
	var name string
	var nameNull bool
	sink := &NullFlagPrep{Value: &name, Null: &nameNull}

	handled, err := DirectAssignBytes(sink, []byte("alice"), false, true, nil)
	if !handled || err != nil || name != "alice" || nameNull {
		t.Fatalf("flag non-null: handled=%v err=%v name=%q null=%v", handled, err, name, nameNull)
	}

	handled, err = DirectAssignBytes(sink, nil, true, true, nil)
	if !handled || err != nil || name != "" || !nameNull {
		t.Fatalf("flag null: handled=%v err=%v name=%q null=%v", handled, err, name, nameNull)
	}

	var id int32
	var idNull bool
	sink2 := &NullFlagPrep{Value: &id, Null: &idNull}
	handled, err = DirectAssignInt(sink2, 99, false, nil)
	if !handled || err != nil || id != 99 || idNull {
		t.Fatalf("flag int: handled=%v err=%v id=%d null=%v", handled, err, id, idNull)
	}
	handled, err = DirectAssignInt(sink2, 0, true, nil)
	if !handled || err != nil || id != 0 || !idNull {
		t.Fatalf("flag int null: handled=%v err=%v id=%d null=%v", handled, err, id, idNull)
	}
}

func TestDirectAssignBytes(t *testing.T) {
	src := []byte("hello")

	var s string
	handled, err := DirectAssignBytes(&s, src, false, true, nil)
	if !handled || err != nil || s != "hello" {
		t.Fatalf("string: handled=%v err=%v val=%q", handled, err, s)
	}

	var bb []byte
	handled, err = DirectAssignBytes(&bb, src, false, true, nil)
	if !handled || err != nil || string(bb) != "hello" {
		t.Fatalf("[]byte: handled=%v err=%v val=%q", handled, err, bb)
	}
	src[0] = 'x'
	if string(bb) != "hello" {
		t.Fatalf("mustCopy did not isolate []byte: %q", bb)
	}

	var buf bytes.Buffer
	handled, err = DirectAssignBytes(&buf, []byte("w"), false, true, nil)
	if !handled || err != nil || buf.String() != "w" {
		t.Fatalf("writer: handled=%v err=%v val=%q", handled, err, buf.String())
	}
}

func TestDirectAssignWriterNoRetain(t *testing.T) {
	var buf bytes.Buffer
	// Pass concrete writer as io.Writer (as Prep does for an interface field).
	var w io.Writer = &buf
	handled, err := DirectAssignBytes(w, []byte("abc"), false, false, nil)
	if !handled || err != nil {
		t.Fatalf("handled=%v err=%v", handled, err)
	}
	if buf.String() != "abc" {
		t.Fatalf("buf=%q", buf.String())
	}
	var isNull bool
	sink := &NullFlagPrep{Value: w, Null: &isNull}
	handled, err = DirectAssignBytes(sink, nil, true, false, nil)
	if !handled || err != nil {
		t.Fatalf("null writer: handled=%v err=%v", handled, err)
	}
	if !isNull {
		t.Fatal("expected null flag")
	}
}

func TestDirectAssignBool(t *testing.T) {
	var b bool
	handled, err := DirectAssignBool(&b, true, false, nil)
	if !handled || err != nil || !b {
		t.Fatalf("bool: handled=%v err=%v val=%v", handled, err, b)
	}
}
