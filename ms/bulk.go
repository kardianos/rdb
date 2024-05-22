package ms

import (
	"fmt"
	"io"
	"strings"

	"github.com/kardianos/rdb"
)

type Bulk struct {
	CheckConstraints bool
	FireTriggers     bool
	KeepNulls        bool
	TabLock          bool
	KBPerBatch       int
	RowsPerBatch     int

	TableName string
	Columns   []rdb.Param

	Row func(row []rdb.Param) error
}

var _ rdb.Bulk = &Bulk{}

func (b *Bulk) Start() (sql string, col []rdb.Param, err error) {
	if len(b.TableName) == 0 {
		return "", nil, fmt.Errorf("missing table name for bulk insert")
	}
	if len(b.Columns) == 0 {
		return "", nil, fmt.Errorf("missing columns name for bulk insert")
	}
	buf := &strings.Builder{}
	buf.WriteString("insert bulk ")
	buf.WriteString(b.TableName)
	buf.WriteString(" (\n")
	for i, c := range b.Columns {
		if i > 0 {
			buf.WriteString(",\n")
		}
		tw, found := sqlTypeLookup[c.Type]
		if !found {
			return "", b.Columns, fmt.Errorf("sql type not setup: %d", c.Type)
		}
		ts := tw.TypeString(&c)
		buf.WriteRune('\t')
		buf.WriteString(c.Name)
		buf.WriteRune(' ')
		buf.WriteString(ts)
	}
	buf.WriteString("\n)")
	var withCount int
	writeWith := func(s string) {
		if withCount == 0 {
			buf.WriteString(" with (\n")
		} else {
			buf.WriteString(",\n")
		}
		withCount++
		buf.WriteRune('\t')
		buf.WriteString(s)
	}
	if b.CheckConstraints {
		writeWith("CHECK_CONSTRAINTS")
	}
	if b.FireTriggers {
		writeWith("FIRE_TRIGGERS")
	}
	if b.KeepNulls {
		writeWith("KEEP_NULLS")
	}
	if b.TabLock {
		writeWith("TABLOCK")
	}
	if b.KBPerBatch > 0 {
		writeWith(fmt.Sprintf("KILOBYTES_PER_BATCH=%d", b.KBPerBatch))
	}
	if b.RowsPerBatch > 0 {
		writeWith(fmt.Sprintf("ROWS_PER_BATCH=%d", b.RowsPerBatch))
	}
	if withCount > 0 {
		buf.WriteString("\n)")
	}

	buf.WriteString(";")
	ret := buf.String()
	return ret, b.Columns, nil
}
func (b *Bulk) Next(batchCount int, row []rdb.Param) error {
	if b == nil || b.Row == nil {
		return io.EOF
	}
	if b.RowsPerBatch > 0 && batchCount >= b.RowsPerBatch {
		return rdb.ErrBulkBatchDone
	}
	return b.Row(row)
}
