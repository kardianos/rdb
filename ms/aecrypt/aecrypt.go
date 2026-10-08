// Copyright 2026 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

// Package aecrypt is the client side cryptography of SQL Server Always Encrypted
// for a column encryption key held in plain form, such as in a DSN, rather than
// unwrapped at run time from a key store.
//
// Always Encrypted protects a column with a column encryption key (CEK): 32
// random bytes. SQL Server never holds the plain CEK. It stores the CEK wrapped
// (RSA encrypted) by a column master key (CMK), usually a certificate on the
// client machines, and sends that wrapped value along with any result or
// parameter description that needs it. A Key is the plain CEK plus a fingerprint
// that binds it to its wrapped value: a client holding several keys can tell
// which one the server means, and a damaged key matches nothing, so it is never
// used to encrypt.
//
// The text form of a Key, as printed by extract_cek.ps1 on the machine holding
// the CMK certificate, is
//
//	<fingerprint>.<key>
//
// both base64url without padding, 11 + 1 + 43 = 55 characters. The fingerprint
// is the first 8 bytes of SHA-256(wrapped value || key).
//
// Cell values use AEAD_AES_256_CBC_HMAC_SHA256, the only algorithm SQL Server
// offers:
//
//	0x01 | HMAC-SHA256 tag (32) | IV (16) | AES-256-CBC ciphertext, PKCS #7 padded
//
// The tag covers 0x01 | IV | ciphertext | 0x01. Deterministic encryption takes
// the IV from an HMAC of the plaintext, so equal plaintexts give equal values;
// randomized encryption uses a random IV.
package aecrypt

import (
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
	"unicode/utf16"
)

const (
	// KeySize is the size of a column encryption key.
	KeySize = 32
	// FingerprintSize is the size of a Key fingerprint.
	FingerprintSize = 8

	cellVersion = 0x01
	tagSize     = sha256.Size
	ivSize      = aes.BlockSize
	cellHeader  = 1 + tagSize + ivSize

	wrappedVersion = 0x01
)

// ErrAuth reports a cell value whose authentication tag does not match: it was
// encrypted with another key, or it is damaged.
var ErrAuth = errors.New("aecrypt: authentication tag mismatch: wrong key or damaged value")

var b64 = base64.RawURLEncoding

// Key is one column encryption key, ready to encrypt and decrypt cell values.
type Key struct {
	fp   [FingerprintSize]byte
	root [KeySize]byte

	enc, mac, iv [sha256.Size]byte // Derived cell keys.
}

// NewKey returns the Key for the plain column encryption key root, fingerprinted
// against wrapped, the key's value exactly as SQL Server stores it
// (sys.column_encryption_key_values.encrypted_value).
func NewKey(root, wrapped []byte) (*Key, error) {
	if len(root) != KeySize {
		return nil, fmt.Errorf("aecrypt: key is %d bytes, want %d", len(root), KeySize)
	}
	if len(wrapped) == 0 {
		return nil, errors.New("aecrypt: missing wrapped key value")
	}
	k := &Key{}
	copy(k.root[:], root)
	k.fp = fingerprint(wrapped, k.root[:])
	k.derive()
	return k, nil
}

// ParseKey parses the text form of a Key, "<fingerprint>.<key>".
func ParseKey(s string) (*Key, error) {
	fpText, rootText, ok := strings.Cut(s, ".")
	if !ok {
		return nil, errors.New(`aecrypt: key text is not "<fingerprint>.<key>"`)
	}
	fp, err := b64.DecodeString(fpText)
	if err != nil {
		return nil, fmt.Errorf("aecrypt: key fingerprint: %w", err)
	}
	if len(fp) != FingerprintSize {
		return nil, fmt.Errorf("aecrypt: key fingerprint is %d bytes, want %d", len(fp), FingerprintSize)
	}
	root, err := b64.DecodeString(rootText)
	if err != nil {
		return nil, fmt.Errorf("aecrypt: key: %w", err)
	}
	if len(root) != KeySize {
		return nil, fmt.Errorf("aecrypt: key is %d bytes, want %d", len(root), KeySize)
	}
	k := &Key{}
	copy(k.fp[:], fp)
	copy(k.root[:], root)
	k.derive()
	return k, nil
}

// Text returns the text form of k, "<fingerprint>.<key>". It holds the secret key.
func (k *Key) Text() string {
	return b64.EncodeToString(k.fp[:]) + "." + b64.EncodeToString(k.root[:])
}

// String names k by its fingerprint and does not reveal the key.
func (k *Key) String() string {
	return "aecrypt.Key(" + b64.EncodeToString(k.fp[:]) + ")"
}

// GoString is String, so %#v does not print the key either.
func (k *Key) GoString() string {
	return k.String()
}

// Fingerprint returns the fingerprint of k. It is not secret.
func (k *Key) Fingerprint() [FingerprintSize]byte {
	return k.fp
}

// Matches reports whether wrapped, a column encryption key value as SQL Server
// sends it, is the stored form of k. A false result means another key, or a
// damaged fingerprint or key.
func (k *Key) Matches(wrapped []byte) bool {
	fp := fingerprint(wrapped, k.root[:])
	return subtle.ConstantTimeCompare(fp[:], k.fp[:]) == 1
}

// Encrypt returns the cell value for plaintext, the bytes of the column's value
// in its SQL Server form (UTF-16LE for nvarchar).
func (k *Key) Encrypt(plaintext []byte, deterministic bool) ([]byte, error) {
	iv := make([]byte, ivSize)
	if deterministic {
		copy(iv, hmacSum(k.iv[:], plaintext))
	} else if _, err := rand.Read(iv); err != nil {
		return nil, fmt.Errorf("aecrypt: random IV: %w", err)
	}
	pad := aes.BlockSize - len(plaintext)%aes.BlockSize
	out := make([]byte, cellHeader+len(plaintext)+pad)
	out[0] = cellVersion
	copy(out[1+tagSize:], iv)
	body := out[cellHeader:]
	copy(body, plaintext)
	for i := len(plaintext); i < len(body); i++ {
		body[i] = byte(pad)
	}
	block, err := aes.NewCipher(k.enc[:])
	if err != nil {
		return nil, err
	}
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(body, body)
	copy(out[1:], k.tag(iv, body))
	return out, nil
}

// Decrypt returns the plaintext of a cell value. It returns ErrAuth if the
// value was not encrypted with k or is damaged.
func (k *Key) Decrypt(value []byte) ([]byte, error) {
	if len(value) < cellHeader+aes.BlockSize || (len(value)-cellHeader)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("aecrypt: cell value is %d bytes, not a valid length", len(value))
	}
	if value[0] != cellVersion {
		return nil, fmt.Errorf("aecrypt: cell value version %d, want %d", value[0], cellVersion)
	}
	tag := value[1 : 1+tagSize]
	iv := value[1+tagSize : cellHeader]
	body := value[cellHeader:]
	if !hmac.Equal(tag, k.tag(iv, body)) {
		return nil, ErrAuth
	}
	block, err := aes.NewCipher(k.enc[:])
	if err != nil {
		return nil, err
	}
	plain := make([]byte, len(body))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(plain, body)
	pad := int(plain[len(plain)-1])
	if pad == 0 || pad > aes.BlockSize {
		return nil, errors.New("aecrypt: bad padding")
	}
	for _, b := range plain[len(plain)-pad:] {
		if int(b) != pad {
			return nil, errors.New("aecrypt: bad padding")
		}
	}
	return plain[:len(plain)-pad], nil
}

// tag is the authentication tag of a cell value with the given IV and ciphertext.
func (k *Key) tag(iv, body []byte) []byte {
	h := hmac.New(sha256.New, k.mac[:])
	h.Write([]byte{cellVersion})
	h.Write(iv)
	h.Write(body)
	h.Write([]byte{1}) // Size of the version field.
	return h.Sum(nil)
}

// derive computes the cell keys from the root key.
func (k *Key) derive() {
	const label = " with encryption algorithm:AEAD_AES_256_CBC_HMAC_SHA256 and key length:256"
	copy(k.enc[:], hmacSum(k.root[:], utf16LE("Microsoft SQL Server cell encryption key"+label)))
	copy(k.mac[:], hmacSum(k.root[:], utf16LE("Microsoft SQL Server cell MAC key"+label)))
	copy(k.iv[:], hmacSum(k.root[:], utf16LE("Microsoft SQL Server cell IV key"+label)))
}

func fingerprint(wrapped, root []byte) [FingerprintSize]byte {
	h := sha256.New()
	h.Write(wrapped)
	h.Write(root)
	var fp [FingerprintSize]byte
	copy(fp[:], h.Sum(nil))
	return fp
}

func hmacSum(key, msg []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(msg)
	return h.Sum(nil)
}

func utf16LE(s string) []byte {
	u := utf16.Encode([]rune(s))
	b := make([]byte, 2*len(u))
	for i, c := range u {
		binary.LittleEndian.PutUint16(b[2*i:], c)
	}
	return b
}

// Wrapped is a column encryption key value as SQL Server stores it
// (sys.column_encryption_key_values.encrypted_value) and sends it:
//
//	0x01 | key path length (2) | ciphertext length (2) | key path (UTF-16LE) |
//	RSA-OAEP (SHA-1) ciphertext of the key | RSA PKCS #1 v1.5 SHA-256 signature
//
// The signature covers every byte before it.
type Wrapped struct {
	KeyPath    string // Column master key path, such as "localmachine/my/<thumbprint>".
	Ciphertext []byte
	Signature  []byte

	signed []byte
}

// ParseWrapped splits a column encryption key value into its parts.
func ParseWrapped(b []byte) (*Wrapped, error) {
	const head = 1 + 2 + 2
	if len(b) < head {
		return nil, fmt.Errorf("aecrypt: wrapped key value is %d bytes, too short", len(b))
	}
	if b[0] != wrappedVersion {
		return nil, fmt.Errorf("aecrypt: wrapped key value version %d, want %d", b[0], wrappedVersion)
	}
	pathLen := int(binary.LittleEndian.Uint16(b[1:]))
	ctLen := int(binary.LittleEndian.Uint16(b[3:]))
	end := head + pathLen + ctLen
	if pathLen%2 != 0 || len(b) <= end {
		return nil, fmt.Errorf("aecrypt: wrapped key value is %d bytes, too short for a %d byte key path and %d byte ciphertext and a signature", len(b), pathLen, ctLen)
	}
	path := make([]uint16, pathLen/2)
	for i := range path {
		path[i] = binary.LittleEndian.Uint16(b[head+2*i:])
	}
	return &Wrapped{
		KeyPath:    string(utf16.Decode(path)),
		Ciphertext: b[head+pathLen : end],
		Signature:  b[end:],
		signed:     b[:end],
	}, nil
}

// Unwrap checks the signature of w against the column master key and returns
// the plain column encryption key.
func (w *Wrapped) Unwrap(cmk *rsa.PrivateKey) ([]byte, error) {
	sum := sha256.Sum256(w.signed)
	if err := rsa.VerifyPKCS1v15(&cmk.PublicKey, crypto.SHA256, sum[:], w.Signature); err != nil {
		return nil, fmt.Errorf("aecrypt: wrapped key value is not signed by this column master key: %w", err)
	}
	root, err := rsa.DecryptOAEP(sha1.New(), nil, cmk, w.Ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("aecrypt: unwrap key: %w", err)
	}
	if len(root) != KeySize {
		return nil, fmt.Errorf("aecrypt: unwrapped key is %d bytes, want %d", len(root), KeySize)
	}
	return root, nil
}

// Wrap returns the column encryption key value of root under the column master
// key cmk at keyPath, for CREATE COLUMN ENCRYPTION KEY ... ENCRYPTED_VALUE.
// The key path is stored in lower case, as SQL Server's own tools do.
func Wrap(root []byte, keyPath string, cmk *rsa.PrivateKey) ([]byte, error) {
	if len(root) != KeySize {
		return nil, fmt.Errorf("aecrypt: key is %d bytes, want %d", len(root), KeySize)
	}
	path := utf16LE(strings.ToLower(keyPath))
	ct, err := rsa.EncryptOAEP(sha1.New(), rand.Reader, &cmk.PublicKey, root, nil)
	if err != nil {
		return nil, fmt.Errorf("aecrypt: wrap key: %w", err)
	}
	b := make([]byte, 0, 5+len(path)+len(ct)+cmk.Size())
	b = append(b, wrappedVersion)
	b = binary.LittleEndian.AppendUint16(b, uint16(len(path)))
	b = binary.LittleEndian.AppendUint16(b, uint16(len(ct)))
	b = append(b, path...)
	b = append(b, ct...)
	sum := sha256.Sum256(b)
	sig, err := rsa.SignPKCS1v15(nil, cmk, crypto.SHA256, sum[:])
	if err != nil {
		return nil, fmt.Errorf("aecrypt: sign wrapped key value: %w", err)
	}
	return append(b, sig...), nil
}
