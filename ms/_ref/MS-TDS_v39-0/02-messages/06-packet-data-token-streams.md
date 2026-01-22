### Packet Data Token and Tokenless Data Streams

The messages contained in packet data that pass between the client and
the server can be one of two types: a \"token stream\" or a \"tokenless
stream\". A token stream consists of one or more \"tokens\" each
followed by some token-specific data. A \"token\" is a single byte
identifier that is used to describe the data that follows it and
contains information such as token data type, token data length, and so
on. Tokenless streams are used for simple messages. Messages that might
require a more detailed description of the data within it are sent as a
token stream. The following table highlights which messages, as
described previously in sections
[2.2.1](#Section_7ea9ee1ab46141f29004141c0e712935) and
[2.2.2](#Section_342f4cbb2b4b489c8b63f99b12021a94), use token streams
and which do not.

  ------------------------------------------------------------------------
  Message type                    Client or server message Token stream?
  ------------------------------- ------------------------ ---------------
  Pre-Login                       Client                   No

  Login                           Client                   No

  Federated Authentication Token  Client                   No

  SQL Command                     Client                   No

  Bulk Load                       Client                   Yes

  Remote Procedure Call (RPC)     Client                   Yes

  Attention                       Client                   No

  Transaction Manager Request     Client                   No

  Pre-Login Response              Server                   No

  Federated Authentication        Server                   Yes
  Information                                              

  FeatureExtAck                   Server                   Yes

  Login Response                  Server                   Yes

  Row Data                        Server                   Yes

  Return Status                   Server                   Yes

  Return Parameters               Server                   Yes

  Response Completion             Server                   Yes

  Session State                   Server                   Yes

  Error and Info                  Server                   Yes

  Attention Acknowledgement       Server                   No
  ------------------------------------------------------------------------

#### Tokenless Stream

As shown in section [2.2.4](#Section_dc3a08548230482fbbb9d94a3b905a26),
some messages do not use tokens to describe the data portion of the
[**data stream**](#gt_151643ce-fb5e-460e-8bdf-dc10bbd1950e). In these
cases, all the information required to describe the packet data is
contained in the packet header. This is referred to as a tokenless
stream and is essentially just a collection of packets and data.

#### Token Stream

More complex messages (for example, column metadata, row data, and data
type data) are constructed by using tokens. As described in section
[2.2.4](#Section_dc3a08548230482fbbb9d94a3b905a26), a token stream
consists of a single byte identifier, followed by token-specific data.
The definitions of the different token streams can be found in section
[2.2.7](#Section_67b6113cd72242d1902c3f6e8de09173).

##### Token Definition

There are four classes of token definitions:

-   [Zero Length
    Token(xx01xxxx)](#Section_9b571ae55d0048748cf798f30ca69dbd)

-   [Fixed Length
    Token(xx11xxxx)](#Section_5d73136323384ebeb1eb5cce0780ef8f)

-   [Variable Length
    Tokens(xx10xxxx)](#Section_d3edea23be1f416098f50de233ffeebc)

-   [Variable Count
    Tokens(xx00xxxx)](#Section_5ff30ab3aaa7474ba095662fcaed5549)

The following sections specify the bit pattern of each token class,
various extensions to this bit pattern for a given token class, and a
description of its function(s).

###### Zero Length Token(xx01xxxx)

This class of token is not followed by a length specification. There is
no data associated with the token. A zero length token always has the
following bit sequence:

  ---------------------------------------------------------------------------
  0          1          2    3    4          5          6          7
  ---------- ---------- ---- ---- ---------- ---------- ---------- ----------
  0 or 1     0 or 1     0    1    0 or 1     0 or 1     0 or 1     0 or 1

  ---------------------------------------------------------------------------

A value of "0 or 1" denotes a bit position that can contain the bit
value "0" or "1".

###### Fixed Length Token(xx11xxxx)

This class of token is followed by 1, 2, 4, or 8 bytes of data. No
length specification follows this token because the length of its
associated data is encoded in the token itself. The different fixed
data-length token definitions take the form of one of the following bit
sequences, depending on whether the token is followed by 1, 2, 4, or 8
bytes of data. Also in the table, a value of "0 or 1" denotes a bit
position that can contain the bit value "0" or "1".

  --------------------------------------------------------------------------
  0      1      2   3   4   5   6      7      Description
  ------ ------ --- --- --- --- ------ ------ ------------------------------
  0 or 1 0 or 1 1   1   0   0   0 or 1 0 or 1 Token is followed by 1 byte of
                                              data.

  0 or 1 0 or 1 1   1   0   1   0 or 1 0 or 1 Token is followed by 2 bytes
                                              of data.

  0 or 1 0 or 1 1   1   1   0   0 or 1 0 or 1 Token is followed by 4 bytes
                                              of data.

  0 or 1 0 or 1 1   1   1   1   0 or 1 0 or 1 Token is followed by 8 bytes
                                              of data.
  --------------------------------------------------------------------------

Fixed-length tokens are used by the following data types: *bigint, int*,
*smallint*, *tinyint, float, real, money, smallmoney, datetime,
smalldatetime,* and *bit*. The type definition is always represented in
COLMETADATA (section
[2.2.7.4](#Section_58880b9f381c43b2bf8b0727a98c4f4c)) and ALTMETADATA
(section [2.2.7.1](#Section_004bba4a8c234d7bab2cd9e7ba864cd0)) data
streams as a single byte **Type**. Additional details are specified in
section [2.2.5.4.2](#Section_859eb3d280d340f6a637414552c9c552).

###### Variable Length Tokens(xx10xxxx)

Except as noted later in this section, this class of token definition is
followed by a length specification. The length, in bytes, of this length
is included in the token itself as a Length value (see section
[2.2.7.3](#Section_aa8466c5ca3d48caa6387c1becebe754)).

The following are the two data types that are of variable length.

-   Real variable length data types like char and binary and nullable
    data types, which are either their normal fixed length corresponding
    to their TYPE_INFO (section
    [2.2.5.6](#Section_cbe9c510eae64b1f9893a098944d430a)), or a special
    length if null.

> Char and binary data types have values that are either null or 0 to
> 65534 (0x0000 to 0xFFFE) bytes in length. Null is represented by a
> length of 65535 (0xFFFF). A char or binary, which cannot be null, can
> still have a length of zero (for example an empty value). A program
> that MUST pad a value to a fixed length adds blanks to the end of a
> char and binary zeros to the end of a binary.

-   Text and image data types have values that are either null, or 0 to
    2 gigabytes (0x00000000 to 0x7FFFFFFF bytes) in length. Null is
    represented by a length of -1 (0xFFFFFFFF). No other length
    specification is supported.

Other nullable data types have a length of 0 if they are null.

**Note** The DATACLASSIFICATION variable length token does not start
with a length specification (see section
[2.2.7.5](#Section_813b88bc0a324e7ebc92d98f62cb8981)).

###### Variable Count Tokens(xx00xxxx)

This class of token definition is followed by a count of the number of
fields that follow the token. Each field length is dependent on the
token type. The total length of the token can be determined only by
walking the fields. As shown in the following table, a variable count
token always has its third and fourth bits set to "0", and a value of "0
or 1" in the remaining bit positions denotes a bit position that can
contain the bit value "0" or "1".

  -----------------------------------------------------------------------
  0        1        2      3       4        5         6         7
  -------- -------- ------ ------- -------- --------- --------- ---------
  0 or 1   0 or 1   0      0       0 or 1   0 or 1    0 or 1    0 or 1

  -----------------------------------------------------------------------

There are two variable count tokens. COLMETADATA (section
[2.2.7.4](#Section_58880b9f381c43b2bf8b0727a98c4f4c)) and ALTMETADATA
(section [2.2.7.1](#Section_004bba4a8c234d7bab2cd9e7ba864cd0)) both use
a 2-byte count.

#### Done and Attention Tokens

The DONE token (section
[2.2.7.6](#Section_3c06f11098bd4d5bb836b1ba66452cb7)) marks the end of
the response for each executed [**SQL
statement**](#gt_dc5ca224-43ec-4b44-9dba-726d6fd6057d). Based on the SQL
statement and the context in which it is executed, the server MAY
generate a DONEPROC (section
[2.2.7.8](#Section_65e24140edea46e5b710209af2016195)) or DONEINPROC
(section [2.2.7.7](#Section_43e891c5f7a1432f8f9f233c4cd96afb)) token
instead.

The attention signal is sent by using the
[**out-of-band**](#gt_26c1caf3-c889-4b99-a22b-9da056d397cf) write
provided by the network library. An out-of-band write is the ability to
send the attention signal no matter if the sender is in the middle of
sending or processing a message or simply sitting idle. If that function
is not supported, the client MUST simply read and discard all of the
data, except SESSIONSTATE data (section
[2.2.7.21](#Section_626fbe19f3564599ba17c70f44005106)), from the server
until the final DONE token, which acknowledges that the attention signal
is read.[\<8\>](\l)

