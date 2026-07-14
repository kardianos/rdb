# UTF-8 in SQL Server and the rdb/ms Package

## 1. UTF-8 Collations in SQL Server

### Background

SQL Server 2019 (TDS 7.4) introduced native UTF-8 support for `char` and `varchar`
types via new collations with the `_UTF8` suffix. Before this, `varchar` was limited
to single-byte code pages (e.g., Windows-1252), and the only way to store full Unicode
was `nvarchar` (UTF-16LE).

UTF-8 collations carry the `_UTF8` suffix appended to an existing supplementary-character
or BIN2 collation:

```
Latin1_General_100_CI_AS_SC_UTF8
Latin1_General_100_BIN2_UTF8
Japanese_Bushu_Kakusu_100_CS_AS_KS_WS_SC_UTF8
```

Restrictions: `_UTF8` requires the base collation to have `_SC` (supplementary characters)
or `_VSS` (variation-selector-sensitive), or to be `_BIN2`. Legacy `SQL_*` collations
and `_BIN` (non-BIN2) collations cannot use it.

### What UTF-8 Collations Change

When a UTF-8 collation is applied to a `char`/`varchar` column, the column stores
full Unicode using UTF-8 encoding (1-4 bytes per character). The `n` in `varchar(n)`
means *bytes*, not characters.

**`nchar`/`nvarchar` are completely unaffected by UTF-8 collations** — they always use
UTF-16LE regardless of the column or database collation.

### How This Interacts with TDS

The 5-byte `COLLATION` field in TDS column metadata carries the encoding information.
From the MS-TDS v39 spec (`ms/_ref/MS-TDS_v39-0/02-messages/07-grammar-definition.md`):

```
LCID             = 20BIT
fIgnoreCase      = BIT
fIgnoreAccent    = BIT
fIgnoreWidth     = BIT
fIgnoreKana      = BIT
fBinary          = BIT
fBinary2         = BIT
fUTF8            = BIT          ; <-- key bit
FRESERVEDBIT     = BIT
Version          = 4BIT
SortId           = BYTE

COLLATION        = LCID ColFlags Version SortId
```

The `fUTF8` bit (bit 6 of the ColFlags byte, i.e., byte 4 bits when combined with LCID)
tells the client that `varchar` data for this column is UTF-8 encoded. When this bit
is clear, the client should interpret `varchar` bytes using the code page derived from
LCID/SortId.

### How This Interacts with the ms Package Today

The ms package reads the 5-byte collation from column metadata (`coder.go:1241-1242`):

```go
if info.IsText {
    copy(column.Collation[:], read(5))
}
```

But it **does not interpret the collation bytes**. The `Collation [5]byte` field on
`SQLColumn` is stored but never parsed — so the `fUTF8` bit is read but ignored.

When *sending* parameters, the collation is **hardcoded** (`coder.go:61-64`):

```go
if ti.IsText {
    // TODO: Handle collation.
    collation := []byte{0x09, 0x04, 0xD0, 0x00, 0x34}
    w.WriteBuffer(collation)
}
```

This is the collation for `SQL_Latin1_General_CP1_CI_AS` — a non-UTF-8 collation.

The current encoding logic is binary: `NChar == true` means UTF-16LE encode/decode;
`NChar == false` means raw bytes with no transcoding (`coder.go:241-252`):

```go
case string:
    if ti.NChar {
        writeBb = uconv.Encode.FromString(v)   // Go string (UTF-8) → UTF-16LE
    } else {
        writeBb = []byte(v)                     // Go string (UTF-8) → raw bytes
    }
```

**This means the ms package already works correctly for UTF-8 varchar columns by
accident**: Go strings are natively UTF-8, and `[]byte(v)` preserves that encoding.
When the server returns UTF-8-collated varchar data, the raw bytes are already valid
UTF-8, so they decode correctly into Go strings without any conversion.

The one issue is the hardcoded collation sent with parameters — it does not signal
UTF-8 to the server. In practice this may not matter for parameterized queries where
the server knows the target column's collation and can convert implicitly. But it would
matter for `varchar` literal expressions in ad-hoc SQL that rely on the parameter's
declared collation.

### UTF-8 Feature Extension Negotiation

TDS 7.4 introduced a `UTF8_SUPPORT` feature extension (FeatureId `0x0A`) in the
LOGIN7 handshake. From `ms/_ref/MS-TDS_v39-0/02-messages/08-packet-header-message-type.md`:

> The presence of the UTF8_SUPPORT FeatureExt indicates whether the client's ability
> to send and receive UTF-8 encoded data SHOULD be supported. [...] Failure of the
> client to receive an acknowledgement of UTF-8 feature extension support from the
> server indicates that the server cannot send or receive UTF-8 encoded data.

The ms package currently does **not** send this feature extension during login. This
means the server may not send UTF-8 encoded varchar data even when the column uses a
UTF-8 collation — the server may fall back to a code-page encoding or reject the
connection's ability to handle UTF-8 data. The behavior here is implementation-defined
by the server version.

## 2. Does a UTF-8 Collation Require UTF-16 at the Protocol Level?

**No.** The whole point of UTF-8 collations + TDS 7.4 is to allow `varchar` data to
travel as UTF-8 bytes on the wire. The key distinction is:

| TDS Type          | Wire Encoding                              | Affected by UTF-8 collation? |
|-------------------|--------------------------------------------|------------------------------|
| `NVARCHARTYPE` (0xE7), `NCHARTYPE` (0xEF), `NTEXTTYPE` (0x63) | Always UTF-16LE | No — always UTF-16LE |
| `BIGVARCHARTYPE` (0xA7), `BIGCHARTYPE` (0xAF), `TEXTTYPE` (0x23) | Encoding determined by collation | Yes — UTF-8 when `fUTF8` bit is set |

When a `varchar` column has a UTF-8 collation:

1. The server includes the `fUTF8` bit in the column's `COLLATION` metadata.
2. The server sends the varchar data as raw UTF-8 bytes.
3. The client reads those bytes and knows (from the `fUTF8` bit) to interpret them as
   UTF-8 rather than a legacy code page.

When the same column has a traditional collation (e.g., `Latin1_General_CI_AS`), the
bytes represent the associated code page (e.g., Windows-1252).

**`nvarchar` columns always use UTF-16LE on the wire**, regardless of any UTF-8
collation setting. The N-types are defined by the TDS protocol to always be UCS-2/UTF-16LE.

### What the ms Package Needs

For *receiving* data: The ms package treats non-NChar data as raw bytes (`coder.go:1390-1392`),
which works for UTF-8 since Go strings are UTF-8. But if a `varchar` column uses a
legacy code page (e.g., Windows-1252), the raw bytes will not be valid UTF-8 for
non-ASCII characters. The package should ideally check the `fUTF8` bit and, when it's
not set, transcode from the code page to UTF-8. Today this is a latent bug for any
non-ASCII varchar data under non-UTF-8 collations.

For *sending* data: The hardcoded collation should either match the server's collation
or, when UTF-8 is negotiated, include the `fUTF8` bit so the server knows the parameter
bytes are UTF-8.

## 3. Efficiency Gains from UTF-8 Transmission

### Byte Cost Per Character

| Character Range                          | UTF-8 Bytes | UTF-16LE Bytes | Winner                 |
|------------------------------------------|-------------|----------------|------------------------|
| ASCII (U+0000–U+007F)                    | 1           | 2              | **UTF-8 (50% saving)** |
| Latin Extended, Greek, Cyrillic, Hebrew, Arabic (U+0080–U+07FF) | 2 | 2 | Tie |
| CJK, most BMP (U+0800–U+FFFF)           | 3           | 2              | **UTF-16 (33% saving)** |
| Supplementary (U+10000–U+10FFFF)         | 4           | 4              | Tie                    |

### Practical Impact

For **English/ASCII-dominant data** (which is the common case for identifiers, URLs,
config values, log messages, most Western business data):

- `nvarchar` transmits 2x the bytes needed — every ASCII character costs 2 bytes on
  the wire.
- `varchar` with UTF-8 collation transmits exactly the UTF-8 bytes — 1 byte per ASCII
  character.
- **~50% wire savings** for ASCII-dominant text.

For **CJK-dominant data** (Chinese, Japanese, Korean):

- `nvarchar` transmits 2 bytes per character.
- `varchar` UTF-8 transmits 3 bytes per character.
- **~50% more wire traffic** with UTF-8.

For **mixed European text** (accented Latin, Cyrillic, Greek):

- Both are 2 bytes per character — no difference.

### Storage vs. Transmission

SQL Server page/row compression can reduce the storage overhead of `nvarchar` for
in-row data — compressed nvarchar with ASCII data approaches 1 byte/char on disk.
But **compression does not affect TDS wire transmission** — the data is always sent
uncompressed over TDS. This means the wire savings from UTF-8 are real even when
storage compression is in use.

For `varchar(max)` / large text values stored off-row (LOB pages), compression does
not apply, so UTF-8 saves both storage and wire bytes.

### What This Means for rdb/ms

The ms package currently defaults all Go `string`/`rdb.TypeVarChar` parameters to
`nvarchar` (UTF-16LE). This means:

1. Every string parameter goes through `uconv.Encode.FromString()` — a UTF-8 → UTF-16LE
   conversion that doubles memory and CPU.
2. Every string result from an nvarchar column goes through `uconv.Decode.ToBytes()` —
   a UTF-16LE → UTF-8 conversion.
3. The wire traffic for string data is ~2x what it needs to be for ASCII-heavy data.

If the ms package supported UTF-8 varchar, string parameters and results would be
zero-copy for Go's native UTF-8 strings: `[]byte(s)` to send, `string(b)` to receive.
No conversion, no extra allocation, and half the wire bytes for ASCII text.

## 4. Testing UTF-8 Storage and Transmission

### Creating a UTF-8 Test Environment

```sql
-- Create a database with UTF-8 default collation
CREATE DATABASE UTF8TestDB
COLLATE Latin1_General_100_CI_AS_SC_UTF8;
GO

USE UTF8TestDB;
GO

-- Create a test table with both UTF-8 varchar and UTF-16 nvarchar
CREATE TABLE dbo.EncodingTest (
    ID          int IDENTITY PRIMARY KEY,
    TextUTF8    varchar(400)  COLLATE Latin1_General_100_CI_AS_SC_UTF8,
    TextUTF16   nvarchar(200)
);

-- Or apply UTF-8 at the column level in any database
CREATE TABLE dbo.MixedCollation (
    PlainVarchar   varchar(100)  COLLATE Latin1_General_CI_AS,         -- code page 1252
    UTF8Varchar    varchar(100)  COLLATE Latin1_General_100_CI_AS_SC_UTF8,  -- UTF-8
    UnicodeNVar    nvarchar(100)                                        -- UTF-16LE
);
```

### Verifying Storage Differences

```sql
INSERT INTO dbo.EncodingTest (TextUTF8, TextUTF16)
VALUES
    (N'Hello world',      N'Hello world'),       -- ASCII
    (N'Chào thế giới',    N'Chào thế giới'),     -- Vietnamese (2-byte UTF-8)
    (N'こんにちは',         N'こんにちは'),          -- Japanese (3-byte UTF-8)
    (N'😀🌍🎉',           N'😀🌍🎉');            -- Emoji (4-byte UTF-8)

SELECT
    TextUTF8,
    DATALENGTH(TextUTF8)  AS UTF8_Bytes,
    DATALENGTH(TextUTF16) AS UTF16_Bytes
FROM dbo.EncodingTest;
```

Expected results:

| Text          | UTF-8 Bytes | UTF-16 Bytes |
|---------------|-------------|--------------|
| Hello world   | 11          | 22           |
| Chào thế giới | 18          | 26           |
| こんにちは      | 15          | 10           |
| 😀🌍🎉        | 12          | 12           |

### Testing from Go with rdb/ms

A Go integration test should:

1. Create a table with a `varchar` column using a UTF-8 collation.
2. Insert strings containing ASCII, multi-byte UTF-8, and supplementary characters.
3. Read them back and verify byte-for-byte equality with the original Go strings.
4. Compare `DATALENGTH()` results for the UTF-8 column vs. an `nvarchar` column.
5. Verify the `Collation` bytes on the `SQLColumn` — check that the `fUTF8` bit
   (bit 27 of the first 4 bytes, little-endian) is set.

```go
func TestUTF8Collation(t *testing.T) {
    // Setup: CREATE TABLE with varchar(...) COLLATE Latin1_General_100_CI_AS_SC_UTF8

    tests := []string{
        "Hello world",
        "Chào thế giới",
        "こんにちは",
        "😀🌍🎉",
        "Mixed: café 日本語 🎵",
    }
    for _, want := range tests {
        // INSERT then SELECT
        // got := result string
        if got != want {
            t.Errorf("roundtrip failed: got %q, want %q", got, want)
        }
    }
}
```

### Verifying Wire Encoding

To confirm UTF-8 bytes are actually on the wire (not transcoded through UTF-16), capture
traffic with Wireshark or a TDS proxy and look at the COLMETADATA token:

1. Find the 5-byte COLLATION for the varchar column.
2. Check byte 4 (0-indexed), bits — the `fUTF8` bit should be set.
3. Verify the ROW data bytes match the expected UTF-8 encoding of the text.

### Checking Available UTF-8 Collations

```sql
SELECT Name, Description
FROM fn_helpcollations()
WHERE Name LIKE '%UTF8'
ORDER BY Name;
```

## 5. Transitioning from nvarchar (UTF-16) to varchar (UTF-8) Collation

### Why Transition?

- **Wire efficiency**: ~50% less data over TDS for ASCII-heavy workloads.
- **Storage efficiency**: ~50% less for ASCII-heavy data (more for `varchar(max)` LOBs).
- **Go driver simplicity**: Eliminates UTF-8↔UTF-16 conversion — Go strings are
  natively UTF-8, so varchar with UTF-8 collation is a zero-copy path.
- **Memory**: Half the allocations in the driver for string parameters and results.

### Option A: In-Place ALTER COLUMN

The simplest approach — change each column's type and collation:

```sql
-- Before: nvarchar(200)
-- After:  varchar(600) with UTF-8 collation
-- Note: varchar(n) is BYTES, so multiply by 3 for safety (worst case: all 3-byte chars)

ALTER TABLE dbo.MyTable
ALTER COLUMN MyColumn VARCHAR(600) COLLATE Latin1_General_100_CI_AS_SC_UTF8;
```

**Pros**: Simple, single statement per column.
**Cons**: Takes a schema modification lock; can be slow for large tables; rebuilds
indexes that reference the column.

### Option B: Shadow Table Migration

Create a new table, copy data, swap:

```sql
CREATE TABLE dbo.MyTable_New (
    ID       int PRIMARY KEY,
    MyColumn varchar(600) COLLATE Latin1_General_100_CI_AS_SC_UTF8,
    ...
);

INSERT INTO dbo.MyTable_New
SELECT * FROM dbo.MyTable;

-- Swap
EXEC sp_rename 'dbo.MyTable', 'MyTable_Old';
EXEC sp_rename 'dbo.MyTable_New', 'MyTable';
-- Recreate foreign keys, indexes, triggers, permissions...
```

**Pros**: Less blocking — the original table remains available during copy.
**Cons**: Complex — must re-create all dependent objects (FKs, indexes, triggers, views,
stored procedures).

### Option C: Database-Level Collation Change

Set the UTF-8 collation as the database default, then alter individual columns:

```sql
ALTER DATABASE MyDB COLLATE Latin1_General_100_CI_AS_SC_UTF8;
```

This changes the default for new columns but does **not** alter existing columns. You
still need to ALTER each existing column. However, this ensures new columns automatically
use UTF-8.

### Option D: Hybrid — Keep nvarchar, Add UTF-8 for New Columns

Don't migrate existing columns. Instead:

- Set the database collation to UTF-8.
- Use `varchar` for all new string columns.
- Keep existing `nvarchar` columns as-is.
- Migrate individual high-traffic columns opportunistically.

This avoids a big-bang migration while getting the benefits for new work.

### Critical Sizing Consideration

`nvarchar(n)` means `n` characters (up to `n*2` bytes).
`varchar(n)` with UTF-8 means `n` *bytes* (variable characters per byte).

A naive `nvarchar(100)` → `varchar(100)` conversion **will truncate** any non-ASCII
data. Sizing guidance:

| Data Profile           | Multiplier | Example: nvarchar(100) → |
|------------------------|------------|--------------------------|
| ASCII only             | 1×         | varchar(100)             |
| Mostly ASCII, some accented | 2× | varchar(200)             |
| Mixed / CJK possible  | 3×         | varchar(300)             |
| Supplementary chars    | 4×         | varchar(400)             |

Audit your actual data first:

```sql
SELECT
    MAX(DATALENGTH(
        CONVERT(varchar(max), MyColumn COLLATE Latin1_General_100_CI_AS_SC_UTF8)
    )) AS MaxUTF8Bytes,
    MAX(LEN(MyColumn)) AS MaxChars
FROM dbo.MyTable;
```

### Risks

1. **Truncation**: As above — `varchar(n)` is bytes, not characters. Size carefully.

2. **Collation conflicts**: Mixing UTF-8 varchar with non-UTF-8 varchar or nvarchar in
   JOINs, UNIONs, or comparisons may cause implicit conversion errors. Use a consistent
   collation across the database.

3. **Performance for CJK**: If your data is predominantly CJK, UTF-8 is *worse* than
   nvarchar — 3 bytes vs 2 bytes per character. Benchmark before committing.

4. **Index size changes**: Indexes on varchar columns may grow or shrink depending on
   the data distribution. ASCII-heavy → smaller indexes. CJK-heavy → larger.

5. **Client driver support**: The TDS client must negotiate `UTF8_SUPPORT` and correctly
   interpret the `fUTF8` collation bit. The ms package does not currently do this
   (see recommendations below).

6. **Stored procedures and views**: Any proc or view that casts between varchar and
   nvarchar, or uses collation-sensitive comparisons, should be tested.

### Recommendations for the ms Package

To fully support UTF-8 collations, the ms package needs:

1. **Parse the collation bytes**: Decode the 5-byte `COLLATION` from column metadata.
   Check the `fUTF8` bit to determine whether varchar data is UTF-8 or code-page encoded.

2. **Negotiate UTF8_SUPPORT**: Send FeatureExt `0x0A` with data byte `0x01` during
   LOGIN7. Handle the `FEATUREEXTACK` response. This tells the server the client can
   handle UTF-8 varchar data.

3. **Set collation on parameters**: Instead of hardcoding `{0x09, 0x04, 0xD0, 0x00, 0x34}`,
   either use the server's negotiated collation or allow callers to specify a collation.
   For UTF-8 targets, the collation bytes should have the `fUTF8` bit set.

4. **Optional: default to varchar instead of nvarchar**: When UTF-8 is negotiated, Go
   string parameters could map to `varchar` (UTF-8) instead of `nvarchar` (UTF-16),
   eliminating the `uconv.Encode.FromString()` conversion entirely. This should be
   opt-in via a connection config flag to avoid breaking existing behavior.

5. **Code page transcoding**: For non-UTF-8 varchar columns (the `fUTF8` bit is clear),
   the ms package should transcode from the code page (derived from LCID/SortId) to
   UTF-8. Today these bytes are passed through raw, which is correct only for ASCII
   data or when the code page happens to be compatible.
