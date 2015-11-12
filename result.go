// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package rdb

import (
	"sync"
	"time"
)

// Manages the life cycle of a query.
// The result must automaticly Close() if the command Arity is Zero after
// execution or after the first Scan() if Arity is One.
type Result struct {
	conn DriverConn
	val  valuer
	cp   *ConnPool

	// If true do not return to the connection pool when closed.
	keepOnClose bool

	m       sync.RWMutex
	lastHit time.Time
	closing chan struct{}
	closed  bool
}

// Results should automatically close when all rows have been read.
func (r *Result) Close() error {
	if r == nil {
		return nil
	}
	return r.close(true)
}
func (r *Result) updateHit() {
	r.m.Lock()
	r.lastHit = time.Now()
	r.m.Unlock()
}

func (r *Result) autoClose(after time.Duration) {
	if r == nil {
		return
	}

	r.lastHit = time.Now()
	go func() {
		<-time.After(after)
		tick := time.NewTicker(time.Millisecond * 100)
		defer tick.Stop()
		for {
			select {
			case now := <-tick.C:
				r.m.RLock()
				if now.Sub(r.lastHit) > after {
					// Place notification in RLock and before r.Close
					// to prevent r.cp from getting altered.
					if r.cp != nil && r.cp.OnAutoClose != nil {
						go r.cp.OnAutoClose(r.val.cmd.Sql)
					}
					r.m.RUnlock()
					r.Close()
					return
				}
				r.m.RUnlock()
			case <-r.closing:
				return
			}
		}
	}()
}

func (r *Result) RowsAffected() uint64 {
	return r.val.rowsAffected
}

func (r *Result) close(explicit bool) error {
	if r == nil {
		return nil
	}
	r.closing <- struct{}{}
	r.m.Lock()
	if r.closed {
		r.m.Unlock()
		return nil
	}
	r.closed = true
	r.m.Unlock()

	if explicit {
		r.val.clearBuffer()
	}
	var err error

	if r.conn == nil {
		return nil
	}

	err = r.conn.NextQuery()
	if err != nil {
		r.cp.releaseConn(r.conn, true)
		return err
	}

	if r.keepOnClose == false {
		err = r.cp.releaseConn(r.conn, false)
		r.cp = nil
		r.conn = nil
	}
	if err == nil && len(r.val.errorList) != 0 {
		err = r.val.errorList
	}
	return err
}

// Fetch the table schema.
func (r *Result) Schema() []*Column {
	return r.val.columns
}

// Informational messages. Do not call concurrently with Scan() or Done().
func (r *Result) Info() []*Message {
	return r.val.infoList
}

// Prepare pointers to values to be populated by name using Prep. After
// preparing call Scan(). Will panic if name is not a valid column name.
func (r *Result) Prep(name string, value interface{}) *Result {
	col, found := r.val.columnLookup[name]
	if !found {
		panic(ErrorColumnNotFound{At: "Prep", Name: name})
	}
	r.val.prep[col.Index] = value
	return r
}

// Prepare pointers to values to be populated by index using Prep. After
// preparing call Scan(). Will panic if index is not a valid column index.
func (r *Result) Prepx(index int, value interface{}) *Result {
	if index < 0 || index >= len(r.val.columns) {
		panic(ErrorColumnNotFound{At: "Prepx", Index: index})
	}
	r.val.prep[index] = value
	return r
}

// Use after Scan(). Can only pull fields which have not already been sent
// into a prepared value. Will panic if name is not a valid column name.
func (r *Result) Get(name string) interface{} {
	col, found := r.val.columnLookup[name]
	if !found {
		panic(ErrorColumnNotFound{At: "Get", Name: name})
	}
	bv := r.val.buffer[col.Index]
	return bv.Value
}

// Use after Scan(). Can only pull fields which have not already been sent
// into a prepared value. Will panic if index is not a valid column index.
func (r *Result) Getx(index int) interface{} {
	if index < 0 || index >= len(r.val.columns) {
		panic(ErrorColumnNotFound{At: "Getx", Index: index})
	}
	bv := r.val.buffer[index]
	return bv.Value
}

// Use after Scan(). Can only pull fields which have not already been sent
// into a prepared value. Will panic if name is not a valid column name.
func (r *Result) GetN(name string) Nullable {
	col, found := r.val.columnLookup[name]
	if !found {
		panic(ErrorColumnNotFound{At: "GetN", Name: name})
	}
	return r.val.buffer[col.Index]
}

// Use after Scan(). Can only pull fields which have not already been sent
// into a prepared value. Will panic if index is not a valid column index.
func (r *Result) GetxN(index int) Nullable {
	if index < 0 || index >= len(r.val.columns) {
		panic(ErrorColumnNotFound{At: "GetxN", Index: index})
	}
	return r.val.buffer[index]
}

// Use after Scan(). Can only pull fields which have not already been sent
// into a prepared value. Not all fields will be populated if some have
// been prepared.
func (r *Result) GetRowN() []Nullable {
	rowBuf := r.val.buffer
	ret := make([]Nullable, len(rowBuf))
	for i := range rowBuf {
		ret[i] = rowBuf[i]
	}
	return ret
}

// Optional to call. Determine if there is another row.
// Scan actually advances to the next row.
func (r *Result) Next() (more bool) {
	if r.conn == nil {
		return false
	}
	if r.conn.Status() == StatusResultDone {
		return false
	}
	return !r.val.eof
}

func (r *Result) NextResult() (more bool, err error) {
	if r.conn == nil {
		return false, nil
	}
	r.updateHit()
	return r.conn.NextResult()
}

// Scans the row into a buffer that can be fetched with Get and scans
// directly into any prepared values.
// Return value "more" is false if no more rows.
// Results should automatically close when all rows have been read.
func (r *Result) Scan(values ...interface{}) error {
	r.updateHit()
	for i := range values {
		if i >= len(r.val.columns) {
			return ErrorColumnNotFound{At: "Scan(values)", Index: i}
		}
		r.val.prep[i] = values[i]
	}

	r.val.clearBuffer()
	err := r.conn.Scan()
	r.val.clearPrep()

	// Only show SQL errors if no connection errors,
	// but show before any other errors.
	if err == nil && len(r.val.errorList) != 0 {
		err = r.val.errorList
	}

	if r.val.cmd.Arity&One != 0 {
		r.val.eof = true
		if r.val.rowCount == 1 {
			serr := r.conn.NextQuery()
			if err == nil {
				err = serr
			}
		}
		if err == nil && r.val.cmd.Arity&ArityMust != 0 && r.val.rowCount > 1 {
			err = ArityError
		}
	}

	if r.val.eof {
		cerr := r.close(false)
		if err == nil {
			err = cerr
		}
	}
	return err
}
