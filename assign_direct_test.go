// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package rdb

import (
	"bytes"
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

func TestDirectAssignBool(t *testing.T) {
	var b bool
	handled, err := DirectAssignBool(&b, true, false, nil)
	if !handled || err != nil || !b {
		t.Fatalf("bool: handled=%v err=%v val=%v", handled, err, b)
	}
}
