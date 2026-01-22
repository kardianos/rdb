### Grammar Definition for Token Description

The Tabular Data Stream consists of a variety of messages. Each message
consists of a set of bytes transmitted in a predefined order. This
predefined order or grammar can be specified by using Augmented
Backus-Naur Form (ABNF)
[\[RFC4234\]](https://go.microsoft.com/fwlink/?LinkId=90462). Details
can be found in the following subsections.

#### General Rules

Data structure encodings in TDS are defined in terms of the following
fundamental definitions.

**BIT**: A single bit value of either 0 or 1.

8.  BIT = %b0 / %b1

**BYTE**: An unsigned single byte (8-bit) value. The range is 0 to 255.

9.  BYTE = 8BIT

**BYTELEN**: An unsigned single byte (8-bit) value representing the
length of the associated data. The range is 0 to 255.

10. BYTELEN = BYTE

**USHORT**: An unsigned 2-byte (16-bit) value. The range is 0 to 65535.

11. USHORT = 2BYTE

**USHORT_MAX**: An unsigned 2-byte (16-bit) value representing the
maximum value of the associated data. The range is 65535 or greater.

12. USHORT_MAX = 2BYTE

**LONG**: A signed 4-byte (32-bit) value. The range is -(2\^31) to
(2\^31)-1.

13. LONG = 4BYTE

**ULONG**: An unsigned 4-byte (32-bit) value. The range is 0 to
(2\^32)-1.

14. ULONG = 4BYTE

**DWORD**: An unsigned 4-byte (32-bit) value. The range when used as a
numeric value is 0 to (2\^32)-1.

15. DWORD = 32BIT

**LONGLONG**: A signed 8-byte (64-bit) value. The range is -(2\^63) to
(2\^63)-1.

16. LONGLONG = 8BYTE

**ULONGLONG**: An unsigned 8-byte (64-bit) value. The range is 0 to
(2\^64)-1.

17. ULONGLONG = 8BYTE

**UCHAR**: An unsigned single byte (8-bit) value representing a
character. The range is 0 to 255.

18. UCHAR = BYTE

**USHORTLEN**: An unsigned 2-byte (16-bit) value representing the length
of the associated data. The range is 0 to 65535.

19. USHORTLEN = 2BYTE

**USHORTCHARBINLEN**: An unsigned 2-byte (16-bit) value representing the
length of the associated character or binary data. The range is 0 to
8000.

20. USHORTCHARBINLEN = 2BYTE

**LONGLEN**: A signed 4-byte (32-bit) value representing the length of
the associated data. The range is -(2\^31) to (2\^31)-1.

21. LONGLEN = 4BYTE

**ULONGLEN**: An unsigned 4-byte (32-bit) value representing the length
of the associated data. The range is 0 to (2\^32)-1.

22. ULONGLEN = 4BYTE

**ULONGLONGLEN**: An unsigned 8-byte (64-bit) value representing the
length of the associated data. The range is 0 to (2\^64)-1.

23. ULONGLONGLEN = 8BYTE

**PRECISION**: An unsigned single byte (8-bit) value representing the
precision of a numeric number.

24. PRECISION = 8BIT

**SCALE**: An unsigned single byte (8-bit) value representing the scale
of a numeric number.

25. SCALE = 8BIT

**GEN_NULL**: A single byte (8-bit) value representing a NULL value.

26. GEN_NULL = %x00

**CHARBIN_NULL**: A 2-byte (16-bit) or 4-byte (32-bit) value
representing a T-SQL NULL value for a character or binary data type.
Please refer to TYPE_VARBYTE (see section
[2.2.5.2.3](#Section_3f983fde0509485a8c40a9fa6679a828)) for additional
details.

27. CHARBIN_NULL = (%xFF %xFF) / (%xFF %xFF %xFF %xFF)

**FRESERVEDBIT**: A FRESERVEDBIT is a BIT value used for padding that
does not transmit information. FRESERVEDBIT fields SHOULD be set to %b0
and MUST be ignored on receipt.

28. FRESERVEDBIT = %b0

**FRESERVEDBYTE**: A FRESERVEDBYTE is a BYTE value used for padding that
does not transmit information. FRESERVEDBYTE fields SHOULD be set to
%x00 and MUST be ignored on receipt.

29. FRESERVEDBYTE = %x00

**UNICODECHAR**: A single
[**Unicode**](#gt_c305d0ab-8b94-461a-bd76-13b40cb8c4d8) character in
UCS-2 encoding, as specified in Unicode
[\[UNICODE\]](https://go.microsoft.com/fwlink/?LinkId=90550).

30. UNICODECHAR = 2BYTE

**Notes**

-   All integer types are represented in reverse byte order
    ([**little-endian**](#gt_079478cb-f4c5-4ce5-b72b-2144da5d2ce7))
    unless otherwise specified.

-   FRESERVEDBIT and FRESERVEDBYTE are often used to pad unused parts of
    a byte or bytes. The value of these reserved bits SHOULD be ignored.
    These elements are generally set to 0.

##### Least Significant Bit Order

Certain tokens possess rules that comprise an array of independent bits.
These are \"flag\" rules in which each bit is a flag indicating that a
specific feature or option is enabled/requested. Normally, the bit array
is arranged in least significant bit order (or typical array index
order) meaning that the first listed flag is placed in the least
significant bit position (identifying the least significant bit as one
would in an integer variable). For example, if *Fn* is the *nth* flag,
then the following rule definition:

31. FLAGRULE = F0 F1 F2 F3 F4 F5 F6 F7

would be observed on the wire in the natural value order
F7F6F5F4F3F2F1F0.

If the rule contains 16 bits, then the order of the bits observed on the
wire follows the
[**little-endian**](#gt_079478cb-f4c5-4ce5-b72b-2144da5d2ce7) byte
ordering. For example, the following rule definition:

32. FLAGRULE = F0 F1 F2 F3 F4 F5 F6 F7 F8 F9 F10 F11 F12 F13 F14 F15

has the following order on the wire: F7F6F5F4F3F2F1F0
F15F14F13F12F11F10F9F8.

##### Collation Rule Definition

The collation rule is used to specify collation information for
character data or metadata describing character data.[\<9\>](\l) This is
specified as part of the LOGIN7 (section
[2.2.6.4](#Section_773a62b6ee894c029e5e344882630aac)) message or part of
a column definition in server results containing character data. For
more information about column definition, see COLMETADATA (section
[2.2.7.4](#Section_58880b9f381c43b2bf8b0727a98c4f4c)).

33. LCID = 20BIT

    fIgnoreCase = BIT

    fIgnoreAccent = BIT

    fIgnoreWidth = BIT

    fIgnoreKana = BIT

    fBinary = BIT

    fBinary2 = BIT

    fUTF8 = BIT

    ColFlags = fIgnoreCase fIgnoreAccent fIgnoreKana

    fIgnoreWidth fBinary fBinary2 fUTF8

    FRESERVEDBIT

    Version = 4BIT

    SortId = BYTE

    COLLATION = LCID ColFlags Version SortId

A SQL collation is one of a predefined set of sort orders. The sort
orders are identified with non-zero SortId values described by
[\[MSDN-SQLCollation\]](https://go.microsoft.com/fwlink/?LinkId=119987).

For a SortId==0 collation, the LCID bits correspond to a LocaleId as
defined by the National Language Support (NLS) functions. For more
details, see
[\[MS-LCID\]](%5bMS-LCID%5d.pdf#Section_70feba9f294e491eb6eb56532684c37f).

**Notes**

-   ColFlags is represented in [least significant bit
    order](#Section_bbc22f15e1a04338a169c79819d39b1c).

-   A COLLATION[\<10\>](\l) value of 0x00 00 00 00 00 specifies a
    request for the use of raw collation.

#### Data Stream Types

##### Unknown Length Data Streams

Unknown length data streams can be used by tokenless data streams. It is
a stream of bytes. The number of bytes within the data stream is defined
in the packet header as specified in section
[2.2.3.1](#Section_7af536671b72470382587984e838f746).

49. BYTESTREAM = \*BYTE

    UNICODESTREAM = \*(2BYTE)

##### Variable-Length Data Streams

Variable-length data streams consist of a stream of characters or a
stream of bytes. The two types are similar, in that they both have a
length rule and a data rule.

**Characters**

Variable-length character streams are defined by a length field followed
by the data itself. There are two types of variable-length character
streams, each dependent on the size of the length field (for example, a
BYTE or USHORT). If the length field is zero, then no data follows the
length field.

51. B_VARCHAR = BYTELEN \*CHAR

    US_VARCHAR = USHORTLEN \*CHAR

Note that the lengths of B_VARCHAR and US_VARCHAR are given in
[**Unicode**](#gt_c305d0ab-8b94-461a-bd76-13b40cb8c4d8) characters.

**Generic Bytes**

Similar to the variable-length character stream, variable-length byte
streams are defined by a length field followed by the data itself. There
are three types of variable-length byte streams, each dependent on the
size of the length field (for example, a BYTE, USHORT, or LONG). If the
value of the length field is zero, then no data follows the length
field.

53. B_VARBYTE = BYTELEN \*BYTE

    US_VARBYTE = USHORTLEN \*BYTE

    L_VARBYTE = LONGLEN \*BYTE

##### Data Type Dependent Data Streams

Some messages contain variable data types. The actual type of a given
variable data type is dependent on the type of the data being sent
within the message as defined in the TYPE_INFO rule (section
[2.2.5.6](#Section_cbe9c510eae64b1f9893a098944d430a)).

For example, the RPCRequest message contains the TYPE_INFO and
TYPE_VARBYTE rules. These two rules contain data of a type that is
dependent on the actual type used in the value of the FIXEDLENTYPE or
VARLENTYPE rules of the TYPE_INFO rule.

Data type-dependent data streams occur in three forms: integers, fixed
and variable bytes, and partially length-prefixed bytes.

**Integers**

Data type-dependent integers can be either a BYTELEN, USHORTCHARBINLEN,
or LONGLEN in length. This length is dependent on the TYPE_INFO
associated with the message. If the data type (for example, FIXEDLENTYPE
or VARLENTYPE rule of the TYPE_INFO rule) is of type SSVARIANTTYPE,
TEXTTYPE, NTEXTTYPE, or IMAGETYPE, the integer length is LONGLEN. If the
data type is BIGCHARTYPE, BIGVARCHARTYPE, NCHARTYPE, NVARCHARTYPE,
BIGBINARYTYPE, or BIGVARBINARYTYPE, the integer length is
USHORTCHARBINLEN. For all other data types, the integer length is
BYTELEN.

56. TYPE_VARLEN = BYTELEN

    /

    USHORTCHARBINLEN

    /

    LONGLEN

**Fixed and Variable Bytes**

The data type to be used in a data type-dependent byte stream is defined
by the TYPE_INFO rule associated with the message.

For variable-length types, with the exception of PLP (see Partially
Length-prefixed Bytes below), the TYPE_VARLEN value defines the length
of the data to follow. As described above, the TYPE_INFO rule defines
the type of TYPE_VARLEN (for example BYTELEN, USHORTCHARBINLEN, or
LONGLEN).

For fixed-length types, the TYPE_VARLEN rule is not present. In these
cases, the number of bytes to be read is determined by the TYPE_INFO
rule. For example, if \"INT2TYPE\" is specified as the value for the
FIXEDLENTYPE rule of the TYPE_INFO rule, 2 bytes are read because
\"INT2TYPE\" is always 2 bytes in length. For more details, see [Data
Types Definitions](#Section_ffb02215af074b5085451fd522106c68).

The data following this can be a stream of bytes or a NULL value. The
2-byte CHARBIN_NULL rule is used for BIGCHARTYPE, BIGVARCHARTYPE,
NCHARTYPE, NVARCHARTYPE, BIGBINARYTYPE, and BIGVARBINARYTYPE types, and
the 4-byte CHARBIN_NULL rule is used for TEXTTYPE, NTEXTTYPE, and
IMAGETYPE. The GEN_NULL rule applies to all other types aside from PLP:

61. TYPE_VARBYTE = GEN_NULL / CHARBIN_NULL / PLP_BODY

    / (\[TYPE_VARLEN\] \*BYTE)

**Partially Length-prefixed Bytes**

Unlike fixed or variable byte stream formats, Partially length-prefixed
bytes (PARTLENTYPE), introduced in TDS 7.2, do not require the full data
length to be specified before the actual data is streamed out. Thus, it
is ideal for those applications where the data length is not known
upfront (that is, xml serialization). A value sent as PLP can be either
NULL, a length followed by chunks (as defined by PLP_CHUNK), or an
unknown length token followed by chunks, which MUST end with a
PLP_TERMINATOR. The rule below describes the stream format (for example,
the format of a singleton PLP value):

63. PLP_BODY= PLP_NULL

    /

    ((ULONGLONGLEN / UNKNOWN_PLP_LEN)

    \*PLP_CHUNK PLP_TERMINATOR)

    PLP_NULL = %xFFFFFFFFFFFFFFFF

    UNKNOWN_PLP_LEN = %xFFFFFFFFFFFFFFFE

    PLP_CHUNK = ULONGLEN 1\*BYTE

    PLP_TERMINATOR = %x00000000

**Notes**

-   TYPE_INFO rule specifies a Partially Length-prefixed Data type
    (PARTLENTYPE, see
    [2.2.5.4.4](#Section_7d26a257083e409b81ba897e0c672be0)).

-   In the UNKNOWN_PLP_LEN case, the data is represented as a series of
    zero or more chunks, each consisting of the length field followed by
    length bytes of data (see the PLP_CHUNK rule). The data is
    terminated by PLP_TERMINATOR (which is essentially a zero-length
    chunk).

-   In the actual data length case, the ULONGLONGLEN specifies the
    length of the data and is followed by any number of PLP_CHUNKs
    containing the data. The length of the data specified by
    ULONGLONGLEN is used as a hint for the receiver. The receiver SHOULD
    validate that the length value specified by ULONGLONGLEN matches the
    actual data length.

#### Packet Data Stream Headers - ALL_HEADERS Rule Definition

Message streams can be preceded by a variable number of headers as
specified by the ALL_HEADERS rule. The ALL_HEADERS rule, Query
Notifications (section
[2.2.5.3.1](#Section_e168d373a7b741aab6ca25985466a7e0)), and Transaction
Descriptor (section
[2.2.5.3.2](#Section_4257dd95ef6c4621b75d270738487d68)) were introduced
in TDS 7.2. Trace Activity (section
[2.2.5.3.3](#Section_6e9f106bdf6e4cbea6eb45ceb10c63be)) was introduced
in TDS 7.4.

The list of headers that are applicable to the different types of
messages are described in the following table.

Stream headers MUST be present only in the first packet of requests that
span more than one packet. The ALL_HEADERS rule applies only to the
three client request types defined in the table below and MUST NOT be
included for other request types. For the applicable request types, each
header MUST appear at most once in the stream or packet\'s ALL_HEADERS
field.

  -----------------------------------------------------------------------------
  Header            Value   SQLBatch   RPCRequest   TransactionManagerRequest
  ----------------- ------- ---------- ------------ ---------------------------
  Query             0x00 01 Optional   Optional     Disallowed
  Notifications                                     

  Transaction       0x00 02 Required   Required     Required
  Descriptor                                        

  Trace Activity    0x00 03 Optional   Optional     Optional
  -----------------------------------------------------------------------------

**Stream-Specific Rules:**

75. TotalLength = DWORD ;including itself

    HeaderLength = DWORD ;including itself

    HeaderType = USHORT;

    HeaderData = \*BYTE

    Header = HeaderLength HeaderType HeaderData

**Stream Definition:**

80. ALL_HEADERS = TotalLength 1\*Header

  --------------------------------------------------------------------------
  Parameter      Description
  -------------- -----------------------------------------------------------
  TotalLength    Total length of ALL_HEADERS stream.

  HeaderLength   Total length of an individual header.

  HeaderType     The type of header, as defined by the value field in the
                 preceding table.

  HeaderData     The data stream for the header. See header definitions in
                 the following subsections.

  Header         A structure containing a single header.
  --------------------------------------------------------------------------

##### Query Notifications Header

This packet data stream header allows the client to specify that a
notification is to be supplied on the results of the request. The
contents of the header specify the information necessary for delivery of
the notification. For more information about [**query
notifications**](#gt_62a2f252-bd68-4d77-a751-d6ff27010678)[\<11\>](\l)
functionality for a database server that supports SQL, see
[\[MSDN-QUERYNOTE\]](https://go.microsoft.com/fwlink/?LinkId=119984).

**Stream Specific Rules:**

81. NotifyId = USHORT UNICODESTREAM ; user specified value

    ; when subscribing to

    ; query notifications

    SSBDeployment = USHORT UNICODESTREAM

    NotifyTimeout = ULONG ; duration in which the query

    ; notification subscription

    ; is valid

The USHORT field defined within the NotifyId and SSBDeployment rules
specifies the length, in bytes, of the actual data value, defined by the
UNICODESTREAM, that follows it.[\<12\>](\l) The time unit of
NotifyTimeout is milliseconds.

**Stream Definition:**

88. HeaderData = NotifyId

    SSBDeployment

    \[NotifyTimeout\]

##### Transaction Descriptor Header

This packet data stream contains information regarding transaction
descriptor and number of outstanding requests as they apply to
[**Multiple Active Result Sets
(MARS)**](#gt_762fe1e3-0979-4402-b963-1e9150de133d)
[\[MSDN-MARS\]](https://go.microsoft.com/fwlink/?LinkId=98459).

The TransactionDescriptor MUST be 0, and OutstandingRequestCount MUST be
1 if the connection is operating in AutoCommit mode. For more
information about autocommit transactions, see
[\[MSDN-Autocommit\]](https://go.microsoft.com/fwlink/?LinkId=145156).

**Stream-Specific Rules:**

91. OutstandingRequestCount = DWORD ; number of requests currently
    active on

    ; the connection

    TransactionDescriptor = ULONGLONG ; for each connection, a number
    that uniquely

    ; identifies the transaction with which the

    ; request is associated; initially generated

    ; by the server when a new transaction is

    ; created and returned to the client as part

    ; of the ENVCHANGE token stream

For more information about processing the Transaction Descriptor header,
see section [2.2.6.9](#Section_0fb28ba5ddcb4d0295c3aa5b05ec6092).

**Stream Definition:**

99. HeaderData = TransactionDescriptor

    OutstandingRequestCount

##### Trace Activity Header

This packet data stream contains a client trace activity ID intended to
be used by the server for debugging purposes, to allow correlating the
server\'s processing of the request with the client request.

A client MUST NOT send a Trace Activity header when the negotiated TDS
major version is less than 7.4. If the negotiated TDS major version is
less than TDS 7.4 and the server receives a Trace Activity header token,
the server MUST reject the request with a TDS protocol error.

**Stream-Specific Rules:**

101. GUID_ActivityID = 16BYTE ; client application activity id

     ; used for debugging purposes

     ActivitySequence = ULONG ; client application activity sequence

     ; used for debugging purposes

     ActivityId = GUID_ActivityID

     ActivitySequence

**Stream Definition:**

107. HeaderData = ActivityId

#### Data Type Definitions

The subsections within this section describe the different sets of data
types and how they are categorized. Specifically, data values are
interpreted and represented in association with their data type. Details
about each data type categorization are described in the following
sections.

##### Zero-Length Data Types

The zero-length data types include the following type.

108. NULLTYPE = 0x1F ; Null

There is no data associated with NULLTYPE.\<13\> For more details, see
section [2.2.4.2.1.1](#Section_9b571ae55d0048748cf798f30ca69dbd).

##### Fixed-Length Data Types

The fixed-length data types include the following types.

109. INT1TYPE = %x30 ; TinyInt

     BITTYPE = %x32 ; Bit

     INT2TYPE = %x34 ; SmallInt

     INT4TYPE = %x38 ; Int

     DATETIM4TYPE = %x3A ; SmallDateTime

     FLT4TYPE = %x3B ; Real

     MONEYTYPE = %x3C ; Money

     DATETIMETYPE = %x3D ; DateTime

     FLT8TYPE = %x3E ; Float

     MONEY4TYPE = %x7A ; SmallMoney

     INT8TYPE = %x7F ; BigInt

     DECIMALTYPE = %x37 ; Decimal (legacy support)

     NUMERICTYPE = %x3F ; Numeric (legacy support)

     FIXEDLENTYPE = INT1TYPE

     /

     BITTYPE

     /

     INT2TYPE

     /

     INT4TYPE

     /

     DATETIM4TYPE

     /

     FLT4TYPE

     /

     MONEYTYPE

     /

     DATETIMETYPE

     /

     FLT8TYPE

     /

     MONEY4TYPE

     /

     INT8TYPE

Non-nullable values are returned using these fixed-length data types.
For the fixed-length data types, the length of data is predefined by the
type. There is no TYPE_VARLEN field in the TYPE_INFO rule for these
types. In the TYPE_VARBYTE rule for these types, the TYPE_VARLEN field
is BYTELEN, and the value is 1 for INT1TYPE/BITTYPE, 2 for INT2TYPE, 4
for INT4TYPE/DATETIM4TYPE/FLT4TYPE/MONEY4TYPE, and 8 for
MONEYTYPE/DATETIMETYPE/FLT8TYPE/INT8TYPE. The value represents the
number of bytes of data to be followed. The SQL data types of the
corresponding fixed-length data types are in the comment part of each
data type.

##### Variable-Length Data Types

The data type token values defined in this section have a length value
associated with the data type because the data values corresponding to
these data types are represented by a variable number of bytes.

144. GUIDTYPE = %x24 ; UniqueIdentifier

     INTNTYPE = %x26 ; (see below)

     BITNTYPE = %x68 ; (see below)

     DECIMALNTYPE = %x6A ; Decimal

     NUMERICNTYPE = %x6C ; Numeric

     FLTNTYPE = %x6D ; (see below)

     MONEYNTYPE = %x6E ; (see below)

     DATETIMNTYPE = %x6F ; (see below)

     DATENTYPE = %x28 ; (introduced in TDS 7.3)

     TIMENTYPE = %x29 ; (introduced in TDS 7.3)

     DATETIME2NTYPE = %x2A ; (introduced in TDS 7.3)

     DATETIMEOFFSETNTYPE = %x2B ; (introduced in TDS 7.3)

     CHARTYPE = %x2F ; Char (legacy support)

     VARCHARTYPE = %x27 ; VarChar (legacy support)

     BINARYTYPE = %x2D ; Binary (legacy support)

     VARBINARYTYPE = %x25 ; VarBinary (legacy support)

     BIGVARBINARYTYPE = %xA5 ; VarBinary

     BIGVARCHARTYPE = %xA7 ; VarChar

     BIGBINARYTYPE = %xAD ; Binary

     BIGCHARTYPE = %xAF ; Char

     NVARCHARTYPE = %xE7 ; NVarChar

     NCHARTYPE = %xEF ; NChar

     XMLTYPE = %xF1 ; XML (introduced in TDS 7.2)

     UDTTYPE = %xF0 ; CLR UDT (introduced in TDS 7.2)

     TEXTTYPE = %x23 ; Text

     IMAGETYPE = %x22 ; Image

     NTEXTTYPE = %x63 ; NText

     SSVARIANTTYPE = %x62 ; sql_variant (introduced in TDS 7.2)

     JSONTYPE = %xF4

     VECTORTYPE = %XF5

     BYTELEN_TYPE = GUIDTYPE

     /

     INTNTYPE

     /

     DECIMALTYPE

     /

     NUMERICTYPE

     /

     BITNTYPE

     /

     DECIMALNTYPE

     /

     NUMERICNTYPE

     /

     FLTNTYPE

     /

     MONEYNTYPE

     /

     DATETIMNTYPE

     /

     DATENTYPE

     /

     TIMENTYPE

     /

     DATETIME2NTYPE

     /

     DATETIMEOFFSETNTYPE

     /

     CHARTYPE

     /

     VARCHARTYPE

     /

     BINARYTYPE

     /

     VARBINARYTYPE ; the length value associated

     ; with these data types is

     ; specified within a BYTE

For DECIMALNTYPE and NUMERICNTYPE, the only valid lengths are 0x05,
0x09, 0x0D, and 0x11 for non-NULL instances.

For MONEYNTYPE, the only valid lengths are 0x04 and 0x08, which map to
smallmoney and money SQL data types respectively.

For DATETIMNTYPE, the only valid lengths are 0x04 and 0x08, which map to
smalldatetime and datetime SQL data types respectively.

For INTNTYPE, the only valid lengths are 0x01, 0x02, 0x04, and 0x08,
which map to tinyint, smallint, int, and bigint SQL data types
respectively.

For FLTNTYPE, the only valid lengths are 0x04 and 0x08, which map to
7-digit precision float and 15-digit precision float SQL data types
respectively.

For GUIDTYPE, the only valid lengths are 0x10 for non-null instances and
0x00 for NULL instances.

For BITNTYPE, the only valid lengths are 0x01 for non-null instances and
0x00 for NULL instances.

For DATENTYPE, the only valid lengths are 0x03 for non-NULL instances
and 0x00 for NULL instances.

For TIMENTYPE, the only valid lengths (along with the associated scale
value) are:

  --------------------------------------------------------------------------
  SCALE       1        2        3        4        5        6        7
  ----------- -------- -------- -------- -------- -------- -------- --------
  LENGTH      0x03     0x03     0x04     0x04     0x05     0x05     0x05

  --------------------------------------------------------------------------

For DATETIME2NTYPE, the only valid lengths (along with the associated
scale value) are:

  --------------------------------------------------------------------------
  SCALE       1        2        3        4        5        6        7
  ----------- -------- -------- -------- -------- -------- -------- --------
  LENGTH      0x06     0x06     0x07     0x07     0x08     0x08     0x08

  --------------------------------------------------------------------------

For DATETIMEOFFSETNTYPE, the only valid lengths (along with the
associated scale value) are:

  --------------------------------------------------------------------------
  SCALE       1        2        3        4        5        6        7
  ----------- -------- -------- -------- -------- -------- -------- --------
  LENGTH      0x08     0x08     0x09     0x09     0x0A     0x0A     0x0A

  --------------------------------------------------------------------------

Exceptions are thrown when invalid lengths are presented to the server
during BulkLoadBCP and RPC requests.

For all variable length data types, the value is 0x00 for NULL
instances.

For VECTORTYPE, the valid scale values are defined by the Vector
Dimension Type Identifier values (see section
[2.2.5.5.7.4](#Section_3404c02191e4445282d69fee2403cb86)).

214. USHORTLEN_TYPE = BIGVARBINARYTYPE

     /

     BIGVARCHARTYPE

     /

     BIGBINARYTYPE

     /

     BIGCHARTYPE

     /

     NVARCHARTYPE

     /

     NCHARTYPE

     /

     VECTORTYPE ; the length value associated with

     ; these data types is specified

     ; within a USHORT

     LONGLEN_TYPE = IMAGETYPE

     /

     NTEXTTYPE

     /

     SSVARIANTTYPE

     /

     TEXTTYPE

     /

     XMLTYPE

     /

     JSONTYPE ; the length value associated with

     ; these data types is specified

     ; within a LONG

**Notes**

-   MaxLength for an SSVARIANTTYPE is 8009 (8000 for strings). For more
    details, see section
    [2.2.5.5.4](#Section_2435e85d9e61492cacb2627ffccb5b92).

-   XMLTYPE is only a valid LONGLEN_TYPE for BulkLoadBCP.

MaxLength for an SSVARIANTTYPE is 8009 (string of 8000 bytes).

245. VARLENTYPE = BYTELEN_TYPE

     /

     USHORTLEN_TYPE

     /

     LONGLEN_TYPE

Nullable values are returned by using the INTNTYPE, BITNTYPE, FLTNTYPE,
GUIDTYPE, MONEYNTYPE, and DATETIMNTYPE tokens that use the length byte
to specify the length of the value or GEN_NULL as appropriate.

There are two types of variable-length data types. These are real
variable-length data types, like char and binary, and nullable data
types, which have either a normal fixed length that corresponds to their
type or to a special length if null.

Char and binary data types have values that either are null or are 0 to
65534 (0x0000 to 0xFFFE) bytes of data. Null is represented by a length
of 65535 (0xFFFF). A non-nullable char or binary can still have a length
of zero (for example, an empty value). A program that MUST pad a value
to a fixed length adds blanks to the end of a char and adds binary zeros
to the end of a binary.

Text and image data types have values that either are null or are 0 to 2
gigabytes (0x00000000 to 0x7FFFFFFF bytes) of data. Null is represented
by a length of -1 (0xFFFFFFFF). No other length specification is
supported.

Other nullable data types have a length of 0 when they are null.

##### Partially Length-Prefixed Data Types

The data value corresponding to the set of data types defined in this
section follows the rule defined in the partially length-prefixed stream
definition (section
[2.2.5.2.3](#Section_3f983fde0509485a8c40a9fa6679a828)).

250. PARTLENTYPE = XMLTYPE

     /

     BIGVARCHARTYPE

     /

     BIGVARBINARYTYPE

     /

     NVARCHARTYPE

     /

     UDTTYPE

     /

     JSONTYPE

BIGVARCHARTYPE, BIGVARBINARYTYPE, and NVARCHARTYPE can represent two
types each:

-   The regular type with a known maximum size range defined by
    USHORTLEN_TYPE. For BIGVARCHARTYPE and BIGVARBINARYTYPE, the range
    is 0 to 8000. For NVARCHARTYPE, the range is 0 to 4000.

-   A type with unlimited max size, known as varchar(max),
    varbinary(max) and nvarchar(max), which has a max size of 0xFFFF,
    defined by PARTLENTYPE. This class of types was introduced in TDS
    7.2.

#### Data Type Details

The subsections within this section specify the formats in which values
of system data types are serialized in TDS.

##### System Data Type Values

The subsections within this section specify the formats in which values
of various common system data types are serialized in TDS.

###### Integers

All integer types are represented in reverse byte order
([**little-endian**](#gt_079478cb-f4c5-4ce5-b72b-2144da5d2ce7)) unless
otherwise specified. Each integer takes a whole number of bytes as
follows:

> **bit:** 1 byte
>
> **tinyint:** 1 byte
>
> **smallint:** 2 bytes
>
> **int:** 4 bytes
>
> **bigint:** 8 bytes

###### Time Stamps

**timestamp/rowversion** is represented as an 8-byte binary sequence
with no particular interpretation.

###### Character and Binary Strings

See Variable-Length Data
Types (section [2.2.5.4.3)](#Section_ce3183a69d8947e8a02fde5a1a1303de)
and Partially Length-Prefixed Data
Types (section [2.2.5.4.4)](#Section_7d26a257083e409b81ba897e0c672be0).

###### Fixed-Point Numbers

**smallmoney** is represented as a 4-byte signed integer. The TDS value
is the **smallmoney** value multiplied by 10^4^.

**money** is represented as an 8-byte signed integer. The TDS value is
the **money** value multiplied by 10^4^. The 8-byte signed integer
itself is represented in the following sequence:

-   One 4-byte integer that represents the more significant half.

-   One 4-byte integer that represents the less significant half.

###### Floating-Point Numbers

**float**(*n*) follows the 32-bit
[\[IEEE754\]](https://go.microsoft.com/fwlink/?LinkId=89903) binary
specification when *n* \<= 24 and the 64-bit \[IEEE754\] binary
specification when 25 \<= *n* \<= 53.

###### Decimals and Numerics

Decimal or Numeric is defined as **decimal**(*p*, *s*) or
**numeric**(*p*, *s*), where *p* is the precision and *s* is the scale.
The value is represented in the following sequence:

-   One 1-byte unsigned integer that represents the sign of the decimal
    value as follows:

    -   0 means negative.

    -   1 means nonnegative.

-   One 4-, 8-, 12-, or 16-byte signed integer that represents the
    decimal value multiplied by 10^s^. The maximum size of this integer
    is determined based on *p* as follows:

    -   4 bytes if 1 \<= *p* \<= 9.

    -   8 bytes if 10 \<= *p* \<= 19.

    -   12 bytes if 20 \<= *p* \<= 28.

    -   16 bytes if 29 \<= *p* \<= 38.

The actual size of this integer could be less than the maximum size,
depending on the value. In all cases, the integer part MUST be 4, 8, 12,
or 16 bytes.

###### GUIDs

**uniqueidentifier** is represented as a 16-byte binary sequence with no
specific interpretation.

###### Dates and Times

**smalldatetime** is represented in the following sequence:

-   One 2-byte unsigned integer that represents the number of days since
    January 1, 1900.

-   One 2-byte unsigned integer that represents the number of minutes
    elapsed since 12 AM that day.

**datetime** is represented in the following sequence:

-   One 4-byte signed integer that represents the number of days since
    January 1, 1900. Negative numbers are allowed to represent dates
    since January 1, 1753.

-   One 4-byte unsigned integer that represents the number of one
    three-hundredths of a second (300 counts per second) elapsed since
    12 AM that day.

**date** is represented as one 3-byte unsigned integer that represents
the number of days since January 1, year 1.

**time**(*n*) is represented as one unsigned integer that represents the
number of 10^-n^ second increments since 12 AM within a day. The length,
in bytes, of that integer depends on the scale *n* as follows:

-   3 bytes if 0 \<= *n* \< = 2.

-   4 bytes if 3 \<= *n* \< = 4.

-   5 bytes if 5 \<= *n* \< = 7.

**datetime2**(*n*) is represented as a concatenation of **time**(*n*)
followed by **date** as specified above.

**datetimeoffset**(*n*) is represented as a concatenation of
**datetime2**(*n*) followed by one 2-byte signed integer that represents
the time zone offset as the number of minutes from UTC. The time zone
offset MUST be between -840 and 840.

##### Common Language Runtime (CLR) Instances

The following data type definition stream is used for UDT_INFO in
TYPE_INFO. This data type was introduced in TDS 7.2.

261. DB_NAME = B_VARCHAR ; database name of the UDT

     SCHEMA_NAME = B_VARCHAR ; schema name of the UDT

     TYPE_NAME = B_VARCHAR ; type name of the UDT

     MAX_BYTE_SIZE = USHORT ; max length in bytes

     ASSEMBLY_QUALIFIED_NAME = US_VARCHAR ; name of the CLR assembly

     UDT_METADATA = ASSEMBLY_QUALIFIED_NAME

     UDT_INFO_IN_COLMETADATA = MAX_BYTE_SIZE

     DB_NAME

     SCHEMA_NAME

     TYPE_NAME

     UDT_METADATA

     UDT_INFO_IN_RPC = DB_NAME ; database name of the UDT

     SCHEMA_NAME ; schema name of the UDT

     TYPE_NAME ; type name of the UDT

     UDT_INFO = UDT_INFO_IN_COLMETADATA ;when sent as part of
     COLMETADATA

     /

     UDT_INFO_IN_RPC ;when sent as part of RPC call

MAX_BYTE_SIZE is only sent from the server to the client in COLMETADATA
(section [2.2.7.4](#Section_58880b9f381c43b2bf8b0727a98c4f4c)) and is an
unsigned short with a value within the range 1 to 8000 or 0xFFFF. The
value 0xFFFF signifies the maximum LOB size indicating a UDT with a
maximum size greater than 8000 bytes (also referred to as a Large UDT;
introduced in TDS 7.3). MAX_BYTE_SIZE is not sent to the server as part
of RPC calls.

**Note**  UserType in the COLMETADATA stream is either 0x0000 or
0x00000000 for UDTs, depending on the TDS version that is used. The
actual data value format associated with a UDT data type definition
stream is specified in
[\[MS-SSCLRT\]](%5bMS-SSCLRT%5d.pdf#Section_77460aa98c2f4449a65e1d649ebd77fa).

##### XML Values

This section defines the XML data type definition stream, which was
introduced in TDS 7.2.

283. SCHEMA_PRESENT= BYTE;

     DbName = B_VARCHAR

     OWNING_SCHEMA = B_VARCHAR

     XML_SCHEMA_COLLECTION = US_VARCHAR

     XML_INFO = SCHEMA_PRESENT

     \[DbName OWNING_SCHEMA

     XML_SCHEMA_COLLECTION\]

SCHEMA_PRESENT specifies \"0x01\" if the type has an associated schema
collection and DbName, OWNING_SCHEMA and XML_SCHEMA_COLLECTION MUST be
included in the stream, or \'0x00\' otherwise.

DbName specifies the name of the database where the schema collection is
defined.

OWNING_SCHEMA specifies the name of the relational schema containing the
schema collection.

XML_SCHEMA_COLLECTION specifies the name of the XML schema collection to
which the type is bound.

**Note**  The actual data value format that is associated with an XML
data type definition stream uses the binary XML structure format, as
specified in
[\[MS-BINXML\]](%5bMS-BINXML%5d.pdf#Section_11ab6e8d247244d1a9e6bddf000e12f6).[\<14\>](\l)

##### sql_variant Values

The SSVARIANTTYPE is a special data type that acts as a place holder for
other data types. When a SSVARIANTTYPE is filled with a data value, it
takes on properties of the base data type that represents the data
value. To support this dynamic change, for those that are not NULL
(GEN_NULL) the SSVARIANTTYPE instance has an SSVARIANT_INSTANCE internal
structure according to the following definition.

291. VARIANT_BASETYPE = BYTE ; data type definition

     VARIANT_PROPBYTES = BYTE ; see below

     VARIANT_PROPERTIES = \*BYTE ; see below

     VARIANT_DATAVAL = 1\*BYTE ; actual data value

     SSVARIANT_INSTANCE = VARIANT_BASETYPE

     VARIANT_PROPBYTES

     VARIANT_PROPERTIES

     VARIANT_DATAVAL

VARIANT_BASETYPE is the TDS token of the base type.

  --------------------------------------------------------------------------
  VARIANT_BASETYPE                  VARIANT_PROPBYTES   VARIANT_PROPERTIES
  --------------------------------- ------------------- --------------------
  GUIDTYPE, BITTYPE, INT1TYPE,      0                   \<not specified\>
  INT2TYPE, INT4TYPE, INT8TYPE,                         
  DATETIMETYPE, DATETIM4TYPE,                           
  FLT4TYPE, FLT8TYPE, MONEYTYPE,                        
  MONEY4TYPE, DATENTYPE                                 

  TIMENTYPE, DATETIME2NTYPE,        1                   1 byte specifying
  DATETIMEOFFSETNTYPE                                   scale

  BIGVARBINARYTYPE, BIGBINARYTYPE   2                   2 bytes specifying
                                                        max length

  NUMERICNTYPE, DECIMALNTYPE        2                   1 byte for precision
                                                        followed by 1 byte
                                                        for scale

  BIGVARCHARTYPE, BIGCHARTYPE,      7                   5-byte COLLATION,
  NVARCHARTYPE, NCHARTYPE                               followed by a 2-byte
                                                        max length
  --------------------------------------------------------------------------

**Note** Data types cannot be NULL when inside a sql_variant. If the
value is NULL, the sql_variant itself has to be NULL, but it is not
allowed to specify a non-null sql_variant instance and have a NULL value
wrapped inside it. A raw collation SHOULD NOT be specified within a
sql_variant.[\<15\>](\l)

##### Table Valued Parameter (TVP) Values

Table Valued Parameters (or User Defined Table Type, as this type is
known on the server) encapsulate an entire table of data with 1 to 1024
columns and an arbitrary number of rows. At the present time, TVPs are
permitted to be used only as input parameters and do not appear as
output parameters or in [**result
set**](#gt_c8a27238-8ccc-442b-9604-75f74d3e6b3d) columns.

TVPs MUST be sent only by a TDS client that reports itself as a TDS
major version 7.3 or later. If a client reporting itself as older than
TDS 7.3 attempts to send a TVP, the server MUST reject the request with
a TDS protocol error.

###### Metadata

300. TVPTYPE = %xF3

     TVP_TYPE_INFO = TVPTYPE

     TVP_TYPENAME

     TVP_COLMETADATA

     \[TVP_ORDER_UNIQUE\]

     \[TVP_COLUMN_ORDERING\]

     TVP_END_TOKEN

     \*TVP_ROW

     TVP_END_TOKEN

  -----------------------------------------------------------------------
  Parameter                            Description
  ------------------------------------ ----------------------------------
  TVPTYPE                              %xF3

  TVP_TYPENAME                         Type name of the TVP

  TVP_COLMETADATA                      Column-specific metadata

  \[TVP_ORDER_UNIQUE\]                 Optional metadata token

  \[TVP_COLUMN_ORDERING\]              Optional metadata token

  TVP_END_TOKEN                        End optional metadata

  \*TVP_ROW                            0..N TVP_ROW tokens

  TVP_END_TOKEN                        End of rows
  -----------------------------------------------------------------------

**TVP_TYPENAME definition**

309. DbName = B_VARCHAR ; Database where TVP type resides

     OwningSchema = B_VARCHAR ; Schema where TVP type resides

     TypeName = B_VARCHAR ; TVP type name

     TVP_TYPENAME = DbName

     OwningSchema

     TypeName

**TVP_COLMETADATA definition**

315. fNullable = BIT ; Column is nullable - %x01

     fCaseSen = BIT ; Column is case-sensitive - %x02

     usUpdateable = 2BIT ; 2-bit value, one of:

     ; 0 = ReadOnly - %x00

     ; 1 = ReadWrite - %x04

     ; 2 = Unknown - %x08

     fIdentity = BIT ; Column is identity column - %x10

     fComputed = BIT ; Column is computed - %x20

     usReservedODBC = 2BIT ; Reserved bits for ODBC - %x40+80

     fFixedLenCLRType = BIT ; Fixed length CLR type - %x100

     fDefault = BIT ; Column is default value - %x200

     usReserved = 6BIT ; Six leftover reserved bits

     Flags = fNullable

     fCaseSen

     usUpdateable

     fIdentity

     fComputed

     usReservedODBC

     fFixedLenCLRType

     fDefault

     usReserved

     Count = USHORT ; Column count up to 1024 max

     ColName = B_VARCHAR ; Name of column

     UserType = ULONG ; UserType of column

     TvpColumnMetaData = UserType

     Flags

     TYPE_INFO

     ColName ; Column metadata instance

     TVP_NULL_TOKEN = %xFFFF

     TVP_COLMETADATA = TVP_NULL_TOKEN / ( Count (\<Count\>
     TvpColumnMetaData) )

DbName, OwningSchema, and TypeName are limited to 128
[**Unicode**](#gt_c305d0ab-8b94-461a-bd76-13b40cb8c4d8) characters max
identifier length.

DbName MUST be zero-length; only OwningSchema and TypeName can be
specified. DbName, OwningSchema, and TypeName are all optional fields
and might ALL contain zero length strings. Client SHOULD follow these
two rules:

-   If the TVP is a parameter to a [**stored
    procedure**](#gt_324d32b3-f4f3-41c9-b695-78c498094fb7) or function
    where parameter metadata is available on the server side, the client
    can send all zero-length strings for TVP_TYPENAME.

-   If the TVP is a parameter to an ad-hoc [**SQL
    statement**](#gt_dc5ca224-43ec-4b44-9dba-726d6fd6057d), parameter
    metadata information is not available on a stored procedure or
    function on the server. In this case, the client is responsible to
    send sufficient type information with the TVP to allow the server to
    resolve the TVP type from sys.types. Failure to send needed type
    information in this case results in complete failure of RPC call
    prior to execution.

Only one new flag, fDefault, is added here from existing COLMETADATA
(section [2.2.7.4](#Section_58880b9f381c43b2bf8b0727a98c4f4c)). ColName
MUST be a zero-length string in the TVP.

**Additional details about input TVPs and usage of flags**

-   For an input TVP, if the fDefault flag is set on a column, then the
    client MUST NOT emit the corresponding TvpColumnData data for the
    associated column when sending each TVP_ROW.

-   For an input TVP, the fCaseSen, usUpdateable, and fFixedLenCLRType
    flags are ignored.

-   usUpdateable is ignored by server on input, it is \"calculated\"
    metadata.

-   The fFixedLenCLRType flag is not used by the server.

-   Output TVPs are not supported.

**TVP Flags Usage Chart**

  -----------------------------------------------------------------------
  Flag                 Input behavior
  -------------------- --------------------------------------------------
  fNullable            Allowed

  fCaseSen             Ignored

  usUpdateable         Ignored

  fIdentity            Allowed

  fComputed            Allowed

  usReservedODBC       Ignored

  fFixedLenCLRType     Ignored

  fDefault             Allowed (if set, data not sent in TvpColumnData)

  usReserved           Ignored
  -----------------------------------------------------------------------

###### Optional Metadata Tokens

**TVP_ORDER_UNIQUE definition**

349. TVP_ORDER_UNIQUE_TOKEN = %x10

     Count = USHORT ; Count of ColNums to follow

     ColNum = USHORT ; A single-column ordinal

     fOrderAsc = BIT ; Column-ordered ascending -- %x01

     fOrderDesc = BIT ; Column-ordered descending -- %x02

     fUnique = BIT ; Column is in unique set -- %x04

     Reserved1 = 5BIT ; Five reserved bits

     OrderUniqueFlags = fOrderAsc

     fOrderDesc

     fUnique

     Reserved1

     TVP_ORDER_UNIQUE = TVP_ORDER_UNIQUE_TOKEN

     ( Count (\<Count\> (ColNum OrderUniqueFlags) ) )

TVP_ORDER_UNIQUE is similar to the ORDER token that is used in TDS
responses from the server.

TVP_ORDER_UNIQUE is optional.

ColNum ordinals are 1..N, where 1 is the first column in
TVP_COLMETADATA. That is, ordinals start with 1.

Each TVP_ORDER_UNIQUE token can describe a set of columns for ordering
and/or a set of columns for uniqueness.

The first column ordinal with an ordering bit set is the primary sort
column, the second column ordinal with an ordering bit set is the
secondary sort column, and so on.

The client can send 0 or 1 TVP_ORDER_UNIQUE tokens in a single TVP.

The TVP_ORDER_UNIQUE token MUST always be sent after TVP_COLMETADATA and
before the first TVP_ROW token.

When a TVP is sent to the server, each ColNum ordinal inside a
TVP_ORDER_UNIQUE token MUST refer to a client generated column. Ordinals
that refer to columns with fDefault set are rejected by the server.

**OrderUniqueFlags Possible Combinations And Meaning**

  --------------------------------------------------------------------------
  fOrderAsc   fOrderDesc   fUnique   Meaning
  ----------- ------------ --------- ---------------------------------------
  FALSE       FALSE        FALSE     Invalid flag state, rejected by server

  FALSE       FALSE        TRUE      Column is in unique set

  FALSE       TRUE         FALSE     Column is ordered descending

  FALSE       TRUE         TRUE      Column is ordered descending and in
                                     unique set

  TRUE        FALSE        FALSE     Column is ordered ascending

  TRUE        FALSE        TRUE      Column is ordered ascending and in
                                     unique set

  TRUE        TRUE         FALSE     Invalid flag state, rejected by server

  TRUE        TRUE         TRUE      Invalid flag state, rejected by server
  --------------------------------------------------------------------------

**TVP_COLUMN_ORDERING**

TVP_COLUMN_ORDERING is an optional TVP metadata token that is used to
allow the TDS client to send a different ordering of the columns in a
TVP from the default ordering.

ColNum ordinals are 1..N, where 1 is first column in the TVP (ordinals
start with 1, in other words). These are the same ordinals used with the
TDS ORDER token, for example, to refer to column ordinal as the columns
appear in left to right order.

364. TVP_COLUMN_ORDERING_TOKEN = %x11

     Count = USHORT ; Count of ColNums to follow

     ColNum = USHORT ; A single-column ordinal

     TVP_COLUMN_ORDERING = TVP_COLUMN_ORDERING_TOKEN

     ( Count (\<Count\> ColNum) )

The client can send 0 or 1 TVP_COLUMN_ORDERING tokens in a single TVP.

The TVP_COLUMN_ORDERING token MUST always be sent after TVP_COLMETADATA
and before the first TVP_ROW token.

**Additional details about TVP_COLUMN_ORDERING**

TVP_COLUMN_ORDERING is used to re-order the columns in a TVP. For
example, say, a TVP is defined as the following:

370. TVP_COLUMN_ORDERING = create type myTvpe as table (f1 int / f2
     varchar (max) / f3 datetime)

Then, the TDS client might want to send the f2 field last inside the TVP
as an optimization (streaming the large value last). So the client can
send TVP_COLUMN_ORDERING with order 1,3,2 to indicate that inside the
TVP_ROW section the column f1 is sent first, f3 is sent second, and f2
is sent third.

In this case, the TVP_COLUMN_ORDERING token on the wire for this example
would be:

371. 11 ; TVP_COLUMN_ORDERING_TOKEN

     03 00 ; Count - Number of ColNums to follow.

     01 00 ; ColNum - TVP column ordinal 1 is sent first in
     TVP_COLMETADATA.

     03 00 ; ColNum - TVP column ordinal 3 is sent second in
     TVP_COLMETADATA.

     02 00 ; ColNum - TVP column ordinal 2 is sent third in
     TVP_COLMETADATA.

Duplicate ColNum values are considered an error condition. The ordinal
values of the columns in the actual TVP type are ordered starting with 1
for the first column and adding one for each column from left to right.
The client MUST send one ColNum for each column described in the
TVP_COLMETADATA (so Count MUST match number of columns in
TVP_COLMETADATA).

**TVP_ROW definition**

376. TVP_ROW_TOKEN = %x01 ; A row as defined by TVP_COLMETADATA follows

     TvpColumnData = TYPE_VARBYTE ; Actual value must match metadata for
     the column

     AllColumnData = \*TvpColumnData ; Chunks of data, one per
     non-default column defined

     ; in TVP_COLMETADATA

     TVP_ROW = TVP_ROW_TOKEN

     AllColumnData

     TVP_END_TOKEN = %x00 ; Terminator tag for TVP type, meaning

     ; no more TVP_ROWs to follow and end of

     ; successful transmission of a single TVP

TvpColumnData is repeated once for each non-default column of data
defined in TVP_COLMETADATA.

Each row contains one data \"cell\" per column specified in
TVP_COLMETADATA. On input, columns with the fDefault flag set in
TVP_COLMETADATA are skipped to avoid sending redundant data.

Column data is ordered in same order as the order of items defined in
TVP_COLMETADATA unless a TVP_COLUMN_ORDERING token has been sent to
indicate a change in the ordering of the row values.

###### TDS Type Restrictions

Within a TVP, the following legacy TDS types are not supported:

  -----------------------------------------------------------------------
  TDS type                   Replacement type
  -------------------------- --------------------------------------------
  Binary                     BigBinary

  VarBinary                  BigVarBinary

  Char                       BigChar

  VarChar                    BigVarChar

  Bit                        BitN

  Int1                       IntN

  Int2                       IntN

  Int4                       IntN

  Int8                       IntN

  Float4                     FloatN

  Float8                     FloatN

  Money                      MoneyN

  Decimal                    DecimalN

  Numeric                    NumericN

  DateTime                   DatetimeN

  DateTime4                  DatetimeN

  Money4                     MoneyN
  -----------------------------------------------------------------------

Additional types not allowed in TVP:

-   Null type (NULLTYPE:=\'0x1f\') is not allowed in a TVP.

-   TVP type is not allowed in a TVP (no nesting of TVP in a TVP).

-   TDS types are not to be confused with data types for a database
    server that supports SQL.

##### JSON Values

JSON values are sent as Partially Length-Prefixed Data types (section
[2.2.5.4.4](#Section_7d26a257083e409b81ba897e0c672be0)). The character
encoding of the data follows the character encoding specification
described in
[\[RFC8259\]](https://go.microsoft.com/fwlink/?linkid=867803).

##### Vector Values

The TDS vector payload is a binary token with an 8-byte header followed
by a stream of bytes. The total length of the binary stream is
calculated as:

385. 8-byte header + (NN \* sizeof(T))

Where:

-   NN is the number of dimensions in the vector.

```{=html}
<!-- -->
```
-   sizeof(T) is the number of bytes that each dimension value consumes.

The header is formed as follows:

+-----------+-----------+----------+----------+----------+-------------+
| Layout    | Layout    | Number   | D        | Reserved | Stream of   |
| Format    | Version   | of       | imension |          | Values of   |
|           |           | Di       | Type     |          | Type T      |
|           |           | mensions |          |          |             |
+===========+===========+==========+==========+==========+=============+
| 1 byte    | 1 byte    | 2 bytes  | 1 byte   | 3 bytes  | NN \*       |
|           |           |          |          |          | sizeof(T)   |
|           |           |          |          |          | bytes       |
+-----------+-----------+----------+----------+----------+-------------+
| 0xA9      | 0x01      | NN       | T        | 0x00     |             |
|           |           |          |          |          |             |
|           |           |          |          | 0x00     |             |
|           |           |          |          |          |             |
|           |           |          |          | 0x00     |             |
+-----------+-----------+----------+----------+----------+-------------+

###### Layout Format

The Layout Format MUST be the value 0xA9. It identifies the format of
the byte layout. Future versions of the vector Feature Extension MAY
include different layouts, for example, to support sparse vectors. This
byte MUST always be present, both when sending data to the server and
when receiving data from the server. Both parties (client and server)
SHOULD reject the vector data if the Layout Format is an undocumented
value.

###### Layout Version

The Layout Version MUST be 0x01. Future versions of the byte layout
assign new Version values. There is no explicit relationship between the
byte Layout Version and the Feature Extension version. Either MAY be
assigned new values independently. Clients and servers MUST NOT assume
that a Feature Extension version implies a Layout Version, and vice
versa.

When writing vector data, the writer MUST always choose the lowest
Layout Version that supports the Dimension Type. For example, the
single-precision float Dimension Type is defined for Layout Version
0x01. If a writer supports Layout Version 0x02 and is writing a vector
of single-precision floats, the Layout Version of that vector MUST be
set to 0x01. This ensures backwards compatibility with readers that only
support Layout Version 0x01.

###### Number of Dimensions

The Number of Dimensions specifies how many elements the vector
comprises. For example, with Layout Version 0x01, a vector(6) is a
single-precision float vector with 6 elements. This multi-byte integer
is represented as little-endian, with the least significant byte
appearing at the earlier offset within the header.

####### Implementation Note

The server implementation restricts vectors to a total of 8000 bytes.
Subtracting the 8-byte header leaves 7992 bytes for data. Assuming that
the only data type currently defined is a 32-bit single precision float,
the server supports a maximum vector Number of Dimensions of 1998:

386. (1998 \* 4) + 8 == 8000

###### Dimension Type

The supported Dimension Types are:

  ---------------------------------------------------------------------------------------------------------------------
  Identifier        Value Type         Size              Description
  ----------------- ------------------ ----------------- --------------------------------------------------------------
  0x00              Single-precision   4 bytes           Float values follows the 32-bit
                    float                                [\[IEEE754\]](https://go.microsoft.com/fwlink/?LinkId=89903)
                                                         binary specification when *n* \<= 24. 

  ---------------------------------------------------------------------------------------------------------------------

###### Reserved

All Reserved bytes MUST be set to 0x00. Both server and clients SHOULD
ignore these bytes when reading the header.

###### Stream of Values

The remainder of the binary stream contains the vector values
themselves. Each value consumes the number of bytes implied by the
Dimension Type. All values are represented as little-endian, with the
least significant byte appearing at the earlier offset within that
value's chunk of bytes.

#### Type Info Rule Definition

The TYPE_INFO rule applies to several messages used to describe column
information. For columns of fixed data length, the type is all that is
required to determine the data length. For columns of a variable-length
type, TYPE_VARLEN defines the length of the data contained within the
column, with the following exceptions introduced in TDS 7.3:

DATE MUST NOT have a TYPE_VARLEN. The value is either 3 bytes or 0 bytes
(null).

TIMENTYPE, DATETIME2NTYPE, and DATETIMEOFFSETNTYPE MUST NOT have a
TYPE_VARLEN. The lengths are determined by the SCALE as indicated in
section [2.2.5.4.3](#Section_ce3183a69d8947e8a02fde5a1a1303de).

PRECISION and SCALE MUST occur if the type is NUMERICTYPE, NUMERICNTYPE,
DECIMALTYPE, or DECIMALNTYPE.

SCALE (without PRECISION) MUST occur if the type is TIMENTYPE,
DATETIME2NTYPE, or DATETIMEOFFSETNTYPE (introduced in TDS 7.3).
PRECISION MUST be less than or equal to decimal 38 and SCALE MUST be
less than or equal to the precision value.

SCALE (without PRECISION) MUST occur if the type is VECTORTYPE. See
section [2.2.5.5.7.4](#Section_3404c02191e4445282d69fee2403cb86) for
valid scale values.

COLLATION occurs only if the type is BIGCHARTYPE, BIGVARCHARTYPE,
TEXTTYPE, NTEXTTYPE, NCHARTYPE, or NVARCHARTYPE.

UDT_INFO always occurs if the type is UDTTYPE.

XML_INFO always occurs if the type is XMLTYPE.

USHORTMAXLEN does not occur if PARTLENTYPE is XMLTYPE, UDTTYPE, or
JSONTYPE.

387. USHORTMAXLEN = %xFFFF

     TYPE_INFO = FIXEDLENTYPE

     /

     (VARLENTYPE TYPE_VARLEN \[COLLATION\])

     /

     (VARLENTYPE TYPE_VARLEN \[PRECISION SCALE\])

     /

     (VARLENTYPE SCALE) ; (introduced in TDS 7.3)

     /

     VARLENTYPE ; (introduced in TDS 7.3)

     /

     (PARTLENTYPE

     \[USHORTMAXLEN\]

     \[COLLATION\]

     \[XML_INFO\]

     \[UDT_INFO\])

#### Encryption Key Rule Definition

The EK_INFO rule applies to messages that have encrypted values and
describes the encryption key information. The encryption key information
includes the various encryption key values that are obtained by securing
an encryption key by using different master keys. This rule applies only
if the column encryption feature is negotiated by the client and the
server and is turned ON.

404. Count = BYTE

     EncryptedKey = US_VARBYTE

     KeyStoreName = B_VARCHAR

     KeyPath = US_VARCHAR

     AsymmetricAlgo = B_VARCHAR

     EncryptionKeyValue = EncryptedKey

     KeyStoreName

     KeyPath

     AsymmetricAlgo

     DatabaseId = ULONG

     CekId = ULONG

     CekVersion = ULONG

     CekMDVersion = ULONGLONG

     EK_INFO = DatabaseId

     CekId

     CekVersion

     CekMDVersion

     Count

     \*EncryptionKeyValue

  ------------------------------------------------------------------------------
  Parameter            Description
  -------------------- ---------------------------------------------------------
  Count                The count of EncryptionKeyValue elements that are present
                       in the message.

  EncryptedKey         The ciphertext containing the encryption key that is
                       secured with the master.

  KeyStoreName         The key store name component of the location where the
                       master key is saved.

  KeyPath              The key path component of the location where the master
                       key is saved.

  AsymmetricAlgo       The name of the algorithm that is used for encrypting the
                       encryption key.

  EncryptionKeyValue   The metadata and encrypted value that describe an
                       encryption key. This is enough information to allow
                       retrieval of plaintext encryption keys.

  DatabaseId           A 4-byte integer value that represents the database ID
                       where the column encryption key is stored.

  CekId                An identifier for the column encryption key.

  CekVersion           The key version of the column encryption key.

  CekMDVersion         The metadata version for the column encryption key.
  ------------------------------------------------------------------------------

#### Data Packet Stream Tokens

The tokens defined as follows are used as part of the token-based data
stream. Details about how each token is used inside the data stream are
in section [2.2.6](#Section_c060af9c6db74360954cc28a132c8949).

434. ALTMETADATA_TOKEN = %x88

     ALTROW_TOKEN = %xD3

     COLMETADATA_TOKEN = %x81

     COLINFO_TOKEN = %xA5

     DATACLASSIFICATION_TOKEN = %xA3 ; (introduced in TDS 7.4)

     DONE_TOKEN = %xFD

     DONEPROC_TOKEN = %xFE

     DONEINPROC_TOKEN = %xFF

     ENVCHANGE_TOKEN = %xE3

     ERROR_TOKEN = %xAA

     FEATUREEXTACK_TOKEN = %xAE ; (introduced in TDS 7.4)

     FEDAUTHINFO_TOKEN = %xEE ; (introduced in TDS 7.4)

     INFO_TOKEN = %xAB

     LOGINACK_TOKEN = %xAD

     NBCROW_TOKEN = %xD2 ; (introduced in TDS 7.3)

     OFFSET_TOKEN = %x78

     ORDER_TOKEN = %xA9

     RETURNSTATUS_TOKEN = %x79

     RETURNVALUE_TOKEN = %xAC

     ROW_TOKEN = %xD1

     SESSIONSTATE_TOKEN = %xE4 ; (introduced in TDS 7.4)

     SSPI_TOKEN = %xED

     TABNAME_TOKEN = %xA4

     TVP_ROW_TOKEN = %x01

