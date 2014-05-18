// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package rdb

// Manages the life cycle of a query.
// The result must automaticly Close() if the command Arity is Zero after
// execution or after the first Scan() if Arity is One.
type Result struct {
	conn Conn
	val  valuer
	cp   *ConnPool
}

// Results should automatically close when all rows have been read.
func (r *Result) Close() error {
	return r.close(true)
}

func (r *Result) close(explicit bool) error {
	if r == nil {
		return nil
	}
	if explicit {
		r.val.clearBuffer()
	}
	var err error
loop:
	for {
		if r.conn == nil {
			return nil
		}
		switch r.conn.Status() {
		case StatusQuery:
			err = r.scan(false)
			if err != nil {
				return err
			}
		case StatusReady:
			// Don't close the connection, just return to pool.
			err = r.cp.releaseConn(r.conn, false)
			r.cp = nil
			r.conn = nil
			break loop
		default:
			// Not sure what the state is, close the entire connection.
			err = r.cp.releaseConn(r.conn, true)
			break loop
		}
	}
	if err == nil && len(r.val.errorList) != 0 {
		err = r.val.errorList
	}
	return err
}

// Fetch the table schema.
func (r *Result) Schema() []*SqlColumn {
	return r.val.columns
}

// Informational messages. Do not call concurrently with Scan() or Done().
func (r *Result) Info() []*SqlMessage {
	return r.val.infoList
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

// Use after Scan(). Can only pull fields which have not already been sent
// into a prepared value.
func (r *Result) Get(name string) (interface{}, error) {
	col, found := r.val.columnLookup[name]
	if !found {
		return nil, ErrorColumnNotFound{At: "Get", Name: name}
	}
	bv := r.val.buffer[col.Index]
	return bv.V, nil
}

// Use after Scan(). Can only pull fields which have not already been sent
// into a prepared value.
func (r *Result) Getx(index int) (interface{}, error) {
	if index < 0 || index >= len(r.val.columns) {
		return nil, ErrorColumnNotFound{At: "Getx", Index: index}
	}
	bv := r.val.buffer[index]
	return bv.V, nil
}

// Use after Scan(). Can only pull fields which have not already been sent
// into a prepared value.
func (r *Result) GetN(name string) (Nullable, error) {
	col, found := r.val.columnLookup[name]
	if !found {
		return Nullable{}, ErrorColumnNotFound{At: "GetN", Name: name}
	}
	return r.val.buffer[col.Index], nil
}

// Use after Scan(). Can only pull fields which have not already been sent
// into a prepared value.
func (r *Result) GetxN(index int) (Nullable, error) {
	if index < 0 || index >= len(r.val.columns) {
		return Nullable{}, ErrorColumnNotFound{At: "GetxN", Index: index}
	}
	return r.val.buffer[index], nil
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
	return !r.val.eof
}

// Scans the row into a buffer that can be fetched with Get and scans
// directly into any prepared values.
// Return value "more" is false if no more rows.
// Results should automatically close when all rows have been read.
func (r *Result) Scan(values ...interface{}) error {
	for i := range values {
		if i >= len(r.val.columns) {
			return ErrorColumnNotFound{At: "Scan(values)", Index: i}
		}
		r.val.prep[i] = values[i]
	}
	return r.scan(true)
}

func (r *Result) scan(reportRow bool) error {
	if reportRow {
		r.val.clearBuffer()
	}
	err := r.conn.Scan(reportRow)
	r.val.clearPrep()

	// Only show SQL errors if no connection errors,
	// but show before any other errors.
	if err == nil && len(r.val.errorList) != 0 {
		err = r.val.errorList
	}

	if r.val.arity&One != 0 {
		r.val.eof = true
		if r.val.rowCount == 1 {
			serr := r.conn.Scan(false)
			if err == nil {
				err = serr
			}
		}
		if err == nil && r.val.arity&ArityMust != 0 && r.val.rowCount > 1 {
			err = arityError
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

// Get the panic'ing version that doesn't return errors.
func (r *Result) Must() ResultMust {
	return ResultMust{norm: r}
}
