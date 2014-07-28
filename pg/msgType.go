// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the rdb LICENSE file.

package pg

import (
	"bytes"
	"io"
)

type MsgAuthenticationOk struct{}
type MsgAuthenticationKerberosV5 struct{}
type MsgAuthenticationCleartextPassword struct{}
type MsgAuthenticationMD5Password struct{ Salt []byte }
type MsgAuthenticationSCMCredential struct{}
type MsgAuthenticationGSS struct{}
type MsgAuthenticationSSPI struct{}
type MsgAuthenticationGSSContinue struct{ Data []byte }

type MsgBackendKeyData struct {
	ProcessID int32
	SecretKey int32
}
type MsgBindComplete struct{}
type MsgCloseComplete struct{}
type MsgCommandComplete struct{ Command string }

type LimitedReader struct {
	r     io.Reader
	limit int
}

func (r *LimitedReader) Read(buf []byte) (n int, err error) {
	if r.limit < len(buf) {
		buf = buf[:r.limit]
	}
	n, err = r.r.Read(buf)
	r.limit -= n
	if r.limit == 0 {
		err = io.EOF
	}
	return n, err
}

// Must read entire reader until empty.
type MsgCopyData struct{ Reader *LimitedReader }
type MsgCopyDone struct{}
type MsgCopyFail struct{ Message string }
type MsgCopyInResponse struct {
	Format       int8
	ColumnCount  int16
	ColumnFormat []int16
}
type MsgCopyOutResponse struct {
	Format       int8
	ColumnCount  int16
	ColumnFormat []int16
}
type MsgCopyBothResponse struct {
	Format       int8
	ColumnCount  int16
	ColumnFormat []int16
}
type MsgDataRow struct {
	ColumnCount int16
	// For each column:
	//	int32, bytea
	//  Value is null if int32 == -1.
}
type MsgEmptyQueryResponse struct{}
type MsgErrorResponse struct {
	Messages []*StatusMessage
}

func (msg MsgErrorResponse) Error() string {
	buf := &bytes.Buffer{}
	for i, m := range msg.Messages {
		if i != 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(m.String())
	}
	return buf.String()
}

type MsgFunctionCallResponse struct {
	Length int32 // If -1 Value is not present and NULL.
	Value  []byte
}
type MsgNoData struct{}
type MsgNoticeResponse struct {
	FieldType byte
	Message   string
}
type MsgNotificationResponse struct {
	ProcessID   int32
	ChannelName string
	Payload     string
}
type MsgParameterDescription struct {
	Count    int16
	ObjectID []int32
}
type MsgParameterStatus struct{}
type MsgParseComplete struct{}
type MsgPortalSuspended struct{}
type MsgReadyForQuery struct {
	TransactionStatus transactionStatus
}
type MsgRowDescription struct {
	// Count int16
	// TODO: Each column.
}
