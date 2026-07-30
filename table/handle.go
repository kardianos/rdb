// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package table

import (
	"context"
	"errors"
	"iter"
	"unsafe"

	"github.com/kardianos/rdb"
)

// ErrNoMoreResults is returned by Handle.Next when the server has no further
// result sets.
var ErrNoMoreResults = errors.New("table: no more result sets")

// ErrNilHandle is returned when Slice/Seq are called with a zero Handle.
var ErrNilHandle = errors.New("table: nil handle")

// Handle is a thin cursor over an open *rdb.Result. It does not own the result:
// the caller (typically Do) must Close the result when finished.
//
// Consume the current set with Slice or Seq, then call Next to advance to the
// next result set. Sets must be consumed (or Drain'ed) in order; do not run two
// Seq iterators at once on the same handle.
type Handle struct {
	res *rdb.Result
}

// NewHandle wraps an open result. Prefer Do when you want Query+Close managed.
func NewHandle(res *rdb.Result) Handle {
	return Handle{res: res}
}

// Result returns the underlying query result, or nil for a zero Handle.
func (h Handle) Result() *rdb.Result {
	return h.res
}

// Next advances to the next result set. It returns ErrNoMoreResults when there
// are no further sets. Finish or Drain the current set before calling Next.
func (h Handle) Next() error {
	if h.res == nil {
		return ErrNilHandle
	}
	more, err := h.res.NextResult()
	if err != nil {
		return err
	}
	if !more {
		return ErrNoMoreResults
	}
	return nil
}

// Drain discards any remaining rows of the current result set so Next is safe
// after a partial Seq or an abandoned row loop. It does not advance sets.
func (h Handle) Drain() error {
	if h.res == nil {
		return ErrNilHandle
	}
	for h.res.Next() {
		if err := h.res.Scan(); err != nil {
			return err
		}
	}
	return nil
}

// Do runs cmd on q, passes a Handle to fn, then closes the result.
// fn should consume one or more result sets via Slice/Seq and Handle.Next.
//
//	err := table.Do(ctx, db, cmd, func(h table.Handle) error {
//	    users, err := table.Slice[User](h)
//	    if err != nil {
//	        return err
//	    }
//	    if err := h.Next(); err != nil {
//	        return err
//	    }
//	    for o, err := range table.Seq[Order](h) {
//	        if err != nil {
//	            return err
//	        }
//	        // ...
//	    }
//	    return nil
//	}, params...)
func Do(ctx context.Context, q rdb.Queryer, cmd *rdb.Command, fn func(Handle) error, params ...rdb.Param) error {
	if fn == nil {
		return errors.New("table: Do fn is nil")
	}
	res, err := q.Query(ctx, cmd, params...)
	if err != nil {
		return err
	}
	defer res.Close()
	return fn(NewHandle(res))
}

// Slice drains the current result set into []T. It does not advance to the next
// set and does not Close the result. Column mapping uses the "db" tag.
func Slice[T any](h Handle) ([]T, error) {
	return SliceTag[T](h, "db")
}

// SliceTag is Slice with an explicit struct tag name (empty means "db").
func SliceTag[T any](h Handle, tagName string) ([]T, error) {
	if h.res == nil {
		return nil, ErrNilHandle
	}
	plan, err := newStructPlan[T](h.res.Schema(), tagName)
	if err != nil {
		return nil, err
	}
	var out []T
	for h.res.Next() {
		out = append(out, *new(T))
		if err := plan.scan(h.res, unsafe.Pointer(&out[len(out)-1])); err != nil {
			return out, err
		}
	}
	return out, nil
}

// Seq streams the current result set as T. It does not Close the result and does
// not call NextResult. When iteration ends (including early break or error),
// remaining rows of the current set are drained so Handle.Next is safe afterward.
//
//	for row, err := range table.Seq[MyRow](h) {
//	    if err != nil {
//	        return err
//	    }
//	    // use row
//	}
func Seq[T any](h Handle) iter.Seq2[T, error] {
	return SeqTag[T](h, "db")
}

// SeqTag is Seq with an explicit struct tag name (empty means "db").
func SeqTag[T any](h Handle, tagName string) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		var zero T
		if h.res == nil {
			yield(zero, ErrNilHandle)
			return
		}
		plan, err := newStructPlan[T](h.res.Schema(), tagName)
		if err != nil {
			yield(zero, err)
			return
		}
		// Always drain so a partial range leaves the set boundary clean for Next.
		defer func() { _ = h.Drain() }()

		for h.res.Next() {
			var row T
			if err := plan.scan(h.res, unsafe.Pointer(&row)); err != nil {
				yield(zero, err)
				return
			}
			if !yield(row, nil) {
				return
			}
		}
	}
}
