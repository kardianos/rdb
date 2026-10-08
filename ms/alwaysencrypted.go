// Copyright 2026 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package ms

// SQL Server Always Encrypted, with the column encryption keys given in the
// config (rdb.Config.ColumnKeys, DSN column_key=) rather than unwrapped from
// a key store. With keys set the login asks for column encryption; the server
// then describes each encrypted result column and the key it uses, and the
// driver decrypts the values and reports the plaintext type. Parameters are
// described by sp_describe_parameter_encryption before the query runs, and
// those the server says are compared with or stored in an encrypted column
// are encrypted.
//
// Decrypted and encrypted value types: char, varchar, nchar, nvarchar
// (including max), binary, varbinary, and the integer types. Others are an
// error, as are encrypted output parameters.

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"

	"github.com/kardianos/rdb"
	"github.com/kardianos/rdb/internal/uconv"
	"github.com/kardianos/rdb/ms/aecrypt"
)

const (
	featureIDColumnEncryption byte = 0x04

	aeVersion        = 1 // Column encryption without enclave computations.
	aeAlgorithmAEAD  = 2 // AEAD_AES_256_CBC_HMAC_SHA256; 0 names a custom algorithm.
	aeDeterministic  = 1
	aeRandomized     = 2
	aeNormalization  = 1
	paramStatusOut   = 0x01
	paramStatusCrypt = 0x08 // RPC parameter status: the value is encrypted.

	maxShortBinary = 8000
)

const rotateHint = "after a key rotation, run ms/aecrypt/extract_cek.ps1 again for a new column_key"

// parseColumnKeys parses the text form of each configured column encryption key.
func parseColumnKeys(list []string) ([]*aecrypt.Key, error) {
	keys := make([]*aecrypt.Key, 0, len(list))
	for i, s := range list {
		k, err := aecrypt.ParseKey(s)
		if err != nil {
			return nil, fmt.Errorf("column_key[%d]: %w", i, err)
		}
		keys = append(keys, k)
	}
	return keys, nil
}

// setupColumnEncryption checks the server took column encryption when keys
// are configured.
func (tds *Connection) setupColumnEncryption(si *ServerInfo) error {
	if len(tds.columnKeys) == 0 {
		return nil
	}
	if !si.ColumnEncryption {
		return fmt.Errorf("column_key is set but server %s does not support Always Encrypted", si)
	}
	tds.ae = true
	// Encrypted parameters must keep their declared type; never remap
	// nvarchar to varchar.
	tds.preferUTF8Varchar = false
	return nil
}

// cekValue is one stored value of a column encryption key: the key wrapped
// by one column master key.
type cekValue struct {
	wrapped []byte
	keyPath string
}

// cekEntry is a column encryption key the server names, in the key table of
// column metadata or in a parameter description.
type cekEntry struct {
	dbID, keyID, keyVersion uint32
	mdVersion               [8]byte
	values                  []cekValue

	key *aecrypt.Key // The configured key one of values is the stored form of; nil if none.
}

func (e *cekEntry) masterKeys() string {
	paths := make([]string, len(e.values))
	for i, v := range e.values {
		paths[i] = v.keyPath
	}
	return strings.Join(paths, ", ")
}

// matchKey returns the configured key whose fingerprint matches one of the
// stored values.
func (tds *Connection) matchKey(values []cekValue) *aecrypt.Key {
	for _, k := range tds.columnKeys {
		for _, v := range values {
			if k.Matches(v.wrapped) {
				return k
			}
		}
	}
	return nil
}

// readCekTable reads the key table at the start of column metadata.
func (tds *Connection) readCekTable(read uconv.PanicReader) []*cekEntry {
	list := make([]*cekEntry, binary.LittleEndian.Uint16(read(2)))
	for i := range list {
		e := &cekEntry{
			dbID:       binary.LittleEndian.Uint32(read(4)),
			keyID:      binary.LittleEndian.Uint32(read(4)),
			keyVersion: binary.LittleEndian.Uint32(read(4)),
		}
		copy(e.mdVersion[:], read(8))
		e.values = make([]cekValue, read(1)[0])
		for j := range e.values {
			n := int(binary.LittleEndian.Uint16(read(2)))
			e.values[j].wrapped = bytes.Clone(read(n))
			uconv.Decode.Prefix1(read) // Key store provider name.
			_, e.values[j].keyPath = uconv.Decode.Prefix2(read)
			uconv.Decode.Prefix1(read) // Key encryption algorithm, RSA_OAEP.
		}
		e.key = tds.matchKey(e.values)
		list[i] = e
	}
	return list
}

// encColumn is the encryption of one result column.
type encColumn struct {
	wire    *SQLColumn // The ciphertext as sent: varbinary.
	ordinal int
	cek     *cekEntry
	alg     byte
	encType byte
	norm    byte
}

// decodeCryptoMetadata reads the crypto metadata that follows the TYPE_INFO
// of encrypted column wire. It returns the column of the plaintext type.
func decodeCryptoMetadata(read uconv.PanicReader, wire *SQLColumn, ceks []*cekEntry) *SQLColumn {
	enc := &encColumn{
		wire:    wire,
		ordinal: int(binary.LittleEndian.Uint16(read(2))),
	}
	userType := binary.LittleEndian.Uint32(read(4))
	col := decodeTypeInfo(read, userType, colFlags{})
	enc.alg = read(1)[0]
	if enc.alg == 0 {
		uconv.Decode.Prefix1(read) // Custom algorithm name.
	}
	enc.encType = read(1)[0]
	enc.norm = read(1)[0]
	if enc.ordinal < len(ceks) {
		enc.cek = ceks[enc.ordinal]
	}
	col.Nullable = wire.Nullable
	col.Key = wire.Key
	col.Serial = wire.Serial
	col.enc = enc
	return col
}

// check reports why the values of col cannot be decrypted.
func (enc *encColumn) check(col *SQLColumn) error {
	switch {
	case enc.cek == nil:
		return fmt.Errorf("encrypted column %q: key ordinal %d is not in the key table", col.Name, enc.ordinal)
	case enc.alg != aeAlgorithmAEAD:
		return fmt.Errorf("encrypted column %q: unsupported encryption algorithm %d", col.Name, enc.alg)
	case enc.encType != aeDeterministic && enc.encType != aeRandomized:
		return fmt.Errorf("encrypted column %q: unsupported encryption type %d", col.Name, enc.encType)
	case enc.norm != aeNormalization:
		return fmt.Errorf("encrypted column %q: unsupported normalization version %d", col.Name, enc.norm)
	case enc.wire.code != typeVarBinary:
		return fmt.Errorf("encrypted column %q: ciphertext sent as %s, want varbinary", col.Name, enc.wire.info.Name)
	case !aeValueType(col.code, col.info):
		return fmt.Errorf("encrypted column %q: decrypting %s values is not supported", col.Name, col.info.Name)
	case enc.cek.key == nil:
		return fmt.Errorf("encrypted column %q: no column_key matches its column encryption key (column master key %s); %s", col.Name, enc.cek.masterKeys(), rotateHint)
	}
	return nil
}

// aeValueType reports whether values of the type can be encrypted and decrypted.
func aeValueType(code driverType, info typeInfo) bool {
	return (info.Bytes && !info.Table) || code == typeIntN
}

// decodeEncryptedValue reads one ciphertext of column, decrypts it, and
// decodes the plaintext as the column's type.
func (tds *Connection) decodeEncryptedValue(read uconv.PanicReader, column *SQLColumn, resultWf writeField, reportRow bool) {
	value, null := readCipherValue(read, column.enc.wire)
	if !reportRow {
		return
	}
	var plain []byte
	if !null {
		var err error
		plain, err = column.enc.cek.key.Decrypt(value)
		if err != nil {
			panic(recoverError{fmt.Errorf("encrypted column %q: %w", column.Name, err)})
		}
	}
	stream, err := plainStream(column, plain, null)
	if err != nil {
		panic(recoverError{err})
	}
	tds.decodePlainValue(sliceReader(stream), column, resultWf, reportRow)
}

// readCipherValue reads one varbinary value. The value may view the message
// buffer; use it before the next read.
func readCipherValue(read uconv.PanicReader, wire *SQLColumn) (value []byte, null bool) {
	if !wire.Unlimit {
		n := int(binary.LittleEndian.Uint16(read(2)))
		if n == 0xFFFF {
			return nil, true
		}
		return read(n), false
	}
	total := binary.LittleEndian.Uint64(read(8))
	if total == textNULL {
		return nil, true
	}
	out := []byte{}
	for {
		n := int(binary.LittleEndian.Uint32(read(4)))
		if n == 0 {
			return out, false
		}
		out = append(out, read(n)...)
	}
}

// plainStream frames a decrypted value as the wire would send the
// column's plaintext type, for decodePlainValue.
func plainStream(column *SQLColumn, plain []byte, null bool) ([]byte, error) {
	info := column.info
	if info.NChar && len(plain)%2 != 0 {
		return nil, fmt.Errorf("encrypted column %q: decrypted %d bytes, not UTF-16", column.Name, len(plain))
	}
	if column.Unlimit {
		if null {
			return binary.LittleEndian.AppendUint64(nil, textNULL), nil
		}
		b := make([]byte, 0, 8+4+len(plain)+4)
		b = binary.LittleEndian.AppendUint64(b, uint64(len(plain)))
		if len(plain) > 0 {
			b = binary.LittleEndian.AppendUint32(b, uint32(len(plain)))
			b = append(b, plain...)
		}
		return binary.LittleEndian.AppendUint32(b, 0), nil
	}
	if column.code == typeIntN {
		if null {
			return []byte{0}, nil
		}
		// Integers are encrypted as 8 bytes, little endian.
		n := column.Length
		if len(plain) < n {
			return nil, fmt.Errorf("encrypted column %q: decrypted %d bytes for a %d byte integer", column.Name, len(plain), n)
		}
		return append([]byte{byte(n)}, plain[:n]...), nil
	}
	switch info.Len {
	case 2:
		if null {
			return []byte{0xFF, 0xFF}, nil
		}
		if len(plain) >= 0xFFFF {
			return nil, fmt.Errorf("encrypted column %q: decrypted %d bytes, too long", column.Name, len(plain))
		}
		return append(binary.LittleEndian.AppendUint16(nil, uint16(len(plain))), plain...), nil
	case 1:
		if null {
			return []byte{0xFF}, nil
		}
		if len(plain) >= 0xFF {
			return nil, fmt.Errorf("encrypted column %q: decrypted %d bytes, too long", column.Name, len(plain))
		}
		return append([]byte{byte(len(plain))}, plain...), nil
	}
	return nil, fmt.Errorf("encrypted column %q: decrypting %s values is not supported", column.Name, info.Name)
}

// sliceReader reads b in order, for decoding a value held in memory.
func sliceReader(b []byte) uconv.PanicReader {
	return func(n int) []byte {
		if n > len(b) {
			panic(recoverError{fmt.Errorf("decrypted value: want %d more bytes, have %d", n, len(b))})
		}
		v := b[:n:n]
		b = b[n:]
		return v
	}
}

// paramEnc is how one RPC parameter is encrypted.
type paramEnc struct {
	cek     *cekEntry
	alg     byte
	encType byte
	norm    byte
}

// describeParamEncryption asks the server which parameters of sql must be
// encrypted. decl declares the parameters, as sp_executesql gets them. The
// result maps "@name" to its encryption; it is empty if none are.
func (tds *Connection) describeParamEncryption(ctx context.Context, sql, decl string, reset bool) (map[string]*paramEnc, error) {
	collect := &tableValuer{}
	saved := tds.val
	tds.val = collect
	defer func() {
		tds.val = saved
		tds.col = nil
	}()

	err := tds.sendRPC(ctx, "sp_describe_parameter_encryption", false, []rdb.Param{
		{Name: "tsql", Type: rdb.Text, Value: sql},
		{Name: "params", Type: rdb.Text, Value: decl},
	}, reset, nil)
	if err != nil {
		return nil, err
	}
	// Mark the response's last packet read, so the next request's reader
	// starts at the next packet.
	defer tds.mr.Close()
	var errList rdb.Errors
	for done := false; !done; {
		var res interface{}
		withLock(&tds.syncClose, func() {
			res, err = tds.getSingleResponse(ctx, tds.mr, true)
		})
		if err != nil {
			tds.resetNext = true
			return nil, fmt.Errorf("describe parameter encryption: %w", err)
		}
		switch v := res.(type) {
		case MsgEom:
			done = true
		case *rdb.Message:
			if v.Type == rdb.SqlError {
				errList = append(errList, v)
			}
		case MsgRow:
			collect.RowScanned()
		case MsgCancel:
			if v.IsAttention {
				return nil, rdb.ErrCancel
			}
		}
	}
	if len(errList) > 0 {
		return nil, errList
	}
	return tds.paramEncryption(collect.sets)
}

// paramEncryption reads the result sets of sp_describe_parameter_encryption:
// the keys, then for each parameter how it is encrypted.
func (tds *Connection) paramEncryption(sets [][][]interface{}) (map[string]*paramEnc, error) {
	if len(sets) < 2 || len(sets[0]) == 0 {
		return nil, nil
	}
	ceks := map[int64]*cekEntry{}
	for _, row := range sets[0] {
		// column_encryption_key_ordinal, database_id, column_encryption_key_id,
		// column_encryption_key_version, column_encryption_key_metadata_version,
		// column_encryption_key_encrypted_value, column_master_key_store_provider_name,
		// column_master_key_path, column_encryption_key_encryption_algorithm_name.
		if len(row) < 9 {
			return nil, fmt.Errorf("describe parameter encryption: key row has %d columns, want 9", len(row))
		}
		ord := anyInt(row[0])
		e := ceks[ord]
		if e == nil {
			e = &cekEntry{
				dbID:       uint32(anyInt(row[1])),
				keyID:      uint32(anyInt(row[2])),
				keyVersion: uint32(anyInt(row[3])),
			}
			copy(e.mdVersion[:], anyBytes(row[4]))
			ceks[ord] = e
		}
		e.values = append(e.values, cekValue{wrapped: anyBytes(row[5]), keyPath: string(anyBytes(row[7]))})
	}
	for _, e := range ceks {
		e.key = tds.matchKey(e.values)
	}
	out := map[string]*paramEnc{}
	for _, row := range sets[1] {
		// parameter_ordinal, parameter_name, column_encryption_algorithm,
		// column_encryption_type, column_encryption_key_ordinal,
		// column_encryption_normalization_rule_version.
		if len(row) < 6 {
			return nil, fmt.Errorf("describe parameter encryption: parameter row has %d columns, want 6", len(row))
		}
		name := string(anyBytes(row[1]))
		encType := byte(anyInt(row[3]))
		if encType == 0 {
			continue // Plaintext.
		}
		pe := &paramEnc{
			cek:     ceks[anyInt(row[4])],
			alg:     byte(anyInt(row[2])),
			encType: encType,
			norm:    byte(anyInt(row[5])),
		}
		switch {
		case pe.cek == nil:
			return nil, fmt.Errorf("parameter %s: encryption key ordinal %d not described", name, anyInt(row[4]))
		case pe.alg != aeAlgorithmAEAD:
			return nil, fmt.Errorf("parameter %s: unsupported encryption algorithm %d", name, pe.alg)
		case pe.norm != aeNormalization:
			return nil, fmt.Errorf("parameter %s: unsupported normalization version %d", name, pe.norm)
		case pe.cek.key == nil:
			return nil, fmt.Errorf("parameter %s: no column_key matches its column encryption key (column master key %s); %s", name, pe.cek.masterKeys(), rotateHint)
		}
		out[strings.ToLower(name)] = pe
	}
	return out, nil
}

// encryptedParam returns how param is encrypted, nil if it is not.
func encryptedParam(enc map[string]*paramEnc, param *rdb.Param) *paramEnc {
	if len(enc) == 0 {
		return nil
	}
	return enc["@"+strings.ToLower(param.Name)]
}

// encodeEncryptedParam writes param as an encrypted RPC parameter: its name,
// the varbinary ciphertext, then the plaintext type and key it was encrypted
// with.
func (tds *Connection) encodeEncryptedParam(w *PacketWriter, param *rdb.Param, value interface{}, pe *paramEnc) error {
	if param.Out {
		return fmt.Errorf("parameter @%s: encrypted output parameters are not supported", param.Name)
	}
	ti, err := getParamTypeInfo(tds.ProtocolVersion, param.Type)
	if err != nil {
		return err
	}
	plain, null, err := aePlaintext(ti, param, value)
	if err != nil {
		return err
	}

	nameUtf16 := w.UCS2Prefixed("@", param.Name)
	w.WriteByte(byte(len(nameUtf16) / 2))
	w.WriteBuffer(nameUtf16)
	w.WriteByte(paramStatusCrypt)

	w.WriteByte(byte(typeVarBinary))
	switch {
	case null:
		w.WriteUint16(maxShortBinary)
		w.WriteUint16(0xFFFF)
	default:
		ct, err := pe.cek.key.Encrypt(plain, pe.encType == aeDeterministic)
		if err != nil {
			return fmt.Errorf("parameter @%s: %w", param.Name, err)
		}
		if len(ct) <= maxShortBinary {
			w.WriteUint16(maxShortBinary)
			w.WriteUint16(uint16(len(ct)))
			w.WriteBuffer(ct)
			break
		}
		w.WriteUint16(0xFFFF)
		w.WriteUint64(uint64(len(ct)))
		w.WriteUint32(uint32(len(ct)))
		w.WriteBuffer(ct)
		w.WriteUint32(0)
	}

	// ParamCipherInfo: TYPE_INFO EncryptionAlgo EncryptionType DatabaseId
	// CekId CekVersion CekMDVersion NormVersion.
	err = encodeType(w, ti, param, tds.paramCollation)
	if err != nil {
		return err
	}
	w.WriteByte(pe.alg)
	w.WriteByte(pe.encType)
	w.WriteUint32(pe.cek.dbID)
	w.WriteUint32(pe.cek.keyID)
	w.WriteUint32(pe.cek.keyVersion)
	w.WriteBuffer(pe.cek.mdVersion[:])
	w.WriteByte(pe.norm)
	return nil
}

// aePlaintext returns the bytes encrypted for a parameter value: UTF-16LE for
// nchar (space padded to its length) and nvarchar, the bytes for binary and
// varbinary, and 8 bytes little endian for integers.
func aePlaintext(ti paramTypeInfo, param *rdb.Param, value interface{}) (plain []byte, null bool, err error) {
	if value == nil || value == rdb.Null || param.Null {
		return nil, true, nil
	}
	if !aeValueType(driverType(ti.T), ti.typeInfo) || (ti.IsText && !ti.NChar) {
		return nil, false, fmt.Errorf("parameter @%s: encrypting %s values is not supported", param.Name, ti.SqlName)
	}
	if ti.T == typeIntN {
		v, ok := paramInt(value)
		if !ok {
			return nil, false, fmt.Errorf("parameter @%s: encrypting %T as %s is not supported", param.Name, value, ti.SqlName)
		}
		return binary.LittleEndian.AppendUint64(nil, uint64(v)), false, nil
	}
	var b []byte
	switch v := value.(type) {
	case string:
		b = []byte(v)
	case []byte:
		b = v
	default:
		return nil, false, fmt.Errorf("parameter @%s: encrypting %T as %s is not supported", param.Name, value, ti.SqlName)
	}
	if ti.NChar {
		b = uconv.Encode.AppendBytes(nil, b)
	}
	limit := param.Length
	if ti.NChar {
		limit *= 2
	}
	if !ti.IsMaxParam(param) && len(b) > limit {
		return nil, false, InputToolong{DataLen: uint32(len(b)), TypeLen: uint32(limit)}
	}
	// The server cannot pad an encrypted nchar(n); pad it as it would.
	if ti.T == typeNChar {
		for len(b) < limit {
			b = append(b, ' ', 0)
		}
	}
	return b, false, nil
}

func paramInt(value interface{}) (int64, bool) {
	switch v := value.(type) {
	case int:
		return int64(v), true
	case int8:
		return int64(v), true
	case int16:
		return int64(v), true
	case int32:
		return int64(v), true
	case int64:
		return v, true
	case uint8:
		return int64(v), true
	case uint16:
		return int64(v), true
	case uint32:
		return int64(v), true
	}
	return 0, false
}

func anyInt(v interface{}) int64 {
	n, _ := paramInt(v)
	return n
}

func anyBytes(v interface{}) []byte {
	switch v := v.(type) {
	case []byte:
		return v
	case string:
		return []byte(v)
	}
	return nil
}

// tableValuer collects every result set of a response, for driver queries.
type tableValuer struct {
	sets [][][]interface{}
	row  []interface{}
}

var _ rdb.DriverValuer = &tableValuer{}

func (t *tableValuer) Columns(cols []*rdb.Column) error {
	t.sets = append(t.sets, nil)
	t.row = make([]interface{}, len(cols))
	return nil
}

func (t *tableValuer) WriteField(c *rdb.Column, v *rdb.DriverValue, _ rdb.Assigner) error {
	if c.Index >= len(t.row) {
		return errors.New("describe parameter encryption: column index out of range")
	}
	switch {
	case v.Null:
		t.row[c.Index] = nil
	case v.Chunked:
		return errors.New("describe parameter encryption: unexpected chunked value")
	default:
		if b, ok := v.Value.([]byte); ok {
			t.row[c.Index] = bytes.Clone(b)
			break
		}
		t.row[c.Index] = v.Value
	}
	return nil
}

func (t *tableValuer) RowScanned() {
	last := len(t.sets) - 1
	t.sets[last] = append(t.sets[last], t.row)
	t.row = make([]interface{}, len(t.row))
}

func (t *tableValuer) Done() error               { return nil }
func (t *tableValuer) Message(*rdb.Message)      {}
func (t *tableValuer) RowsAffected(count uint64) {}
