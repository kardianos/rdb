// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package rdb

import "fmt"

const debugConnectionReuse = false

// Manages the life cycle of a query.
// The result must automaticly Close() if the command Arity is Zero after
// execution or after the first Scan() if Arity is One.
type Result struct {
	conn Conn
	val  valuer
	cp   *ConnPool
}

// Results should automatically close when all rows have been read.
func (res *Result) Close() error {
	if debugConnectionReuse {
		fmt.Println("Result.Close() start")
	}
	for {
		if res.conn == nil {
			return nil
		}
		switch res.conn.Status() {
		case StatusQuery:
			_, err := res.Scan()
			if err != nil {
				return err
			}
		case StatusReady:
			if debugConnectionReuse {
				fmt.Println("Result.Close() REUSE")
			}
			// Don't close the connection, just return to pool.
			res.cp.releaseConn(res.conn, false)
			res.cp = nil
			res.conn = nil
		default:
			break
		}
	}

	if debugConnectionReuse {
		fmt.Println("Result.Close() CLOSE")
	}
	// Not sure what the state is, close the entire connection.

	return res.cp.releaseConn(res.conn, true)
}

// Fetch the table schema.
func (res *Result) Schema() []*SqlColumn {
	return res.val.columns
}

// Prepare pointers to values to be populated by name using Prep. After
// preparing call Scan().
func (r *Result) Prep(name string, value interface{}) error {
	col, found := r.val.columnLookup[name]
	if !found {
		return ErrorColumnNotFound{At: "Prep", Name: name}
	}
	r.val.prep[col.Index] = value
	return nil
}

// Prepare pointers to values to be populated by index using Prep. After
// preparing call Scan().
func (r *Result) Prepx(index int, value interface{}) error {
	if index < 0 || index >= len(r.val.columns) {
		return ErrorColumnNotFound{At: "Prepx", Index: index}
	}
	r.val.prep[index] = value
	return nil
}

// Prepare pointers to values to be populated by index using Prep. After
// preparing call Scan().
func (r *Result) PrepAll(values ...interface{}) error {
	for i := range values {
		if i >= len(r.val.columns) {
			return ErrorColumnNotFound{At: "PrepAll", Index: i}
		}
		r.val.prep[i] = values[i]
	}
	return nil
}

// Use after Scan(). Can only pull fields which have not already been sent
// into a prepared value.
func (r *Result) Get(name string) (interface{}, error) {
	col, found := r.val.columnLookup[name]
	if !found {
		return nil, ErrorColumnNotFound{At: "Get", Name: name}
	}
	bv := r.val.buffer[col.Index]
	if bv == nil {
		return nil, nil
	}
	return bv.Value, nil
}

// Use after Scan(). Can only pull fields which have not already been sent
// into a prepared value.
func (r *Result) Getx(index int) (interface{}, error) {
	if index < 0 || index >= len(r.val.columns) {
		return nil, ErrorColumnNotFound{At: "Getx", Index: index}
	}
	bv := r.val.buffer[index]
	if bv == nil {
		return nil, nil
	}
	return bv.Value, nil
}

// Use after Scan(). Can only pull fields which have not already been sent
// into a prepared value.
func (r *Result) GetN(name string) (Nullable, error) {
	col, found := r.val.columnLookup[name]
	if !found {
		return Nullable{}, ErrorColumnNotFound{At: "GetN", Name: name}
	}
	bv := r.val.buffer[col.Index]
	return Nullable{
		Null: bv.Null,
		V:    bv.Value,
	}, nil
}

// Use after Scan(). Can only pull fields which have not already been sent
// into a prepared value.
func (r *Result) GetxN(index int) (Nullable, error) {
	if index < 0 || index >= len(r.val.columns) {
		return Nullable{}, ErrorColumnNotFound{At: "GetxN", Index: index}
	}
	bv := r.val.buffer[index]
	return Nullable{
		Null: bv.Null,
		V:    bv.Value,
	}, nil
}

// Scans the row into a buffer that can be fetched with Get and scans
// directly into any prepared values.
// Return value "more" is false if no more rows.
// Results should automatically close when all rows have been read.
func (res *Result) Scan() (more bool, err error) {
	res.val.clearBuffer()
	err = res.conn.Scan()
	if res.val.eof {
		cerr := res.Close()
		if err == nil {
			err = cerr
		}
		return false, err
	}
	return true, err
}

// Get the panic'ing version that doesn't return errors.
func (res *Result) Must() ResultMust {
	return ResultMust{norm: res}
}
