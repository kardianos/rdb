// Copyright 2026 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package aecrypt

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf16"
)

// Fixtures in testdata come from github.com/swisscom/mssql-always-encrypted;
// see testdata/NOTICE.
const (
	fixtureRoot    = "0ff9e45335df3dec7be0649f741e6ea870e9d49d16fe4be7437ce22489f48ead"
	fixtureKeyPath = "currentuser/my/0be978ba81eed610015fd8b7caef55f1614ca3b6"
	fixturePlain   = "12345     " // An nchar(10) cell.

	// fixtureText is the Key of the fixture. extract_cek.ps1 prints
	// "column_key=" + fixtureText for testdata/cek_value.bin with the fixture
	// certificate imported into CurrentUser\My.
	fixtureText = "8xMmOS3TniY.D_nkUzXfPex74GSfdB5uqHDp1J0W_kvnQ3ziJIn0jq0"
)

// TestCell checks one real cell value kept out of the repository. Set
// AECRYPT_KEY to a key's text form and AECRYPT_CELL to a cell value of a
// deterministic column in hex, such as a SampleName read with Always Encrypted
// off. AECRYPT_WRAPPED, the key's stored value in hex, also checks the
// fingerprint. The test logs the plaintext.
func TestCell(t *testing.T) {
	keyText, cellHex := os.Getenv("AECRYPT_KEY"), os.Getenv("AECRYPT_CELL")
	if len(keyText) == 0 || len(cellHex) == 0 {
		t.Skip("set AECRYPT_KEY and AECRYPT_CELL to check a real cell value")
	}
	k, err := ParseKey(strings.TrimPrefix(keyText, "column_key="))
	if err != nil {
		t.Fatal(err)
	}
	if wrappedHex := os.Getenv("AECRYPT_WRAPPED"); len(wrappedHex) > 0 {
		if !k.Matches(mustHex(t, trimHex(wrappedHex))) {
			t.Fatal("key does not match AECRYPT_WRAPPED")
		}
	}
	cell := mustHex(t, trimHex(cellHex))
	plain, err := k.Decrypt(cell)
	if err != nil {
		t.Fatal(err)
	}
	u := make([]uint16, len(plain)/2)
	for i := range u {
		u[i] = uint16(plain[2*i]) | uint16(plain[2*i+1])<<8
	}
	t.Logf("plaintext %x, as UTF-16LE %q", plain, string(utf16.Decode(u)))
	again, err := k.Encrypt(plain, true)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(again, cell) {
		t.Fatalf("deterministic re-encryption differs: the column is randomized, or the IV derivation is wrong\ngot  %x\nwant %x", again, cell)
	}
}

func trimHex(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "0x")
	return strings.TrimPrefix(s, "0X")
}

func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func fixtureCMK(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	blk, _ := pem.Decode(readFixture(t, "cmk_private.pem"))
	if blk == nil {
		t.Fatal("cmk_private.pem: no PEM block")
	}
	pk, err := x509.ParsePKCS8PrivateKey(blk.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	return pk.(*rsa.PrivateKey)
}

func mustHex(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func fixtureKey(t *testing.T) *Key {
	t.Helper()
	k, err := NewKey(mustHex(t, fixtureRoot), readFixture(t, "cek_value.bin"))
	if err != nil {
		t.Fatal(err)
	}
	return k
}

// TestFixture checks this package against values written by SQL Server's own
// client: a stored column encryption key value and a cell it encrypted.
func TestFixture(t *testing.T) {
	stored := readFixture(t, "cek_value.bin")
	cell := readFixture(t, "cell_value.bin")
	plain := utf16LE(fixturePlain)

	t.Run("unwrap stored key", func(t *testing.T) {
		w, err := ParseWrapped(stored)
		if err != nil {
			t.Fatal(err)
		}
		if w.KeyPath != fixtureKeyPath {
			t.Fatalf("key path %q, want %q", w.KeyPath, fixtureKeyPath)
		}
		blk, _ := pem.Decode(readFixture(t, "cmk_cert.pem"))
		cert, err := x509.ParseCertificate(blk.Bytes)
		if err != nil {
			t.Fatal(err)
		}
		cmk := fixtureCMK(t)
		if !cmk.PublicKey.Equal(cert.PublicKey) {
			t.Fatal("cmk_private.pem is not the key of cmk_cert.pem")
		}
		root, err := w.Unwrap(cmk)
		if err != nil {
			t.Fatal(err)
		}
		if got := hex.EncodeToString(root); got != fixtureRoot {
			t.Fatalf("root %s, want %s", got, fixtureRoot)
		}
	})
	t.Run("decrypt stored cell", func(t *testing.T) {
		got, err := fixtureKey(t).Decrypt(cell)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, plain) {
			t.Fatalf("got %x, want %x", got, plain)
		}
	})
	// The stored cell is randomized (its IV is not the HMAC of its plaintext),
	// so it cannot check deterministic encryption; TestCell can.
	t.Run("text", func(t *testing.T) {
		if got := fixtureKey(t).Text(); got != fixtureText {
			t.Fatalf("got %s, want %s", got, fixtureText)
		}
	})
}

func TestParseKey(t *testing.T) {
	good := fixtureKey(t).Text()
	fpText, rootText, _ := strings.Cut(good, ".")

	list := []struct {
		Name    string
		Text    string
		WantErr string
	}{
		{Name: "valid", Text: good},
		{Name: "empty", Text: "", WantErr: "is not"},
		{Name: "no separator", Text: fpText + rootText, WantErr: "is not"},
		{Name: "extra part", Text: good + ".AA", WantErr: "key:"},
		{Name: "padded base64", Text: fpText + "=." + rootText, WantErr: "fingerprint:"},
		{Name: "standard base64 alphabet", Text: fpText + "." + strings.Replace(rootText, rootText[:1], "+", 1), WantErr: "key:"},
		{Name: "surrounding space", Text: " " + good, WantErr: "fingerprint:"},
		{Name: "short fingerprint", Text: fpText[:8] + "." + rootText, WantErr: "fingerprint is 6 bytes"},
		{Name: "short key", Text: fpText + "." + rootText[:40], WantErr: "key is 30 bytes"},
		{Name: "long key", Text: fpText + "." + rootText + "AAAA", WantErr: "key is 35 bytes"},
	}
	for _, tc := range list {
		t.Run(tc.Name, func(t *testing.T) {
			k, err := ParseKey(tc.Text)
			if len(tc.WantErr) > 0 {
				if err == nil || !strings.Contains(err.Error(), tc.WantErr) {
					t.Fatalf("got error %v, want one containing %q", err, tc.WantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got := k.Text(); got != tc.Text {
				t.Fatalf("round trip %s, want %s", got, tc.Text)
			}
		})
	}
}

func TestKeyMatches(t *testing.T) {
	stored := readFixture(t, "cek_value.bin")
	other := bytes.Clone(stored)
	other[len(other)-1] ^= 1
	fpText, rootText, _ := strings.Cut(fixtureKey(t).Text(), ".")
	flip := func(s string) string {
		c := byte('A')
		if s[0] == c {
			c = 'B'
		}
		return string(c) + s[1:]
	}

	list := []struct {
		Name    string
		Text    string
		Wrapped []byte
		Want    bool
	}{
		{Name: "its stored value", Text: fpText + "." + rootText, Wrapped: stored, Want: true},
		{Name: "another stored value", Text: fpText + "." + rootText, Wrapped: other},
		{Name: "damaged key", Text: fpText + "." + flip(rootText), Wrapped: stored},
		{Name: "damaged fingerprint", Text: flip(fpText) + "." + rootText, Wrapped: stored},
		{Name: "empty stored value", Text: fpText + "." + rootText},
	}
	for _, tc := range list {
		t.Run(tc.Name, func(t *testing.T) {
			k, err := ParseKey(tc.Text)
			if err != nil {
				t.Fatal(err)
			}
			if got := k.Matches(tc.Wrapped); got != tc.Want {
				t.Fatalf("got %t, want %t", got, tc.Want)
			}
		})
	}
}

func TestEncryptDecrypt(t *testing.T) {
	k := fixtureKey(t)
	list := []struct {
		Name  string
		Plain []byte
	}{
		{Name: "empty", Plain: []byte{}},
		{Name: "one byte", Plain: []byte{7}},
		{Name: "one short of a block", Plain: bytes.Repeat([]byte{1}, 15)},
		{Name: "one block", Plain: bytes.Repeat([]byte{2}, 16)},
		{Name: "one over a block", Plain: bytes.Repeat([]byte{3}, 17)},
		{Name: "nvarchar", Plain: utf16LE("260601001HLA_C1q")},
	}
	for _, tc := range list {
		for _, det := range []bool{true, false} {
			t.Run(fmt.Sprintf("%s deterministic=%t", tc.Name, det), func(t *testing.T) {
				a, err := k.Encrypt(tc.Plain, det)
				if err != nil {
					t.Fatal(err)
				}
				b, err := k.Encrypt(tc.Plain, det)
				if err != nil {
					t.Fatal(err)
				}
				if got := bytes.Equal(a, b); got != det {
					t.Fatalf("two encryptions equal: %t, want %t", got, det)
				}
				for _, v := range [][]byte{a, b} {
					got, err := k.Decrypt(v)
					if err != nil {
						t.Fatal(err)
					}
					if !bytes.Equal(got, tc.Plain) {
						t.Fatalf("got %x, want %x", got, tc.Plain)
					}
				}
			})
		}
	}
}

func TestDecryptRejects(t *testing.T) {
	k := fixtureKey(t)
	cell := readFixture(t, "cell_value.bin")
	otherRoot := make([]byte, KeySize)
	if _, err := rand.Read(otherRoot); err != nil {
		t.Fatal(err)
	}
	other, err := NewKey(otherRoot, readFixture(t, "cek_value.bin"))
	if err != nil {
		t.Fatal(err)
	}
	flipAt := func(i int) []byte {
		b := bytes.Clone(cell)
		b[i] ^= 0x80
		return b
	}
	// A correctly tagged value whose plaintext is one zero block: zero is no pad.
	badPad := func() []byte {
		iv := make([]byte, ivSize)
		body := make([]byte, aes.BlockSize)
		block, err := aes.NewCipher(k.enc[:])
		if err != nil {
			t.Fatal(err)
		}
		cipher.NewCBCEncrypter(block, iv).CryptBlocks(body, body)
		v := append([]byte{cellVersion}, k.tag(iv, body)...)
		v = append(v, iv...)
		return append(v, body...)
	}()

	list := []struct {
		Name    string
		Key     *Key
		Value   []byte
		WantErr error
		WantMsg string
	}{
		{Name: "other key", Key: other, Value: cell, WantErr: ErrAuth},
		{Name: "damaged tag", Key: k, Value: flipAt(1), WantErr: ErrAuth},
		{Name: "damaged IV", Key: k, Value: flipAt(1 + tagSize), WantErr: ErrAuth},
		{Name: "damaged ciphertext", Key: k, Value: flipAt(len(cell) - 1), WantErr: ErrAuth},
		{Name: "version 2", Key: k, Value: append([]byte{2}, cell[1:]...), WantMsg: "version 2"},
		{Name: "too short", Key: k, Value: cell[:cellHeader], WantMsg: "not a valid length"},
		{Name: "not whole blocks", Key: k, Value: cell[:len(cell)-1], WantMsg: "not a valid length"},
		{Name: "bad padding", Key: k, Value: badPad, WantMsg: "bad padding"},
	}
	for _, tc := range list {
		t.Run(tc.Name, func(t *testing.T) {
			_, err := tc.Key.Decrypt(tc.Value)
			if err == nil {
				t.Fatal("no error")
			}
			if tc.WantErr != nil && !errors.Is(err, tc.WantErr) {
				t.Fatalf("got %v, want %v", err, tc.WantErr)
			}
			if len(tc.WantMsg) > 0 && !strings.Contains(err.Error(), tc.WantMsg) {
				t.Fatalf("got %v, want one containing %q", err, tc.WantMsg)
			}
		})
	}
}

func TestWrap(t *testing.T) {
	cmk := fixtureCMK(t)
	root := mustHex(t, fixtureRoot)
	stored, err := Wrap(root, "LocalMachine/My/07980C4EBF9225C65ADCC4E1B165B5CE6E91B35B", cmk)
	if err != nil {
		t.Fatal(err)
	}
	otherCMK, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	badSig := bytes.Clone(stored)
	badSig[len(badSig)-1] ^= 1

	list := []struct {
		Name    string
		Stored  []byte
		CMK     *rsa.PrivateKey
		WantMsg string
	}{
		{Name: "round trip", Stored: stored, CMK: cmk},
		{Name: "other master key", Stored: stored, CMK: otherCMK, WantMsg: "not signed by this column master key"},
		{Name: "damaged signature", Stored: badSig, CMK: cmk, WantMsg: "not signed by this column master key"},
	}
	for _, tc := range list {
		t.Run(tc.Name, func(t *testing.T) {
			w, err := ParseWrapped(tc.Stored)
			if err != nil {
				t.Fatal(err)
			}
			if want := "localmachine/my/07980c4ebf9225c65adcc4e1b165b5ce6e91b35b"; w.KeyPath != want {
				t.Fatalf("key path %q, want %q", w.KeyPath, want)
			}
			got, err := w.Unwrap(tc.CMK)
			if len(tc.WantMsg) > 0 {
				if err == nil || !strings.Contains(err.Error(), tc.WantMsg) {
					t.Fatalf("got error %v, want one containing %q", err, tc.WantMsg)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, root) {
				t.Fatalf("got %x, want %x", got, root)
			}
		})
	}
}

func TestParseWrappedRejects(t *testing.T) {
	stored := readFixture(t, "cek_value.bin")
	oddPath := bytes.Clone(stored)
	oddPath[1]++

	list := []struct {
		Name    string
		Stored  []byte
		WantMsg string
	}{
		{Name: "empty", Stored: nil, WantMsg: "too short"},
		{Name: "version 2", Stored: append([]byte{2}, stored[1:]...), WantMsg: "version 2"},
		{Name: "odd key path length", Stored: oddPath, WantMsg: "too short"},
		{Name: "no signature", Stored: stored[:len(stored)-256], WantMsg: "too short"},
	}
	for _, tc := range list {
		t.Run(tc.Name, func(t *testing.T) {
			_, err := ParseWrapped(tc.Stored)
			if err == nil || !strings.Contains(err.Error(), tc.WantMsg) {
				t.Fatalf("got error %v, want one containing %q", err, tc.WantMsg)
			}
		})
	}
}

// TestKeyRedacted checks that formatting a Key, alone or inside a struct,
// never prints the secret key.
func TestKeyRedacted(t *testing.T) {
	k := fixtureKey(t)
	_, rootText, _ := strings.Cut(k.Text(), ".")
	type holder struct {
		Name string
		Key  *Key
	}
	list := []struct {
		Name   string
		Format string
		Value  any
	}{
		{Name: "v", Format: "%v", Value: k},
		{Name: "s", Format: "%s", Value: k},
		{Name: "+v", Format: "%+v", Value: k},
		{Name: "#v", Format: "%#v", Value: k},
		{Name: "struct v", Format: "%v", Value: holder{Name: "x", Key: k}},
		{Name: "struct #v", Format: "%#v", Value: holder{Name: "x", Key: k}},
	}
	for _, tc := range list {
		t.Run(tc.Name, func(t *testing.T) {
			got := fmt.Sprintf(tc.Format, tc.Value)
			if strings.Contains(got, rootText) || strings.Contains(got, fixtureRoot) {
				t.Fatalf("%s printed the key: %s", tc.Format, got)
			}
		})
	}
}
