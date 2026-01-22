### Packet Data Token Stream Definition

This section describes the various tokens supported in a token-based
packet data stream, as described in section
[2.2.4.2](#Section_7cfc401bf9b6404da44722e21e593742). The corresponding
message types that use token-based packet data streams are identified in
the table in section [2.2.4](#Section_dc3a08548230482fbbb9d94a3b905a26).

#### ALTMETADATA

**Token Stream Name:**

874. ALTMETADATA

**Token Stream Function:**

Describes the data type, length, and name of column data that result
from a [**SQL statement**](#gt_dc5ca224-43ec-4b44-9dba-726d6fd6057d)
that generates totals.

**Token Stream Comments:**

The token value is 0x88.

This token is used to tell the client the data type and length of the
column data. It describes the format of the data found in an ALTROW
[**data stream**](#gt_151643ce-fb5e-460e-8bdf-dc10bbd1950e). ALTMETADATA
and the corresponding ALTROW MUST be in the same [**result
set**](#gt_c8a27238-8ccc-442b-9604-75f74d3e6b3d).

All ALTMETADATA data streams are grouped.

A preceding COLMETADATA (section
[2.2.7.4](#Section_58880b9f381c43b2bf8b0727a98c4f4c)) MUST exist before
an ALTMETADATA token. There might be COLINFO and TABNAME streams between
COLMETADATA and ALTMETADATA.

**Note**  ALTMETADATA was deprecated in TDS 7.4.

**Token Stream-Specific Rules:**

875. TokenType = BYTE

     Count = USHORT

     Id = USHORT

     ByCols = UCHAR

     Op = BYTE

     Operand = USHORT

     UserType = USHORT/ULONG; (changed to ULONG in TDS 7.2)

     fNullable = BIT

     fCaseSen = BIT

     usUpdateable = 2BIT ; 0 = ReadOnly

     ; 1 = Read/Write

     ; 2 = Unused

     fIdentity = BIT

     fComputed = BIT ; (introduced in TDS 7.2)

     usReservedODBC = 2BIT

     fFixedLenCLRType = BIT ; (introduced in TDS 7.2)

     usReserved = 7BIT

     Flags = fNullable

     fCaseSen

     usUpdateable

     fIdentity

     (FRESERVEDBIT / fComputed)

     usReservedODBC

     (FRESERVEDBIT / fFixedLenCLRType)

     usReserved

     NumParts = BYTE ; (introduced in TDS 7.2)

     PartName = US_VARCHAR ; (introduced in TDS 7.2)

     TableName = US_VARCHAR ; (removed in TDS 7.2)

     /

     (NumParts

     1\*PartName) ; (introduced in TDS 7.2)

     ColName = B_VARCHAR

     ColNum = USHORT

     ComputeData = Op

     Operand

     UserType

     Flags

     TYPE_INFO

     \[TableName\]

     ColName

The **TableName** field is specified only if a text, ntext, or image
column is included in the result set.

**Token Stream Definition:**

922. ALTMETADATA = TokenType

     Count

     Id

     ByCols

     \*(\<ByCols\> ColNum)

     1\*ComputeData

**Token Stream Parameter Details:**

+-------+--------------------------------------------------------------+
| Para  | Description                                                  |
| meter |                                                              |
+=======+==============================================================+
| Toke  | ALTMETADATA_TOKEN[\<41\>](\l)                                |
| nType |                                                              |
+-------+--------------------------------------------------------------+
| Count | The count of columns (number of aggregate operators) in the  |
|       | token stream.                                                |
+-------+--------------------------------------------------------------+
| Id    | The Id of the SQL statement to which the total column        |
|       | formats apply. Each ALTMETADATA token MUST have its own      |
|       | unique Id in the same result set. This Id lets the client    |
|       | correctly interpret later ALTROW data streams.               |
+-------+--------------------------------------------------------------+
| B     | The number of grouping columns in the SQL statement that     |
| yCols | generates totals. For example, the SQL clause *compute       |
|       | count(sales) by year, month, division, department* has four  |
|       | grouping columns.                                            |
+-------+--------------------------------------------------------------+
| Op    | The type of aggregate operator.                              |
|       |                                                              |
|       | 928. AOPSTDEV = %x30 ; Standard deviation (STDEV)            |
|       |                                                              |
|       |      AOPSTDEVP = %x31 ; Standard deviation of the population |
|       |      (STDEVP)                                                |
|       |                                                              |
|       |      AOPVAR = %x32 ; Variance (VAR)                          |
|       |                                                              |
|       |      AOPVARP = %x33 ; Variance of population (VARP)          |
|       |                                                              |
|       |      AOPCNT = %x4B ; Count of rows (COUNT)                   |
|       |                                                              |
|       |      AOPSUM = %x4D ; Sum of the values in the rows (SUM)     |
|       |                                                              |
|       |      AOPAVG = %x4F ; Average of the values in the rows (AVG) |
|       |                                                              |
|       |      AOPMIN = %x51 ; Minimum value of the rows (MIN)         |
|       |                                                              |
|       |      AOPMAX = %x52 ; Maximum value of the rows (MAX)         |
+-------+--------------------------------------------------------------+
| Op    | The column number, starting from 1, in the result set that   |
| erand | is the operand to the aggregate operator.                    |
+-------+--------------------------------------------------------------+
| Use   | The user type ID of the data type of the column. Depending   |
| rType | on the TDS version that is used, valid values are 0x0000 or  |
|       | 0x00000000, with the exceptions of data type timestamp       |
|       | (0x0050 or 0x00000050) and alias types (greater than 0x00FF  |
|       | or 0x000000FF).                                              |
+-------+--------------------------------------------------------------+
| Flags | These bit flags are described in [least significant bit      |
|       | order](#Section_bbc22f15e1a04338a169c79819d39b1c). With the  |
|       | exception of **fNullable**, all of these bit flags SHOULD be |
|       | set to zero. For a description of each bit flag, see section |
|       | 2.2.7.4:                                                     |
|       |                                                              |
|       | -   fNullable is a bit flag, 1 if the column is nullable.    |
|       |                                                              |
|       | -   fCaseSen                                                 |
|       |                                                              |
|       | -   usUpdateable                                             |
|       |                                                              |
|       | -   fIdentity                                                |
|       |                                                              |
|       | -   fComputed                                                |
|       |                                                              |
|       | -   usReservedODBC                                           |
|       |                                                              |
|       | -   fFixedLenCLRType                                         |
+-------+--------------------------------------------------------------+
| Tabl  | See section 2.2.7.4 for a description of TableName. This     |
| eName | field SHOULD never be sent because SQL statements that       |
|       | generate totals exclude NTEXT/TEXT/IMAGE.                    |
+-------+--------------------------------------------------------------+
| Co    | The column name. Contains the column name length and column  |
| lName | name.                                                        |
+-------+--------------------------------------------------------------+
| C     | USHORT specifying the column number as it appears in the     |
| olNum | COMPUTE clause. ColNum appears ByCols times.                 |
+-------+--------------------------------------------------------------+

#### ALTROW

**Token Stream Name:**

937. ALTROW

**Token Stream Function:**

Used to send a complete row of total data, where the data format is
provided by the ALTMETADATA token.

**Token Stream Comments:**

-   The token value is 0xD3.

-   The ALTROW token is similar to the ROW_TOKEN, but also contains an
    Id field. This Id matches an Id given in ALTMETADATA (one Id for
    each [**SQL statement**](#gt_dc5ca224-43ec-4b44-9dba-726d6fd6057d)).
    This provides the mechanism for matching row data with correct SQL
    statements. ALTROW and the corresponding ALTMETADATA MUST be in the
    same [**result set**](#gt_c8a27238-8ccc-442b-9604-75f74d3e6b3d).

-   **Note**  ALTROW was deprecated in TDS 7.4.

**Token Stream-Specific Rules:**

938. TokenType = BYTE

     Id = USHORT

     Data = TYPE_VARBYTE

     ComputeData = Data

**Token Stream Definition:**

944. ALTMETADATA = TokenType

     Id

     1\*ComputeData

The **ComputeData** element is repeated Count times, where Count is
specified in ALTMETADATA_TOKEN.

**Token Stream Parameter Details:**

  --------------------------------------------------------------------------
  Parameter   Description
  ----------- --------------------------------------------------------------
  TokenType   ALTROW_TOKEN[\<42\>](\l)

  Id          The Id of the SQL statement that generates totals to which the
              total column formats apply. This Id lets the client correctly
              interpret later ALTROW [**data
              streams**](#gt_151643ce-fb5e-460e-8bdf-dc10bbd1950e).

  Data        The actual data for the column. The TYPE_INFO information
              describing the data type of this data is given in the
              preceding COLMETADATA_TOKEN, ALTMETADATA_TOKEN, or
              OFFSET_TOKEN.
  --------------------------------------------------------------------------

#### COLINFO

**Token Stream Name:**

947. COLINFO

**Token Stream Function:**

Describes the column information in browse mode
[\[MSDN-BROWSE\]](https://go.microsoft.com/fwlink/?LinkId=140931),
sp_cursoropen, and sp_cursorfetch.

**Token Stream Comments**

-   The token value is 0xA5.

-   The TABNAME token contains the actual table name associated with
    COLINFO.

**Token Stream Specific Rules:**

948. TokenType = BYTE

     Length = USHORT

     ColNum = BYTE

     TableNum = BYTE

     Status = BYTE

     ColName = B_VARCHAR

     ColProperty = ColNum

     TableNum

     Status

     \[ColName\]

The **ColProperty** element is repeated for each column in the [**result
set**](#gt_c8a27238-8ccc-442b-9604-75f74d3e6b3d).

**Token Stream Definition:**

960. COLINFO = TokenType

     Length

     1\*ColProperty

**Token Stream Parameter Details:**

+--------+-------------------------------------------------------------+
| Par    | Description                                                 |
| ameter |                                                             |
+========+=============================================================+
| Tok    | COLINFO_TOKEN                                               |
| enType |                                                             |
+--------+-------------------------------------------------------------+
| Length | The actual data length, in bytes, of the ColProperty        |
|        | stream. The length does not include token type and length   |
|        | field.                                                      |
+--------+-------------------------------------------------------------+
| ColNum | The column number in the result set.                        |
+--------+-------------------------------------------------------------+
| Ta     | The number of the base table that the column was derived    |
| bleNum | from. The value is 0 if the value of Status is EXPRESSION.  |
+--------+-------------------------------------------------------------+
| Status | 0x4: EXPRESSION (the column was the result of an            |
|        | expression).                                                |
|        |                                                             |
|        | 0x8: KEY (the column is part of a key for the associated    |
|        | table).                                                     |
|        |                                                             |
|        | 0x10: HIDDEN (the column was not requested, but was added   |
|        | because it was part of a key for the associated table).     |
|        |                                                             |
|        | 0x20: DIFFERENT_NAME (the column name is different than the |
|        | requested column name in the case of a column alias).       |
+--------+-------------------------------------------------------------+
| C      | The base column name. This only occurs if DIFFERENT_NAME is |
| olName | set in Status.                                              |
+--------+-------------------------------------------------------------+

#### COLMETADATA

**Token Stream Name:**

963. COLMETADATA

**Token Stream Function:**

Describes the [**result set**](#gt_c8a27238-8ccc-442b-9604-75f74d3e6b3d)
for interpretation of following ROW [**data
streams**](#gt_151643ce-fb5e-460e-8bdf-dc10bbd1950e).

**Token Stream Comments:**

-   The token value is 0x81.

-   This token is used to tell the client the data type and length of
    the column data. It describes the format of the data found in a ROW
    data stream.

-   All COLMETADATA data streams are grouped together.

**Token Stream-Specific Rules:**

964. TokenType = BYTE

     Count = USHORT

     UserType = USHORT/ULONG; (Changed to ULONG in TDS 7.2)

     fNullable = BIT

     fCaseSen = BIT

     usUpdateable = 2BIT ; 0 = ReadOnly

     ; 1 = Read/Write

     ; 2 = Unused

     fIdentity = BIT

     fComputed = BIT ; (introduced in TDS 7.2)

     usReservedODBC = 2BIT ; (only exists in TDS 7.3.A and below)

     fSparseColumnSet = BIT ; (introduced in TDS 7.3.B)

     fEncrypted = BIT ; (introduced in TDS 7.4)

     usReserved3 = BIT ; (introduced in TDS 7.4)

     fFixedLenCLRType = BIT ; (introduced in TDS 7.2)

     usReserved = 4BIT

     fHidden = BIT ; (introduced in TDS 7.2)

     fKey = BIT ; (introduced in TDS 7.2)

     fNullableUnknown = BIT ; (introduced in TDS 7.2)

     Flags = fNullable

     fCaseSen

     usUpdateable

     fIdentity

     (FRESERVEDBIT / fComputed)

     usReservedODBC

     (FRESERVEDBIT / fFixedLenCLRType)

     (usReserved / (FRESERVEDBIT fSparseColumnSet fEncrypted
     usReserved3))

     ; (introduced in TDS 7.4)

     (FRESERVEDBIT / fHidden)

     (FRESERVEDBIT / fKey)

     (FRESERVEDBIT / fNullableUnknown)

     NumParts = BYTE ; (introduced in TDS 7.2)

     PartName = US_VARCHAR ; (introduced in TDS 7.2)

     TableName = NumParts

     1\*PartName

     ColName = B_VARCHAR

     BaseTypeInfo = TYPE_INFO ; (BaseTypeInfo introduced in TDS 7.4)

     EncryptionAlgo = BYTE ; (EncryptionAlgo introduced in TDS 7.4)

     AlgoName = B_VARCHAR ; (introduced in TDS 7.4)

     EncryptionAlgoType = BYTE ; (introduced in TDS 7.4)

     NormVersion = BYTE ; (introduced in TDS 7.4)

     Ordinal = USHORT ; (introduced in TDS 7.4)

     CryptoMetaData = Ordinal ; (CryptoMetaData introduced in TDS 7.4)

     UserType

     BaseTypeInfo

     EncryptionAlgo

     \[AlgoName\]

     EncryptionAlgoType

     NormVersion

     EkValueCount = USHORT ; (introduced in TDS 7.4)

     CekTable = EkValueCount ; (introduced in TDS 7.4)

     \*EK_INFO ; (introduced in TDS 7.4)

     ColumnData = UserType

     Flags

     TYPE_INFO

     \[TableName\]

     \[CryptoMetaData\]

     ColName

     NoMetaData = %xFF %xFF

The **TableName** element is specified only if a text, ntext, or image
column is included in the result set.

**Token Stream Definition:**

1040. COLMETADATA = TokenType

      Count

      \[CekTable\]

      NoMetaData / (1\*ColumnData)

**Token Stream Parameter Details:**

+------------+---------------------------------------------------------+
| Parameter  | Description                                             |
+============+=========================================================+
| TokenType  | COLMETADATA_TOKEN                                       |
+------------+---------------------------------------------------------+
| Count      | The count of columns (number of aggregate operators) in |
|            | the token stream. In the event that the client          |
|            | requested no metadata to be returned (see section       |
|            | [2.2.6.6](#Section_619c43b694954a589e49a4950db245b3)    |
|            | for information about the OptionFlags parameter in the  |
|            | RPCRequest token), the value of Count is 0xFFFF. This   |
|            | has the same effect on Count as a zero value (for       |
|            | example, no ColumnData is sent).                        |
+------------+---------------------------------------------------------+
| UserType   | The user type ID of the data type of the column.        |
|            | Depending on the TDS version that is used, valid values |
|            | are 0x0000 or 0x00000000, with the exceptions of data   |
|            | type timestamp (0x0050 or 0x00000050) and alias types   |
|            | (greater than 0x00FF or 0x000000FF).                    |
+------------+---------------------------------------------------------+
| Flags      | The size of the Flags parameter is always fixed at 16   |
|            | bits regardless of the TDS version. Each of the 16 bits |
|            | of the Flags parameter is interpreted based on the TDS  |
|            | version negotiated during login. Bit flags, in [least   |
|            | significant bit                                         |
|            | order](#Section_bbc22f15e1a04338a169c79819d39b1c):      |
|            |                                                         |
|            | -   fNullable is a bit flag. Its value is 1 if the      |
|            |     column is nullable.                                 |
|            |                                                         |
|            | -   fCaseSen is a bit flag. Set to 1 for string columns |
|            |     with binary collation and always for the XML data   |
|            |     type. Set to 0 otherwise.                           |
|            |                                                         |
|            | -   usUpdateable is a 2-bit field. Its value is 0 if    |
|            |     column is read-only, 1 if column is read/write and  |
|            |     2 if updateable is unknown.                         |
|            |                                                         |
|            | -   fIdentity is a bit flag. Its value is 1 if the      |
|            |     column is an identity column.                       |
|            |                                                         |
|            | -   fComputed is a bit flag. Its value is 1 if the      |
|            |     column is a COMPUTED column.                        |
|            |                                                         |
|            | -   usReservedODBC is a 2-bit field that is used by ODS |
|            |     gateways supporting the ODBC ODS gateway driver.    |
|            |                                                         |
|            | -   fFixedLenCLRType is a bit flag. Its value is 1 if   |
|            |     the column is a fixed-length [**common language     |
|            |     runtime user-defined type (CLR                      |
|            |     UDT)**](#gt_5f1d976e-cd4b-4a78-a6a1-d0bdb0aa0360).  |
|            |                                                         |
|            | -   fSparseColumnSet, introduced in TDS version 7.3.B,  |
|            |     is a bit flag. Its value is 1 if the column is the  |
|            |     special XML column for the sparse column set. For   |
|            |     information about using column sets, see            |
|            |     [\[MSDN-ColS                                        |
|            | ets\]](https://go.microsoft.com/fwlink/?LinkId=128616). |
|            |                                                         |
|            | -   fEncrypted is a bit flag. Its value is 1 if the     |
|            |     column is encrypted transparently and has to be     |
|            |     decrypted to view the plaintext value. This flag is |
|            |     valid when the column encryption feature is         |
|            |     negotiated between client and server and is turned  |
|            |     on.                                                 |
|            |                                                         |
|            | -   fHidden is a bit flag. Its value is 1 if the column |
|            |     is part of a hidden primary key created to support  |
|            |     a T-SQL SELECT statement containing FOR             |
|            |     BROWSE.[\<43\>](\l)                                 |
|            |                                                         |
|            | -   fKey is a bit flag. Its value is 1 if the column is |
|            |     part of a primary key for the row and the T-SQL     |
|            |     SELECT statement contains FOR BROWSE.               |
|            |                                                         |
|            | -   fNullableUnknown is a bit flag. Its value is 1 if   |
|            |     it is unknown whether the column might be nullable. |
+------------+---------------------------------------------------------+
| TableName  | The fully qualified base table name for this column. It |
|            | contains the table name length and table name. This     |
|            | exists only for text, ntext, and image columns. It      |
|            | specifies the number of parts that are returned and     |
|            | then repeats PartName once for each NumParts.           |
+------------+---------------------------------------------------------+
| ColName    | The column name. It contains the column name length and |
|            | column name.                                            |
+------------+---------------------------------------------------------+
| Ba         | The TYPEINFO for the plaintext data.                    |
| seTypeInfo |                                                         |
+------------+---------------------------------------------------------+
| Ek         | The size of CekTable. It represents the number of       |
| ValueCount | entries in CekTable.                                    |
+------------+---------------------------------------------------------+
| CekTable   | A table of various encryption keys that are used to     |
|            | secure the plaintext data. It contains one row for each |
|            | encryption key. Each row can have multiple encryption   |
|            | key values, and each value represents the cipher text   |
|            | of the same encryption key that is secured by using a   |
|            | different master key. The size of this table is         |
|            | determined by EkValueCount. This table MUST be sent     |
|            | when COLUMNENCRYPTION is negotiated by client and       |
|            | server and is turned on.                                |
+------------+---------------------------------------------------------+
| Encr       | This byte describes the encryption algorithm that is    |
| yptionAlgo | used.                                                   |
|            |                                                         |
|            | For a custom encryption algorithm, the EncryptionAlgo   |
|            | value MUST be set to 0 and the actual encryption        |
|            | algorithm MUST be inferred from the AlgoName. For all   |
|            | other values, AlgoName MUST NOT be sent.                |
|            |                                                         |
|            | If EncryptionAlgo is set to 1, the algorithm that is    |
|            | used is AEAD_AES_256_CBC_HMAC_SHA512, as described in   |
|            | [\[IETF-Auth                                            |
|            | Encr\]](https://go.microsoft.com/fwlink/?LinkId=524322) |
|            | section 5.4.                                            |
|            |                                                         |
|            | If EncryptionAlgo is set to 2, the algorithm that is    |
|            | used is AEAD_AES_256_CBC_HMAC_SHA256.                   |
|            |                                                         |
|            | Other values are reserved for future use.               |
+------------+---------------------------------------------------------+
| AlgoName   | Reserved for future use.                                |
|            |                                                         |
|            | Algorithm name literal that is used for encrypting the  |
|            | plaintext value. This is an optional field and MUST be  |
|            | sent when EncryptionAlgo = 0. For all other values of   |
|            | EncryptionAlgo, this field MUST NOT be sent.            |
+------------+---------------------------------------------------------+
| Encrypti   | A field that describes the encryption algorithm type.   |
| onAlgoType | Available values are defined as follows:                |
|            |                                                         |
|            | 1 = Deterministic encryption.                           |
|            |                                                         |
|            | 2 = Randomized encryption.                              |
+------------+---------------------------------------------------------+
| N          | The normalization version to which plaintext data MUST  |
| ormVersion | be normalized. Version numbering starts at 0x01.        |
+------------+---------------------------------------------------------+
| Ordinal    | Where the encryption key information is located in      |
|            | CekTable. Ordinal starts at 0.                          |
+------------+---------------------------------------------------------+
| Cryp       | This describes the encryption metadata for a column. It |
| toMetaData | contains the ordinal, the UserType, the TYPE_INFO       |
|            | (BaseTypeInfo) for the plaintext value, the encryption  |
|            | algorithm that is used, the algorithm name literal, the |
|            | encryption algorithm type, and the normalization        |
|            | version.                                                |
+------------+---------------------------------------------------------+
| NoMetaData | This notifies client that no metadata follows the       |
|            | COLMETADATA token. When fNoMetaData is set to 1, client |
|            | notifies server that it has already cached the metadata |
|            | from a previous RPC Request (section 2.2.6.6), and      |
|            | server sends no metadata.[\<44\>](\l)                   |
+------------+---------------------------------------------------------+

#### DATACLASSIFICATION

**Token Stream Name:**

1044. DATACLASSIFICATION

**Token Stream Function:**

Introduced in TDS 7.4, the DATACLASSIFICATION token SHOULD[\<45\>](\l)
describe the [**data
classification**](#gt_acdeafb0-9b24-420e-b712-9284ad49eb56) of the query
[**result set**](#gt_c8a27238-8ccc-442b-9604-75f74d3e6b3d).

**Token Stream Comments:**

-   The token value is 0xA3.

-   This token is sent by the server only if the client sends a
    DATACLASSIFICATION FeatureExt in the Login message and the server
    responds with a DATACLASSIFICATION FeatureExtAck. Additionally, for
    this token to be sent, the query result set MUST contain output
    columns whose results are based on sources that are classified.

**Token Stream-Specific Rules:**

1045. TokenType = BYTE

      SensitivityLabelCount = USHORT

      SensitivityLabelName = B_VARCHAR

      SensitivityLabelId = B_VARCHAR

      InformationTypeCount = USHORT

      InformationTypeName = B_VARCHAR

      InformationTypeId = B_VARCHAR

      SensitivityLabelIndex = USHORT

      InformationTypeIndex = USHORT

      NumSensitivityProperties = USHORT

      NumResultSetColumns = USHORT

      SensitivityRank = LONG

      SensitivityLabel = SensitivityLabelName

      SensitivityLabelId

      SensitivityLabels = SensitivityLabelCount

      \[SensitivityLabelCount\] \*SensitivityLabel

      InformationType = InformationTypeName

      InformationTypeId

      InformationTypes = InformationTypeCount

      \[InformationTypeCount\] \*InformationType

      SensitivityProperty = SensitivityLabelIndex

      InformationTypeIndex

      \[SensitivityRank\]

      ColumnSensitivityMetadata = NumSensitivityProperties

      \[NumSensitivityProperties\] \*SensitivityProperty

      DataClassificationPerColumnData = NumResultSetColumns

      \[NumResultSetColumns\] \*ColumnSensitivityMetadata

**Token Stream Definition:**

1080. DATACLASSIFICATION = TokenType

      SensitivityLabels

      InformationTypes

      \[SensitivityRank\]

      DataClassificationPerColumnData

**Token Stream Parameter Details:**

+----------------+-----------------------------------------------------+
| Parameter      | Description                                         |
+================+=====================================================+
| TokenType      | DATACLASSIFICATION_TOKEN                            |
+----------------+-----------------------------------------------------+
| Sensiti        | The count of sensitivity labels for this result     |
| vityLabelCount | set. The value can be 0 or greater.                 |
+----------------+-----------------------------------------------------+
| Sensit         | The name for a sensitivity label. It contains the   |
| ivityLabelName | sensitivity label name length and sensitivity label |
|                | name. It is intended to be human readable.          |
+----------------+-----------------------------------------------------+
| Sens           | The identifier for a sensitivity label. It contains |
| itivityLabelId | the sensitivity label identifier length and         |
|                | sensitivity label identifier. It is intended for    |
|                | linking the sensitivity label to an information     |
|                | protection system.                                  |
+----------------+-----------------------------------------------------+
| Inform         | The count of information types for this result set. |
| ationTypeCount | The value can be 0 or greater.                      |
+----------------+-----------------------------------------------------+
| Infor          | The name for an information type. It contains the   |
| mationTypeName | information type name length and information type   |
|                | name. It is intended to be human readable.          |
+----------------+-----------------------------------------------------+
| Inf            | The identifier for an information type. It contains |
| ormationTypeId | the information type identifier length and          |
|                | information type identifier. It is intended for     |
|                | linking the information type to an information      |
|                | protection system.                                  |
+----------------+-----------------------------------------------------+
| Sensiti        | The index into the SensitivityLabels array that     |
| vityLabelIndex | indicates which SensitivityLabel is associated with |
|                | SensitivityProperty. A value of USHORT_MAX (0xFFFF) |
|                | indicates that there is no sensitivity label for    |
|                | SensitivityProperty.                                |
+----------------+-----------------------------------------------------+
| Inform         | The index into the InformationTypes array that      |
| ationTypeIndex | indicates which InformationType is associated with  |
|                | SensitivityProperty. A value of USHORT_MAX (0xFFFF) |
|                | indicates that there is no information type for     |
|                | SensitivityProperty.                                |
+----------------+-----------------------------------------------------+
| NumRe          | Depending on its configuration, the server can send |
| sultSetColumns | additional information about the data               |
|                | classification for each column. The values of this  |
|                | field are as follows:                               |
|                |                                                     |
|                | -   0 = Additional information is not sent.         |
|                |                                                     |
|                | -   The number of columns in the result set. This   |
|                |     number MUST be the same number provided by the  |
|                |     Count parameter in the COLMETADATA token        |
|                |     (section                                        |
|                |     [2.                                             |
|                | 2.7.4](#Section_58880b9f381c43b2bf8b0727a98c4f4c)). |
+----------------+-----------------------------------------------------+
| NumSensiti     | The number of sensitivity properties that are       |
| vityProperties | associated with a column. The value can be 0 or     |
|                | greater.                                            |
+----------------+-----------------------------------------------------+
| S              | A relative ranking of the sensitivity of a query or |
| ensitivityRank | of a column that is part of per-column data.        |
|                | Available values are defined as follows:            |
|                |                                                     |
|                | -   -1 = Not defined                                |
|                |                                                     |
|                | -   0 = None                                        |
|                |                                                     |
|                | -   10 = Low                                        |
|                |                                                     |
|                | -   20 = Medium                                     |
|                |                                                     |
|                | -   30 = High                                       |
|                |                                                     |
|                | -   40 = Critical                                   |
|                |                                                     |
|                | A sensitivity ranking is sent by the server only if |
|                | both of the following are true:                     |
|                |                                                     |
|                | -   The client sends a DATACLASSIFICATION feature   |
|                |     extension in a Login message in which           |
|                |     DATACLASSIFICATION_VERSION is set to 2.         |
|                |                                                     |
|                | -   The server responds with a DATACLASSIFICATION   |
|                |     feature extension acknowledgement in which      |
|                |     DATACLASSIFICATION_VERSION is set to 2.         |
+----------------+-----------------------------------------------------+

#### DONE

**Token Stream Name:**

1085. DONE

**Token Stream Function:**

Indicates the completion status of a [**SQL
statement**](#gt_dc5ca224-43ec-4b44-9dba-726d6fd6057d).

**Token Stream Comments**

-   The token value is 0xFD.

-   This token is used to indicate the completion of a SQL statement. As
    multiple SQL statements can be sent to the server in a single SQL
    batch, multiple DONE tokens can be generated. In this case, all but
    the final DONE token has a Status value with DONE_MORE bit set
    (details follow).

-   A DONE token is returned for each SQL statement in the SQL batch
    except variable declarations.

-   For execution of SQL statements within [**stored
    procedures**](#gt_324d32b3-f4f3-41c9-b695-78c498094fb7), DONEPROC
    and DONEINPROC tokens are used in place of DONE tokens.

**Token Stream-Specific Rules:**

1086. TokenType = BYTE

      Status = USHORT

      CurCmd = USHORT

      DoneRowCount = LONG / ULONGLONG; (Changed to ULONGLONG in TDS 7.2)

The type of the **DoneRowCount** element depends on the version of TDS.

**Token Stream Definition:**

1090. DONE = TokenType

      Status

      CurCmd

      DoneRowCount

**Token Stream Parameter Details:**

+----------+-----------------------------------------------------------+
| P        | Description                                               |
| arameter |                                                           |
+==========+===========================================================+
| T        | DONE_TOKEN                                                |
| okenType |                                                           |
+----------+-----------------------------------------------------------+
| Status   | The Status field MUST be a bitwise \'OR\' of the          |
|          | following:                                                |
|          |                                                           |
|          | -   0x00: DONE_FINAL. This DONE is the final DONE in the  |
|          |     request.                                              |
|          |                                                           |
|          | -   0x1: DONE_MORE. This DONE message is not the final    |
|          |     DONE message in the response. Subsequent data streams |
|          |     to follow.                                            |
|          |                                                           |
|          | -   0x2: DONE_ERROR. An error occurred on the current SQL |
|          |     statement. A preceding ERROR token SHOULD be sent     |
|          |     when this bit is set.                                 |
|          |                                                           |
|          | -   0x4: DONE_INXACT. A transaction is in                 |
|          |     progress.[\<46\>](\l)                                 |
|          |                                                           |
|          | -   0x10: DONE_COUNT. The DoneRowCount value is valid.    |
|          |     This is used to distinguish between a valid value of  |
|          |     0 for DoneRowCount or just an initialized variable.   |
|          |                                                           |
|          | -   0x20: DONE_ATTN. The DONE message is a server         |
|          |     acknowledgement of a client ATTENTION message.        |
|          |                                                           |
|          | -   0x100: DONE_SRVERROR. Used in place of DONE_ERROR     |
|          |     when an error occurred on the current SQL statement,  |
|          |     which is severe enough to require the [**result       |
|          |     set**](#gt_c8a27238-8ccc-442b-9604-75f74d3e6b3d), if  |
|          |     any, to be discarded.                                 |
+----------+-----------------------------------------------------------+
| CurCmd   | The token of the current SQL statement. The token value   |
|          | is provided and controlled by the application layer,      |
|          | which utilizes TDS. The TDS layer does not evaluate the   |
|          | value.                                                    |
+----------+-----------------------------------------------------------+
| Done     | The count of rows that were affected by the SQL           |
| RowCount | statement. The value of DoneRowCount is valid if the      |
|          | value of Status includes DONE_COUNT.[\<47\>](\l)          |
+----------+-----------------------------------------------------------+

#### DONEINPROC

**Token Stream Name:**

1094. DONEINPROC

**Token Stream Function:**

Indicates the completion status of a [**SQL
statement**](#gt_dc5ca224-43ec-4b44-9dba-726d6fd6057d) within a
[**stored procedure**](#gt_324d32b3-f4f3-41c9-b695-78c498094fb7).

**Token Stream Comments**

-   The token value is 0xFF.

-   A DONEINPROC token is sent for each executed SQL statement within a
    stored procedure.

-   A DONEINPROC token MUST be followed by another DONEPROC token or a
    DONEINPROC token.

**Token Stream-Specific Rules:**

1095. TokenType = BYTE

      Status = USHORT

      CurCmd = USHORT

      DoneRowCount = LONG / ULONGLONG; (Changed to ULONGLONG in TDS 7.2)

The type of the **DoneRowCount** element depends on the version of TDS.

**Token Stream Definition:**

1099. DONEINPROC = TokenType

      Status

      CurCmd

      DoneRowCount

**Token Stream Parameter Details:**

+----------+-----------------------------------------------------------+
| P        | Description                                               |
| arameter |                                                           |
+==========+===========================================================+
| T        | DONEINPROC_TOKEN                                          |
| okenType |                                                           |
+----------+-----------------------------------------------------------+
| Status   | The Status field MUST be a bitwise \'OR\' of the          |
|          | following:                                                |
|          |                                                           |
|          | -   0x1: DONE_MORE. This DONEINPROC message is not the    |
|          |     final DONE/DONEPROC/DONEINPROC message in the         |
|          |     response; more data streams are to follow.            |
|          |                                                           |
|          | -   0x2: DONE_ERROR. An error occurred on the current SQL |
|          |     statement or execution of a stored procedure was      |
|          |     interrupted. A preceding ERROR token SHOULD be sent   |
|          |     when this bit is set.                                 |
|          |                                                           |
|          | -   0x4: DONE_INXACT. A transaction is in                 |
|          |     progress.[\<48\>](\l)                                 |
|          |                                                           |
|          | -   0x10: DONE_COUNT. The DoneRowCount value is valid.    |
|          |     This is used to distinguish between a valid value of  |
|          |     0 for DoneRowCount or just an initialized variable.   |
|          |                                                           |
|          | -   0x100: DONE_SRVERROR. Used in place of DONE_ERROR     |
|          |     when an error occurred on the current SQL statement   |
|          |     that is severe enough to require the [**result        |
|          |     set**](#gt_c8a27238-8ccc-442b-9604-75f74d3e6b3d), if  |
|          |     any, to be discarded.                                 |
+----------+-----------------------------------------------------------+
| CurCmd   | The token of the current SQL statement. The token value   |
|          | is provided and controlled by the application layer,      |
|          | which utilizes TDS. The TDS layer does not evaluate the   |
|          | value.                                                    |
+----------+-----------------------------------------------------------+
| Done     | The count of rows that were affected by the SQL           |
| RowCount | statement. The value of DoneRowCount is valid if the      |
|          | value of Status includes DONE_COUNT.                      |
+----------+-----------------------------------------------------------+

#### DONEPROC

**Token Stream Name:**

1103. DONEPROC

**Token Stream Function:**

Indicates the completion status of a [**stored
procedure**](#gt_324d32b3-f4f3-41c9-b695-78c498094fb7). This is also
generated for stored procedures executed through [**SQL
statements**](#gt_dc5ca224-43ec-4b44-9dba-726d6fd6057d).

**Token Stream Comments:**

-   The token value is 0xFE.

-   A DONEPROC token is sent when all the SQL statements within a stored
    procedure have been executed.

-   A DONEPROC token can be followed by another DONEPROC token or a
    DONEINPROC only if the DONE_MORE bit is set in the Status value.

-   There is a separate DONEPROC token sent for each stored procedure
    called.

**Token Stream-Specific Rules:**

1104. TokenType = BYTE

      Status = USHORT

      CurCmd = USHORT

      DoneRowCount = LONG / ULONGLONG; (Changed to ULONGLONG in TDS 7.2)

The type of the **DoneRowCount** element depends on the version of TDS.

**Token Stream Definition:**

1108. DONEPROC = TokenType

      Status

      CurCmd

      DoneRowCount

**Token Stream Parameter Details:**

+----------+-----------------------------------------------------------+
| P        | Description                                               |
| arameter |                                                           |
+==========+===========================================================+
| T        | DONEPROC_TOKEN                                            |
| okenType |                                                           |
+----------+-----------------------------------------------------------+
| Status   | The Status field MUST be a bitwise \'OR\' of the          |
|          | following:                                                |
|          |                                                           |
|          | -   0x00: DONE_FINAL. This DONEPROC is the final DONEPROC |
|          |     in the request.                                       |
|          |                                                           |
|          | -   0x1: DONE_MORE. This DONEPROC message is not the      |
|          |     final DONEPROC message in the response; more data     |
|          |     streams are to follow.                                |
|          |                                                           |
|          | -   0x2: DONE_ERROR. An error occurred on the current     |
|          |     stored procedure. A preceding ERROR token SHOULD be   |
|          |     sent when this bit is set.                            |
|          |                                                           |
|          | -   0x4: DONE_INXACT. A transaction is in                 |
|          |     progress.[\<49\>](\l)                                 |
|          |                                                           |
|          | -   0x10: DONE_COUNT. The DoneRowCount value is valid.    |
|          |     This is used to distinguish between a valid value of  |
|          |     0 for DoneRowCount or just an initialized variable.   |
|          |                                                           |
|          | -   0x80: DONE_RPCINBATCH. This DONEPROC message is       |
|          |     associated with an RPC within a set of batched RPCs.  |
|          |     This flag is not set on the last RPC in the RPC       |
|          |     batch.                                                |
|          |                                                           |
|          | -   0x100: DONE_SRVERROR. Used in place of DONE_ERROR     |
|          |     when an error occurred on the current stored          |
|          |     procedure, which is severe enough to require the      |
|          |     [**result                                             |
|          |     set**](#gt_c8a27238-8ccc-442b-9604-75f74d3e6b3d), if  |
|          |     any, to be discarded.                                 |
+----------+-----------------------------------------------------------+
| CurCmd   | The token of the SQL statement for executing stored       |
|          | procedures. The token value is provided and controlled by |
|          | the application layer, which utilizes TDS. The TDS layer  |
|          | does not evaluate the value.                              |
+----------+-----------------------------------------------------------+
| Done     | The count of rows that were affected by the command. The  |
| RowCount | value of DoneRowCount is valid if the value of Status     |
|          | includes DONE_COUNT.                                      |
+----------+-----------------------------------------------------------+

#### ENVCHANGE

**Token Stream Name:**

1112. ENVCHANGE

**Token Stream Function:**

A notification of an environment change (for example, database,
language, and so on).

**Token Stream Comments:**

-   The token value is 0xE3.

-   Includes old and new environment values.

-   Type 4 (Packet size) is sent in response to a LOGIN7 message. The
    server MAY send a value different from the packet size requested by
    the client. That value MUST be greater than or equal to 512 and
    smaller than or equal to 32767. Both the client and the server MUST
    start using this value for packet size with the message following
    the login response message.

-   Type 13 (Database Mirroring) is sent in response to a LOGIN7 message
    whenever connection is requested to a database that it is being
    served as primary in real-time log shipping. The ENVCHANGE stream
    reflects the name of the partner node of the database that is being
    log shipped.

-   Type 15 (Promote Transaction) is sent in response to [**transaction
    manager**](#gt_4553803e-9d8d-407c-ad7d-9e65e01d6eb3) requests with
    requests of type 6 (TM_PROMOTE_XACT).

-   Type 16 (Transaction Manager Address) is sent in response to
    transaction manager requests with requests of type 0
    (TM_GET_DTC_ADDRESS).

-   Type 20 (Routing) is sent in response to a LOGIN7 message when the
    server wants to route the client to an alternate server. The
    ENVCHANGE stream returns routing information for the alternate
    server. If the server decides to send the Routing ENVCHANGE token,
    the Routing ENVCHANGE token MUST be sent after the LOGINACK token in
    the login response.

-   Type 21 (Enhanced Routing) is sent in response to a LOGIN7 message
    when the server wants to route the client to a specific database at
    an alternate server. The ENVCHANGE stream returns routing
    information for the alternate server and the alternate database. If
    the server decides to send the Enhanced Routing ENVCHANGE token, the
    Enhanced Routing ENVCHANGE token MUST be sent after the LOGINACK
    token in the login response.

-   The server may only send one of Type 20 (Routing) or Type 21
    (Enhanced Routing) in a login response.

**Token Stream-Specific Rules:**

1113. TokenType = BYTE

      Length = USHORT

      Type = BYTE

      EnvValueData = Type

      NewValue

      \[OldValue\]

**Token Stream Definition:**

1121. ENVCHANGE = TokenType

      Length

      EnvValueData

**Token Stream Parameter Details**

+--------+-------------------------------------------------------------+
| Par    | Description                                                 |
| ameter |                                                             |
+========+=============================================================+
| Tok    | ENVCHANGE_TOKEN                                             |
| enType |                                                             |
+--------+-------------------------------------------------------------+
| Length | The total length of the ENVCHANGE data stream               |
|        | (EnvValueData).                                             |
+--------+-------------------------------------------------------------+
| Type   | The type of environment change:                             |
|        |                                                             |
|        | **Note** Types 8 to 19 were introduced in TDS 7.2. Type 20  |
|        | was introduced in TDS 7.4.                                  |
|        |                                                             |
|        | -   1: Database                                             |
|        |                                                             |
|        | -   2: Language                                             |
|        |                                                             |
|        | -   3: Character set                                        |
|        |                                                             |
|        | -   4: Packet size                                          |
|        |                                                             |
|        | -   5:                                                      |
|        |     [**Unicode**](#gt_c305d0ab-8b94-461a-bd76-13b40cb8c4d8) |
|        |     data sorting local id                                   |
|        |                                                             |
|        | -   6: Unicode data sorting comparison flags                |
|        |                                                             |
|        | -   7: SQL Collation                                        |
|        |                                                             |
|        | -   8: Begin Transaction (described in                      |
|        |     [\[MSD                                                  |
|        | N-BEGIN\]](https://go.microsoft.com/fwlink/?LinkId=144544)) |
|        |                                                             |
|        | -   9: Commit Transaction (described in                     |
|        |     [\[MSDN                                                 |
|        | -COMMIT\]](https://go.microsoft.com/fwlink/?LinkId=144542)) |
|        |                                                             |
|        | -   10: Rollback Transaction                                |
|        |                                                             |
|        | -   11: Enlist DTC Transaction                              |
|        |                                                             |
|        | -   12: Defect Transaction                                  |
|        |                                                             |
|        | -   13: Real Time Log Shipping                              |
|        |                                                             |
|        | -   15: Promote Transaction                                 |
|        |                                                             |
|        | -   16: Transaction Manager Address[\<50\>](\l)             |
|        |                                                             |
|        | -   17: Transaction ended                                   |
|        |                                                             |
|        | -   18: RESETCONNECTION/RESETCONNECTIONSKIPTRAN Completion  |
|        |     Acknowledgement                                         |
|        |                                                             |
|        | -   19: Sends back name of user instance started per login  |
|        |     request                                                 |
|        |                                                             |
|        | -   20: Sends routing information to client                 |
|        |                                                             |
|        | -   21: Sends routing information and database name to      |
|        |     client                                                  |
+--------+-------------------------------------------------------------+

+-------------------+----------+---------------------------------------+
| Type              | Old      | New Value                             |
|                   | Value    |                                       |
+===================+==========+=======================================+
| 1: Database       | OLDVALUE | NEWVALUE = B_VARCHAR                  |
|                   | =        |                                       |
|                   | B        |                                       |
|                   | _VARCHAR |                                       |
+-------------------+----------+---------------------------------------+
| 2: Language       | OLDVALUE | NEWVALUE = B_VARCHAR                  |
|                   | =        |                                       |
|                   | B        |                                       |
|                   | _VARCHAR |                                       |
+-------------------+----------+---------------------------------------+
| 3: Character Set  | OLDVALUE | NEWVALUE = B_VARCHAR                  |
|                   | =        |                                       |
|                   | B        |                                       |
|                   | _VARCHAR |                                       |
+-------------------+----------+---------------------------------------+
| 4: Packet Size    | OLDVALUE | NEWVALUE = B_VARCHAR                  |
|                   | =        |                                       |
|                   | B        |                                       |
|                   | _VARCHAR |                                       |
+-------------------+----------+---------------------------------------+
| 5: Unicode data   | OLDVALUE | NEWVALUE = B_VARCHAR                  |
| sorting local id  | = %x00   |                                       |
+-------------------+----------+---------------------------------------+
| 6: Unicode data   | OLDVALUE | NEWVALUE = B_VARCHAR                  |
| sorting           | = %x00   |                                       |
| comparison flags  |          |                                       |
+-------------------+----------+---------------------------------------+
| 7: SQL Collation  | OLDVALUE | NEWVALUE = B_VARBYTE                  |
|                   | =        |                                       |
|                   | B        |                                       |
|                   | _VARBYTE |                                       |
+-------------------+----------+---------------------------------------+
| 8: Begin          | OLDVALUE | NEWVALUE = B_VARBYTE                  |
| Transaction       | = %x00   |                                       |
+-------------------+----------+---------------------------------------+
| 9: Commit         | OLDVALUE | NEWVALUE = %0x00                      |
| Transaction       | =        |                                       |
|                   | B        |                                       |
|                   | _VARBYTE |                                       |
+-------------------+----------+---------------------------------------+
| 10: Rollback      | OLDVALUE | NEWVALUE = %x00                       |
| Transaction       | =        |                                       |
|                   | B        |                                       |
|                   | _VARBYTE |                                       |
+-------------------+----------+---------------------------------------+
| 11: Enlist DTC    | OLDVALUE | NEWVALUE = %x00                       |
| Transaction       | =        |                                       |
|                   | B        |                                       |
|                   | _VARBYTE |                                       |
+-------------------+----------+---------------------------------------+
| 12: Defect        | OLDVALUE | NEWVALUE = B_VARBYTE                  |
| Transaction       | = %x00   |                                       |
+-------------------+----------+---------------------------------------+
| 13: Database      | OLDVALUE | PARTNER_NODE = B_VARCHAR              |
| Mirroring Partner | = %x00   |                                       |
|                   |          | NEWVALUE = PARTNER_NODE               |
+-------------------+----------+---------------------------------------+
| 15: Promote       | OLDVALUE | DTC_TOKEN = L_VARBYTE;                |
| Transaction       | = %x00   |                                       |
|                   |          | NEWVALUE = DTC_TOKEN                  |
+-------------------+----------+---------------------------------------+
| 16: Transaction   | OLDVALUE | XACT_MANAGER_ADDRESS = B_VARBYTE      |
| Manager Address   | = %x00   |                                       |
| (not used)        |          | NEWVALUE = XACT_MANAGER_ADDRESS       |
+-------------------+----------+---------------------------------------+
| 17: Transaction   | OLDVALUE | NEWVALUE = %x00                       |
| Ended             | =        |                                       |
|                   | B        |                                       |
|                   | _VARBYTE |                                       |
+-------------------+----------+---------------------------------------+
| 18: Reset         | OLDVALUE | NEWVALUE = %x00                       |
| Completion        | = %x00   |                                       |
| Acknowledgement   |          |                                       |
+-------------------+----------+---------------------------------------+
| 19: Sends back    | OLDVALUE | NEWVALUE = B_VARCHAR                  |
| info of user      | = %x00   |                                       |
| instance for      |          |                                       |
| logins (login7)   |          |                                       |
| requesting so.    |          |                                       |
+-------------------+----------+---------------------------------------+
| 20: Routing       | OLDVALUE | Protocol = BYTE                       |
|                   | = %x00   |                                       |
|                   | %x00     | ProtocolProperty = USHORT             |
|                   |          |                                       |
|                   |          | AlternateServer = US_VARCHAR          |
|                   |          |                                       |
|                   |          | Protocol MUST be 0, specifying TCP-IP |
|                   |          | protocol. ProtocolProperty represents |
|                   |          | the TCP-IP port when Protocol is 0. A |
|                   |          | ProtocolProperty value of zero is not |
|                   |          | allowed when Protocol is TCP-IP.      |
|                   |          |                                       |
|                   |          | RoutingDataValue = Protocol           |
|                   |          |                                       |
|                   |          | ProtocolProperty                      |
|                   |          |                                       |
|                   |          | AlternateServer                       |
|                   |          |                                       |
|                   |          | RoutingDataValueLength = USHORT       |
|                   |          |                                       |
|                   |          | RoutingDataValueLength is the total   |
|                   |          | length, in bytes, of the following    |
|                   |          | fields: Protocol, ProtocolProperty,   |
|                   |          | and AlternateServer.                  |
|                   |          |                                       |
|                   |          | RoutingData = RoutingDataValueLength  |
|                   |          |                                       |
|                   |          | \[RoutingDataValue\]                  |
|                   |          |                                       |
|                   |          | NEWVALUE = RoutingData                |
+-------------------+----------+---------------------------------------+
| 21: Enhanced      | OLDVALUE | Protocol = BYTE                       |
| Routing           | = %x00   |                                       |
|                   | %x00     | ProtocolProperty = USHORT             |
|                   |          |                                       |
|                   |          | AlternateServer = US_VARCHAR          |
|                   |          |                                       |
|                   |          | AlternateDatabase = US_VARCHAR        |
|                   |          |                                       |
|                   |          | Protocol MUST be 0, specifying TCP-IP |
|                   |          | protocol. ProtocolProperty represents |
|                   |          | the TCP-IP port when Protocol is 0. A |
|                   |          | ProtocolProperty value of zero is not |
|                   |          | allowed when Protocol is TCP-IP.      |
|                   |          |                                       |
|                   |          | AlternateDatabase must not exceed 128 |
|                   |          | characters.                           |
|                   |          |                                       |
|                   |          | RoutingDataValue = Protocol           |
|                   |          |                                       |
|                   |          | ProtocolProperty                      |
|                   |          |                                       |
|                   |          | AlternateServer                       |
|                   |          |                                       |
|                   |          | AlternateDatabase                     |
|                   |          |                                       |
|                   |          | RoutingDataValueLength = USHORT       |
|                   |          |                                       |
|                   |          | RoutingDataValueLength is the total   |
|                   |          | length, in bytes, of the following    |
|                   |          | fields: Protocol, ProtocolProperty,   |
|                   |          | AlternateServer, and                  |
|                   |          | AlternateDatabase.                    |
|                   |          |                                       |
|                   |          | RoutingData = RoutingDataValueLength  |
|                   |          |                                       |
|                   |          | \[RoutingDataValue\]                  |
|                   |          |                                       |
|                   |          | NEWVALUE = RoutingData                |
+-------------------+----------+---------------------------------------+

**Notes**

-   For types 1, 2, 3, 4, 5, 6, 13, and 19, the payload is a Unicode
    string; the LENGTH always reflects the number of bytes.

-   ENVCHANGE types 3, 5, and 6 are only sent back to clients running
    TDS 7.0 or earlier.

-   For Types 8, 9, 10, 11, and 12, the ENVCHANGE event is returned only
    if the transaction lifetime is controlled by the user, for example,
    explicit transaction commands, including transactions started by SET
    IMPLICIT_TRANSACTIONS ON.

-   For transactions started/committed under auto commit, no stream is
    generated.

-   For operations that change only the value of @@trancount, no
    ENVCHANGE stream is generated.

-   The payload of NEWVALUE for ENVCHANGE types 8, 11, and 17 and the
    payload of OLDVALUE for ENVCHANGE types 9, 10, and 12 is a
    ULONGLONG.

-   ENVCHANGE type 11 is sent by the server to confirm that it has
    joined a distributed transaction as requested through a
    TM_PROPAGATE_XACT request from the client.

-   ENVCHANGE type 12 is only sent when a batch defects from either a
    DTC or bound session transaction.

-   LENGTH for ENVCHANGE type 15 is sent as 0x01 indicating only the
    length of the type token. Client drivers are responsible for reading
    the additional payload if type is 15.

-   ENVCHANGE type 17 is sent when a batch is used that specified a
    descriptor for a transaction that has ended. This is only sent in
    the bound session case. For information about using bound sessions,
    see
    [\[MSDN-BOUND\]](https://go.microsoft.com/fwlink/?LinkId=144543).

-   ENVCHANGE type 18 always produces empty (0x00) old and new values.
    It simply acknowledges completion of execution of a
    RESETCONNECTION/RESETCONNECTIONSKIPTRAN request.

-   ENVCHANGE type 19 is sent after LOGIN and after
    /RESETCONNECTION/RESETCONNECTIONSKIPTRAN when a client has requested
    use of user instances. It is sent prior to the LOGINACK token.

-   ENVCHANGE type 20 can be sent back to a client running TDS 7.4 or
    later regardless of whether the fReadOnlyIntent bit is set in the
    preceding LOGIN7 record. If a client is running TDS 7.1 to 7.3, type
    20 can be sent only if the fReadOnlyIntent bit is set in the
    preceding LOGIN7 record.

-   ENVCHANGE 21 is introduced in TDS 7.4. It may only be sent back to a
    client if the client sends the ENHANCEDROUTINGSUPPORT FeatureExt. It
    can be sent back to a client regardless of whether the
    fReadOnlyIntent bit is set in the preceding LOGIN7 record.

#### ERROR

**Token Stream Name:**

1124. ERROR

**Token Stream Function:**

Used to send an error message to the client.

**Token Stream Comments:**

-   The token value is 0xAA.

**Token Stream-Specific Rules:**

1125. TokenType = BYTE

      Length = USHORT

      Number = LONG

      State = BYTE

      Class = BYTE

      MsgText = US_VARCHAR

      ServerName = B_VARCHAR

      ProcName = B_VARCHAR

      LineNumber = USHORT / LONG; (Changed to LONG in TDS 7.2)

The type of the **LineNumber** element depends on the version of TDS.

**Token Stream Definition:**

1134. ERROR = TokenType

      Length

      Number

      State

      Class

      MsgText

      ServerName

      ProcName

      LineNumber

**Token Stream Parameter Details**

  --------------------------------------------------------------------------
  Parameter    Description
  ------------ -------------------------------------------------------------
  TokenType    ERROR_TOKEN

  Length       The total length of the ERROR data stream, in bytes.

  Number       The error number.[\<51\>](\l)

  State        The error state, used as a modifier to the error number.

  Class        The class (severity) of the error. A class of less than 10
               indicates an informational message.

  MsgText      The message text length and message text using US_VARCHAR
               format.

  ServerName   The server name length and server name using B_VARCHAR
               format.

  ProcName     The [**stored
               procedure**](#gt_324d32b3-f4f3-41c9-b695-78c498094fb7) name
               length and the stored procedure name using B_VARCHAR format.

  LineNumber   The line number in the SQL batch or stored procedure that
               caused the error. Line numbers begin at 1. If the line number
               is not applicable to the message, the value of LineNumber is
               0.
  --------------------------------------------------------------------------

+-----+----------------------------------------------------------------+
| Cl  | Description                                                    |
| ass |                                                                |
| le  |                                                                |
| vel |                                                                |
+=====+================================================================+
| 0-9 | Informational messages that return status information or       |
|     | report errors that are not severe.[\<52\>](\l)                 |
+-----+----------------------------------------------------------------+
| 10  | Informational messages that return status information or       |
|     | report errors that are not severe.[\<53\>](\l)                 |
+-----+----------------------------------------------------------------+
| 11  | Errors that can be corrected by the user.                      |
| -16 |                                                                |
+-----+----------------------------------------------------------------+
| 11  | The given object or entity does not exist.                     |
+-----+----------------------------------------------------------------+
| 12  | A special severity for [**SQL                                  |
|     | statements**](#gt_dc5ca224-43ec-4b44-9dba-726d6fd6057d) that   |
|     | do not use locking because of special options. In some cases,  |
|     | read operations performed by these SQL statements could result |
|     | in inconsistent data, because locks are not taken to guarantee |
|     | consistency.                                                   |
+-----+----------------------------------------------------------------+
| 13  | Transaction deadlock errors.                                   |
+-----+----------------------------------------------------------------+
| 14  | Security-related errors, such as permission denied.            |
+-----+----------------------------------------------------------------+
| 15  | Syntax errors in the SQL statement.                            |
+-----+----------------------------------------------------------------+
| 16  | General errors that can be corrected by the user.              |
+-----+----------------------------------------------------------------+
| 17  | Software errors that cannot be corrected by the user. These    |
| -19 | errors require system administrator action.                    |
+-----+----------------------------------------------------------------+
| 17  | The SQL statement caused the database server to run out of     |
|     | resources (such as memory, locks, or disk space for the        |
|     | database) or to exceed some limit set by the system            |
|     | administrator.                                                 |
+-----+----------------------------------------------------------------+
| 18  | There is a problem in the Database Engine software, but the    |
|     | SQL statement completes execution, and the connection to the   |
|     | instance of the Database Engine is maintained. System          |
|     | administrator action is required.                              |
+-----+----------------------------------------------------------------+
| 19  | A non-configurable Database Engine limit has been exceeded and |
|     | the current SQL batch has been terminated. Error messages with |
|     | a severity level of 19 or higher stop the execution of the     |
|     | current SQL batch. Severity level 19 errors are rare and can   |
|     | be corrected only by the system administrator. Error messages  |
|     | with a severity level from 19 through 25 are written to the    |
|     | error log.                                                     |
+-----+----------------------------------------------------------------+
| 20  | System problems have occurred. These are fatal errors, which   |
| -25 | means the Database Engine task that was executing a SQL batch  |
|     | is no longer running. The task records information about what  |
|     | occurred and then terminates. In most cases, the application   |
|     | connection to the instance of the Database Engine can also     |
|     | terminate. If this happens, depending on the problem, the      |
|     | application might not be able to reconnect.                    |
|     |                                                                |
|     | Error messages in this range can affect all of the processes   |
|     | accessing data in the same database and might indicate that a  |
|     | database or object is damaged. Error messages with a severity  |
|     | level from 19 through 25 are written to the error log.         |
+-----+----------------------------------------------------------------+
| 20  | Indicates that a SQL statement has encountered a problem.      |
|     | Because the problem has affected only the current task, it is  |
|     | unlikely that the database itself has been damaged.            |
+-----+----------------------------------------------------------------+
| 21  | Indicates that a problem has been encountered that affects all |
|     | tasks in the current database, but it is unlikely that the     |
|     | database itself has been damaged.                              |
+-----+----------------------------------------------------------------+
| 22  | Indicates that the table or index specified in the message has |
|     | been damaged by a software or hardware problem.                |
|     |                                                                |
|     | Severity level 22 errors occur rarely. If one occurs, run DBCC |
|     | CHECKDB to determine whether other objects in the database are |
|     | also damaged. The problem might be in the buffer cache only    |
|     | and not on the disk itself. If so, restarting the instance of  |
|     | the Database Engine corrects the problem. To continue working, |
|     | reconnect to the instance of the Database Engine; otherwise,   |
|     | use DBCC to repair the problem. In some cases, restoration of  |
|     | the database might be required.                                |
|     |                                                                |
|     | If restarting the instance of the Database Engine does not     |
|     | correct the problem, then the problem is on the disk.          |
|     | Sometimes destroying the object specified in the error message |
|     | can solve the problem. For example, if the message reports     |
|     | that the instance of the Database Engine has found a row with  |
|     | a length of 0 in a non-clustered index, delete the index and   |
|     | rebuild it.                                                    |
+-----+----------------------------------------------------------------+
| 23  | Indicates that the integrity of the entire database is in      |
|     | question because of a hardware or software problem.            |
|     |                                                                |
|     | Severity level 23 errors occur rarely. If one occurs, run DBCC |
|     | CHECKDB to determine the extent of the damage. The problem     |
|     | might be in the cache only and not on the disk itself. If so,  |
|     | restarting the instance of the Database Engine corrects the    |
|     | problem. To continue working, reconnect to the instance of the |
|     | Database Engine; otherwise, use DBCC to repair the problem. In |
|     | some cases, restoration of the database might be required.     |
+-----+----------------------------------------------------------------+
| 24  | Indicates a media failure. The system administrator might have |
|     | to restore the database or resolve a hardware issue.           |
+-----+----------------------------------------------------------------+

If an error is produced within a [**result
set**](#gt_c8a27238-8ccc-442b-9604-75f74d3e6b3d), the ERROR token is
sent before the DONE token for the SQL statement, and such DONE token is
sent with the error bit set.

#### FEATUREEXTACK

**Token Stream Name:**

1143. FEATUREEXTACK

**Token Stream Function:**

Introduced in TDS 7.4, FEATUREEXTACK is used to send an optional
acknowledge message to the client for features that are defined in
FeatureExt. The token stream is sent only along with the LOGINACK in a
Login Response message.

**Token Stream Comments:**

-   The token value is 0xAE.

**Token Stream-Specific Rules:**

1144. TokenType = BYTE

      FeatureId = BYTE

      FeatureAckDataLen = DWORD

      FeatureAckData = \*BYTE

      TERMINATOR = %xFF ; signal of end of feature ack data

      FeatureAckOpt = (FeatureId

      FeatureAckDataLen

      FeatureAckData)

      /

      TERMINATOR

**Token Stream Definition:**

1157. FEATUREEXTACK = TokenType

      1\*FeatureAckOpt

**Token Stream Parameter Details**

+------------+---------------------------------------------------------+
| Parameter  | Description                                             |
+============+=========================================================+
| TokenType  | FEATUREEXTACK_TOKEN                                     |
+------------+---------------------------------------------------------+
| FeatureId  | The unique identifier number of a feature. Each feature |
|            | MUST use the same ID number here as in FeatureExt. If   |
|            | the client did not send a request for a specific        |
|            | feature but the FeatureId is returned, the client MUST  |
|            | consider it as a TDS Protocol error and MUST terminate  |
|            | the connection.                                         |
|            |                                                         |
|            | Each feature defines its own logic if it wants to use   |
|            | FeatureAckOpt to send information back to the client    |
|            | during the login response. The features available to    |
|            | use by a FeatureId are defined in the following table.  |
+------------+---------------------------------------------------------+
| Feature    | The length of FeatureAckData, in bytes.                 |
| AckDataLen |                                                         |
+------------+---------------------------------------------------------+
| Feat       | The acknowledge data of a specific feature. Each        |
| ureAckData | feature SHOULD define its own data format in the        |
|            | FEATUREEXTACK token if it is selected to acknowledge    |
|            | the feature.                                            |
+------------+---------------------------------------------------------+

The following table describes the FeatureExtAck feature option and
description.

+----------------------+-----------------------------------------------+
| FeatureId            | FeatureExtData Description                    |
+======================+===============================================+
| %0x00                | Reserved.                                     |
+----------------------+-----------------------------------------------+
| %0x01                | Session Recovery feature. Content is defined  |
|                      | as follows:                                   |
| (SESSIONRECOVERY)    |                                               |
|                      | 1159. InitSessionStateData =                  |
| (introduced in TDS   |       SessionStateDataSet                     |
| 7.4)                 |                                               |
|                      |       FeatureAckData = InitSessionStateData   |
|                      |                                               |
|                      | SessionStateDataSet is described in section   |
|                      | [2.2.7.21                                     |
|                      | ](#Section_626fbe19f3564599ba17c70f44005106). |
|                      | The length of SessionStateDataSet is          |
|                      | specified by the corresponding                |
|                      | FeatureAckDataLen.                            |
|                      |                                               |
|                      | On a recovery connection, the client sends a  |
|                      | login request with SessionRecoveryDataToBe.   |
|                      | The server MUST set the session state as      |
|                      | requested by the client. If the server cannot |
|                      | do so, the server MUST fail the login request |
|                      | and terminate the connection.                 |
+----------------------+-----------------------------------------------+
| %0x02                | Whenever a login response stream is sent for  |
|                      | a TDS connection whose login request includes |
| (                    | a FEDAUTH FeatureExt, the server login        |
| FEDAUTH)[\<54\>](\l) | response message stream MUST include a        |
|                      | FEATUREEXTACK token, and the FEATUREEXTACK    |
|                      | token stream MUST include the FEDAUTH         |
|                      | FeatureId. The format is described below      |
|                      | based on the bFedAuthLibrary that is used in  |
|                      | FEDAUTH FeatureExt.                           |
|                      |                                               |
|                      | When the bFedAuthLibrary is Live ID Compact   |
|                      | Token, the format is as follows:              |
|                      |                                               |
|                      | 1161. Nonce = 32BYTE                          |
|                      |                                               |
|                      |       Signature = 32BYTE                      |
|                      |                                               |
|                      |       FeatureAckData = Nonce                  |
|                      |                                               |
|                      |       Signature                               |
|                      |                                               |
|                      | Nonce: The client-specified nonce in          |
|                      | PRELOGIN.                                     |
|                      |                                               |
|                      | Signature: The HMAC-SHA-256                   |
|                      | [\[RFC6234\]](ht                              |
|                      | tps://go.microsoft.com/fwlink/?LinkId=328921) |
|                      | of the client-specified nonce, using the      |
|                      | session key retrieved from the [**federated   |
|                      | authentication                                |
|                      | **](#gt_5ae22a0e-5ff4-441b-80d4-224ef4dd4d19) |
|                      | context as the shared secret.                 |
|                      |                                               |
|                      | When the bFedAuthLibrary is Security Token,   |
|                      | the format is as follows:                     |
|                      |                                               |
|                      | 1166. Nonce = 32BYTE                          |
|                      |                                               |
|                      |       FeatureAckData = \[Nonce\]              |
|                      |                                               |
|                      | Nonce: The client-specified nonce in          |
|                      | PRELOGIN. This field MUST be present if the   |
|                      | client's PRELOGIN message included a NONCE    |
|                      | field. Otherwise, this field MUST NOT be      |
|                      | present.                                      |
+----------------------+-----------------------------------------------+
| %0x04                | The presence of the COLUMNENCRYPTION          |
|                      | FeatureExt SHOULD[\<55\>](\l) indicate that   |
| (COLUMNENCRYPTION)   | the client is capable of performing           |
|                      | cryptographic operations on data. The feature |
| (introduced in TDS   | data is described as follows:                 |
| 7.4)                 |                                               |
|                      | 1169. Length = BYTE                           |
|                      |                                               |
|                      |       COLUMNENCRYPTION_VERSION = BYTE         |
|                      |                                               |
|                      |       FeatureData = COLUMNENCRYPTION_VERSION  |
|                      |                                               |
|                      |       \[Length EnclaveType\]                  |
|                      |                                               |
|                      | COLUMNENCRYPTION_VERSION: This field defines  |
|                      | the cryptographic protocol version that the   |
|                      | client understands. The values of this field  |
|                      | are as follows:                               |
|                      |                                               |
|                      | -   1 = The client supports column encryption |
|                      |     without [**enclave                        |
|                      |     computations*                             |
|                      | *](#gt_6fe5534f-5cd8-4ab6-aba4-637a4344eda0). |
|                      |                                               |
|                      | ```{=html}                                    |
|                      | <!-- -->                                      |
|                      | ```                                           |
|                      | -   2 = The client SHOULD[\<56\>](\l) support |
|                      |     column encryption when encrypted data     |
|                      |     require enclave computations.             |
|                      |                                               |
|                      | ```{=html}                                    |
|                      | <!-- -->                                      |
|                      | ```                                           |
|                      | -   3 = The client SHOULD[\<57\>](\l) support |
|                      |     column encryption when encrypted data     |
|                      |     require enclave computations with the     |
|                      |     additional ability to cache column        |
|                      |     encryption keys that are to be sent to    |
|                      |     the enclave and the ability to retry      |
|                      |     queries when the keys sent by the client  |
|                      |     do not match what is needed for the query |
|                      |     to run.                                   |
|                      |                                               |
|                      | EnclaveType: This field is a string that      |
|                      | SHOULD\<58\> be populated by the server and   |
|                      | used by the client to identify the type of    |
|                      | [**enclave                                    |
|                      | **](#gt_ef41e9e0-3e5d-432f-96d6-39515bdc5340) |
|                      | that the server is configured to use. During  |
|                      | login for the initial connection, the client  |
|                      | can request COLUMNENCRYPTION with **Length**  |
|                      | as 1 and COLUMNENCRYPTION_VERSION as either 1 |
|                      | or 2. When the client requests                |
|                      | COLUMNENCRYPTION_VERSION as 2, the server     |
|                      | MUST return COLUMNENCRYPTION_VERSION as 2     |
|                      | together with the value of **EnclaveType**,   |
|                      | if the server contains an enclave that is     |
|                      | configured for use. If **EnclaveType** is not |
|                      | returned and the column encryption version is |
|                      | returned as 2, the client driver MUST raise   |
|                      | an error.                                     |
+----------------------+-----------------------------------------------+
| %0x05                | Whenever a login response stream is sent for  |
|                      | a TDS connection whose login request includes |
| (GLOBALTRANS         | a GLOBALTRANSACTIONS FeatureExt token, the    |
| ACTIONS)[\<59\>](\l) | server login response message stream can      |
|                      | optionally include a FEATUREEXTACK token by   |
|                      | including the GLOBALTRANSACTIONS FeatureId in |
|                      | the FEATUREEXTACK token stream. The           |
|                      | corresponding FeatureAckData MUST then        |
|                      | include a flag that indicates whether the     |
|                      | server supports [**Global                     |
|                      | Transactions*                                 |
|                      | *](#gt_57552b13-b14e-4601-9621-500ce3297d15). |
|                      | The FeatureAckData format is as follows:      |
|                      |                                               |
|                      | 1174. IsEnabled = BYTE                        |
|                      |                                               |
|                      |       FeatureAckData = IsEnabled              |
|                      |                                               |
|                      | IsEnabled: Specifies whether the server       |
|                      | supports Global Transactions. The values of   |
|                      | this field are as follows:                    |
|                      |                                               |
|                      | -   0 = The server does not support Global    |
|                      |     Transactions.                             |
|                      |                                               |
|                      | -   1 = The server supports Global            |
|                      |     Transactions.                             |
+----------------------+-----------------------------------------------+
| %0x08                | The presence of the AZURESQLSUPPORT           |
|                      | FeatureExt indicates whether failover partner |
| (AZURESQLSUPPORT)    | login with read-only intent to Azure SQL      |
|                      | Database MAY[\<60\>](\l) be supported. For    |
| (introduced in TDS   | information about failover partner, see       |
| 7.4)                 | [\[MSDOCS-DBMirror\]](htt                     |
|                      | ps://go.microsoft.com/fwlink/?linkid=874052). |
|                      |                                               |
|                      | Whenever a login response stream is sent for  |
|                      | a TDS connection whose login request includes |
|                      | an AZURESQLSUPPORT FeatureExt token, the      |
|                      | server login response message stream can      |
|                      | optionally include a FEATUREEXTACK token by   |
|                      | setting the corresponding feature switch in   |
|                      | Azure SQL Database. If it is included, the    |
|                      | FEATUREEXTACK token stream MUST include the   |
|                      | AZURESQLSUPPORT FeatureId.                    |
|                      |                                               |
|                      | 1177. FeatureAckData = BYTE                   |
|                      |                                               |
|                      | BYTE: The Bit 0 flag specifies whether        |
|                      | failover partner login with read-only intent  |
|                      | is supported. The values of this BYTE are as  |
|                      | follows:                                      |
|                      |                                               |
|                      | -   0 = The server does not support the       |
|                      |     AZURESQLSUPPORT feature extension.        |
|                      |                                               |
|                      | -   1 = The server supports the               |
|                      |     AZURESQLSUPPORT feature extension.        |
+----------------------+-----------------------------------------------+
| %0x09                | Whenever a login response stream is sent for  |
|                      | a TDS connection whose login request includes |
| (DATACLASSIFICATION) | a DATACLASSIFICATION FeatureExt token, the    |
|                      | server login response message stream          |
| (introduced in TDS   | SHOULD[\<61\>](\l) be capable of optionally   |
| 7.4)                 | containing a FEATUREEXTACK token by including |
|                      | the DATACLASSIFICATION FeatureId in the       |
|                      | FEATUREEXTACK token stream. The corresponding |
|                      | FeatureAckData MUST then include the          |
|                      | following information that indicates whether  |
|                      | the server supports [**data                   |
|                      | classification                                |
|                      | **](#gt_acdeafb0-9b24-420e-b712-9284ad49eb56) |
|                      | and to what extent. The FeatureAckData format |
|                      | is as follows:                                |
|                      |                                               |
|                      | 1178. DATACLASSIFICATION_VERSION = BYTE       |
|                      |                                               |
|                      |       IsEnabled = BYTE                        |
|                      |                                               |
|                      |       VersionSpecificData = \*2147483647BYTE  |
|                      |       ; The actual length                     |
|                      |                                               |
|                      |       ; of data is                            |
|                      |                                               |
|                      |       ; FeatureAckDataLen - 2                 |
|                      |                                               |
|                      |       FeatureAckData =                        |
|                      |       DATACLASSIFICATION_VERSION              |
|                      |                                               |
|                      |       IsEnabled                               |
|                      |                                               |
|                      |       VersionSpecificData                     |
|                      |                                               |
|                      | DATACLASSIFICATION_VERSION: This field        |
|                      | specifies the version number of the data      |
|                      | classification information that is to be used |
|                      | for this connection. This value MUST be 1 or  |
|                      | 2, as specified for                           |
|                      | DATACLASSIFICATION_VERSION in section         |
|                      | [2.2.6.4                                      |
|                      | ](#Section_773a62b6ee894c029e5e344882630aac). |
|                      |                                               |
|                      | IsEnabled: This field specifies whether the   |
|                      | server supports data classification. The      |
|                      | values of this field are as follows:          |
|                      |                                               |
|                      | -   0 = The server does not support data      |
|                      |     classification.                           |
|                      |                                               |
|                      | -   1 = The server supports data              |
|                      |     classification.                           |
|                      |                                               |
|                      | VersionSpecificData: This field specifies     |
|                      | which version of data classification          |
|                      | information is returned. The values of this   |
|                      | field are as follows:                         |
|                      |                                               |
|                      | When the value of the                         |
|                      | DATACLASSIFICATION_VERSION field is 1 or 2,   |
|                      | the response in the feature extension         |
|                      | acknowledgement contains no version-specific  |
|                      | data.                                         |
+----------------------+-----------------------------------------------+
| %0x0A                | The presence of the UTF8_SUPPORT              |
|                      | FeatureExtAck token in the response message   |
| (UTF8_SUPPORT)       | stream indicates whether the server's ability |
|                      | to receive and send UTF-8 encoded data        |
| (introduced in TDS   | SHOULD[\<62\>](\l) be supported.              |
| 7.4)                 |                                               |
|                      | Whenever a login response stream is sent for  |
|                      | a TDS connection whose login request includes |
|                      | a UTF8_SUPPORT FeatureExt token, the server   |
|                      | login response message stream can optionally  |
|                      | include a FEATUREEXTACK token. If that token  |
|                      | is included, the FEATUREEXTACK token MUST     |
|                      | include the UTF8_SUPPORT FeatureId and the    |
|                      | appropriate feature acknowledgement data. The |
|                      | FeatureAckData format is as follows:          |
|                      |                                               |
|                      | 1187. FeatureAckData = BYTE                   |
|                      |                                               |
|                      | BYTE: The Bit 0 value specifies whether the   |
|                      | server can receive and send UTF-8 encoded     |
|                      | data. The values of this BYTE are as follows: |
|                      |                                               |
|                      | -   0 = The server does not support the       |
|                      |     UFT8_SUPPORT feature extension.           |
|                      |                                               |
|                      | -   1 = The server supports the UTF8_SUPPORT  |
|                      |     feature extension.                        |
+----------------------+-----------------------------------------------+
| %0x0B                | Whenever a login response stream is sent for  |
|                      | a TDS connection that has a login request     |
| (AZURESQLDNSCACHING) | that includes an AZURESQLDNSCACHING           |
|                      | FeatureExt token, the server login response   |
| (introduced in TDS   | message can optionally include this           |
| 7.4)                 | FeatureExtAck token. The contents of the      |
|                      | token are as follows:                         |
|                      |                                               |
|                      | 1188. IsSupported = BYTE                      |
|                      |                                               |
|                      |       FeatureAckData = IsSupported            |
|                      |                                               |
|                      | IsSupported: The Bit 0 specifies whether the  |
|                      | server supports client DNS caching. The       |
|                      | values of this BIT are as follows:            |
|                      |                                               |
|                      | -   0 = The server does not support client    |
|                      |     DNS caching.                              |
|                      |                                               |
|                      | -   1 = The server supports client DNS        |
|                      |     caching.                                  |
|                      |                                               |
|                      | A server response with IsSupported set to 1   |
|                      | indicates to the client that it is safe to    |
|                      | cache the entry. When the server responds     |
|                      | with IsSupported set to 0, the client SHOULD  |
|                      | NOT[\<63\>](\l) cache the entry.              |
+----------------------+-----------------------------------------------+
| %0x0D (JSONSUPPORT)  | Whenever a login response stream is sent for  |
| (introduced in TDS   | a TDS connection whose login request includes |
| 7.4)                 | a JSONSUPPORT FeatureExt token, the server    |
|                      | login response message stream                 |
|                      | SHOULD[\<64\>](\l) be capable of optionally   |
|                      | containing a FEATUREEXTACK token by including |
|                      | the JSON_SUPPORT FeatureId in the             |
|                      | FEATUREEXTACK token stream. The corresponding |
|                      | FeatureAckData MUST then include the          |
|                      | following information that indicates whether  |
|                      | the server supports the json datatype. The    |
|                      | FeatureAckData format is as follows:          |
|                      |                                               |
|                      | 1190. JSONSUPPORT_VERSION = BYTE              |
|                      |                                               |
|                      |       FeatureAckData = JSONSUPPORT_VERSION    |
|                      |                                               |
|                      | JSONSUPPORT_VERSION: This field specifies the |
|                      | version number of the json datatype that is   |
|                      | to be used for this connection. This value is |
|                      | 1.                                            |
+----------------------+-----------------------------------------------+
| %0x0E                | Whenever a login response stream is sent for  |
| (VECTORSUPPORT)      | a TDS connection whose login request includes |
| (introduced in TDS   | a VECTORSUPPORT FeatureExt token, the server  |
| 7.4)                 | login response message stream                 |
|                      | SHOULD[\<65\>](\l) be capable of optionally   |
|                      | containing a FEATUREEXTACK token by including |
|                      | the VECTOR_SUPPORT FeatureId in the           |
|                      | FEATUREEXTACK token stream. The corresponding |
|                      | FeatureAckData MUST then include the          |
|                      | following information that indicates whether  |
|                      | the server supports the vector datatype. The  |
|                      | FeatureAckData format is as follows:          |
|                      |                                               |
|                      | 1192. VECTORSUPPORT_VERSION = BYTE            |
|                      |                                               |
|                      |       FeatureAckData = VECTORSUPPORT_VERSION  |
|                      |                                               |
|                      | VECTORSUPPORT_VERSION: This field specifies   |
|                      | the version number of the vector datatype     |
|                      | that is to be used for this connection. This  |
|                      | value is 1.                                   |
+----------------------+-----------------------------------------------+
| %0x0F                | Whenever a login response stream is sent for  |
| (ENH                 | a TDS connection whose login request includes |
| ANCEDROUTINGSUPPORT) | an ENHANCEDROUTINGSUPPORT FeatureExt token,   |
| (introduced in TDS   | the server login response message stream can  |
| 7.4)                 | optionally[\<66\>](\l) include a              |
|                      | FEATUREEXTACK token by including the          |
|                      | ENHANCEDROUTINGSUPPORT FeatureId in the       |
|                      | FEATUREEXTACK token stream. The corresponding |
|                      | FeatureAckData MUST then include a flag that  |
|                      | indicates whether the server supports         |
|                      | Enhanced Routing. The FeatureAckData format   |
|                      | is as follows:                                |
|                      |                                               |
|                      | 1194. IsEnabled=BYTE                          |
|                      |                                               |
|                      |       FeatureAckData=IsEnabled                |
|                      |                                               |
|                      | IsEnabled: Specifies whether the server       |
|                      | supports Enhanced Routing. The values of this |
|                      | field are as follows:                         |
|                      |                                               |
|                      | -   0 = The server does not support Enhanced  |
|                      |     Routing.                                  |
|                      |                                               |
|                      | -   1 = The server supports Enhanced Routing. |
|                      |                                               |
|                      | When the value of IsEnabled is 0, the client  |
|                      | should not accept Enhanced Routing ENVCHANGE  |
|                      | tokens.                                       |
+----------------------+-----------------------------------------------+
| %xFF                 | This option signals the end of the            |
|                      | FeatureExtAck feature and MUST be the         |
| (TERMINATOR)         | feature\'s last option.                       |
+----------------------+-----------------------------------------------+

#### FEDAUTHINFO

**Token Stream Name:**

1197. FEDAUTHINFO

**Token Stream Function:**

Introduced in TDS 7.4, [**federated
authentication**](#gt_5ae22a0e-5ff4-441b-80d4-224ef4dd4d19) information
is returned to the client to be used for generating a Federated
Authentication Token during the login process. This token MUST be the
only token in a Federated Authentication Information message and MUST
NOT be included in any other message type.[\<67\>](\l)

**Token Stream Comments:**

-   The token value is 0xEE.

**Token Stream-Specific Rules:**

1198. TokenType = BYTE

      TokenLength = DWORD ; (introduced in TDS 7.4)

      CountOfInfoIDs = DWORD ; (introduced in TDS 7.4)

      FedAuthInfoID = BYTE ; (introduced in TDS 7.4)

      FedAuthInfoDataLen = DWORD ; (introduced in TDS 7.4)

      FedAuthInfoDataOffset = DWORD ; (introduced in TDS 7.4)

      FedAuthInfoData = VARBYTES ; (introduced in TDS 7.4)

      FedAuthInfoOpt = (FedAuthInfoID ; (introduced in TDS 7.4)

      FedAuthInfoDataLen

      FedAuthInfoDataOffset)

**Token Stream Definition:**

1212. FEDAUTHINFO = TokenType ; (introduced in TDS 7.4)

      TokenLength

      CountOfInfoIDs

      1\*FedAuthInfoOpt

      FedAuthInfoData

**Token Stream Parameter Details**

  -------------------------------------------------------------------------------
  **Parameter**           **Description**
  ----------------------- -------------------------------------------------------
  TokenType               FEDAUTHINFO_TOKEN

  TokenLength             The length of the whole Federated Authentication
                          Information token, not including the size occupied by
                          TokenLength itself. The minimum value for this field is
                          sizeof(DWORD) because the field CountOfInfoIDs MUST be
                          present even if no federated authentication information
                          is sent as part of the token.

  CountOfInfoIDs          The number of federated authentication information
                          options that are sent in the token. If no
                          FedAuthInfoOpt is sent in the token, this field MUST be
                          present and set to 0.

  FedAuthInfoID           The unique identifier number for the type of
                          information.

  FedAuthInfoDataLen      The length of FedAuthInfoData, in bytes.

  FedAuthInfoDataOffset   The offset at which the federated authentication
                          information data for FedAuthInfoID is present, measured
                          from the address of CountOfInfoIDs.

  FedAuthInfoData         The actual information data as binary, with the length
                          in bytes equal to FedAuthInfoDataLen.
  -------------------------------------------------------------------------------

The following table describes the FedAuthInfo feature option and
description.

+----------+-----------------------------------------------------------+
| *        | **FedAuthInfoData Description**                           |
| *FedAuth |                                                           |
| InfoID** |                                                           |
+==========+===========================================================+
| %0x00    | Reserved.                                                 |
+----------+-----------------------------------------------------------+
| %0x01    | A [**Unicode**](#gt_c305d0ab-8b94-461a-bd76-13b40cb8c4d8) |
|          | string that represents the token endpoint URL from which  |
| (STSURL) | to acquire a Federated Authentication Token.              |
+----------+-----------------------------------------------------------+
| %0x02    | A Unicode string that represents the Service Principal    |
|          | Name (SPN) to use for acquiring a Federated               |
| (SPN)    | Authentication Token. SPN is a string that represents the |
|          | resource in a directory.                                  |
+----------+-----------------------------------------------------------+

#### INFO

**Token Stream Name:**

1217. INFO

**Token Stream Function:**

Used to send an information message to the client.

**Token Stream Comments**

-   The token value is 0xAB.

**Token Stream-Specific Rules:**

1218. TokenType = BYTE

      Length = USHORT

      Number = LONG

      State = BYTE

      Class = BYTE

      MsgText = US_VARCHAR

      ServerName = B_VARCHAR

      ProcName = B_VARCHAR

      LineNumber = USHORT / LONG; (Changed to LONG in TDS 7.2)

The type of the **LineNumber** element depends on the version of TDS.

**Token Stream Definition:**

1227. INFO = TokenType

      Length

      Number

      State

      Class

      MsgText

      ServerName

      ProcName

      LineNumber

**Token Stream Parameter Details**

  --------------------------------------------------------------------------
  Parameter    Description
  ------------ -------------------------------------------------------------
  TokenType    INFO_TOKEN

  Length       The total length of the INFO [**data
               stream**](#gt_151643ce-fb5e-460e-8bdf-dc10bbd1950e), in
               bytes.

  Number       The info number.[\<68\>](\l)

  State        The error state, used as a modifier to the info Number.

  Class        The class (severity) of the error. A class of less than 10
               indicates an informational message.

  MsgText      The message text length and message text using US_VARCHAR
               format.

  ServerName   The server name length and server name using B_VARCHAR
               format.

  ProcName     The [**stored
               procedure**](#gt_324d32b3-f4f3-41c9-b695-78c498094fb7) name
               length and stored procedure name using B_VARCHAR format.

  LineNumber   The line number in the SQL batch or stored procedure that
               caused the error. Line numbers begin at 1; therefore, if the
               line number is not applicable to the message as determined by
               the upper layer, the value of LineNumber is 0.
  --------------------------------------------------------------------------

#### LOGINACK

**Token Stream Name:**

1236. LOGINACK

**Token Stream Function:**

Used to send a response to a login request (LOGIN7) to the client.

**Token Stream Comments**

-   The token value is 0xAD.

-   If a LOGINACK is not received by the client as part of the login
    procedure, the login to the server is unsuccessful.

**Token Stream-Specific Rules:**

1237. TokenType = BYTE

      Length = USHORT

      Interface = BYTE

      TDSVersion = DWORD

      ProgName = B_VARCHAR

      MajorVer = BYTE

      MinorVer = BYTE

      BuildNumHi = BYTE

      BuildNumLow = BYTE

      ProgVersion = MajorVer

      MinorVer

      BuildNumHi

      BuildNumLow

**Token Stream Definition:**

1252. LOGINACK = TokenType

      Length

      Interface

      TDSVersion

      ProgName

      ProgVersion

**Token Stream Parameter Details**

+---------+------------------------------------------------------------+
| Pa      | Description                                                |
| rameter |                                                            |
+=========+============================================================+
| To      | LOGINACK_TOKEN                                             |
| kenType |                                                            |
+---------+------------------------------------------------------------+
| Length  | The total length, in bytes, of the following fields:       |
|         | Interface, TDSVersion, ProgName, and ProgVersion.          |
+---------+------------------------------------------------------------+
| In      | The type of                                                |
| terface | [**interface**](#gt_95913fbd-3262-47ae-b5eb-18e6806824b9)  |
|         | with which the server accepts client requests:             |
|         |                                                            |
|         | 0: SQL_DFLT (server confirms that whatever is sent by the  |
|         | client is acceptable. If the client requested SQL_DFLT,    |
|         | SQL_TSQL is used).                                         |
|         |                                                            |
|         | 1: SQL_TSQL (TSQL is accepted).                            |
+---------+------------------------------------------------------------+
| TDS     | The TDS version being used by the server.[\<69\>](\l)      |
| Version |                                                            |
+---------+------------------------------------------------------------+
| P       | The name of the server.                                    |
| rogName |                                                            |
+---------+------------------------------------------------------------+
| M       | The major version number (0-255).                          |
| ajorVer |                                                            |
+---------+------------------------------------------------------------+
| M       | The minor version number (0-255).                          |
| inorVer |                                                            |
+---------+------------------------------------------------------------+
| Bui     | The high byte of the build number (0-255).                 |
| ldNumHi |                                                            |
+---------+------------------------------------------------------------+
| Buil    | The low byte of the build number (0-255).                  |
| dNumLow |                                                            |
+---------+------------------------------------------------------------+

#### NBCROW

**Token Stream Name:**

1258. NBCROW

**Token Stream Function:**

NBCROW, introduced in TDS 7.3.B, is used to send a row as defined by the
COLMETADATA token (section
[2.2.7.4](#Section_58880b9f381c43b2bf8b0727a98c4f4c)) to the client with
null bitmap compression. Null bitmap compression is implemented by using
a single bit to specify whether the column is null or not null and also
by removing all null column values from the row. Removing the null
column values (which can be up to 8 bytes per null instance) from the
row provides the compression. The null bitmap contains one bit for each
column defined in COLMETADATA. In the null bitmap, a bit value of 1
means that the column is null and therefore not present in the row, and
a bit value of 0 means that the column is not null and is present in the
row. The null bitmap is always rounded up to the nearest multiple of 8
bits, so there might be 1 to 7 leftover reserved bits at the end of the
null bitmap in the last byte of the null bitmap. NBCROW is only used by
TDS [**result set**](#gt_c8a27238-8ccc-442b-9604-75f74d3e6b3d) streams
from server to client. NBCROW MUST NOT be used in BulkLoadBCP streams.
NBCROW MUST NOT be used in TVP row streams.

**Token Stream Comments**

-   The token value is 0xD2/210.

**Token Stream-Specific Rules:**

1259. TokenType = BYTE

      TextPointer = B_VARBYTE

      Timestamp = 8BYTE

      Data = TYPE_VARBYTE

      NullBitmap = \<NullBitmapByteCount\> BYTE ; see note on
      NullBitmapByteCount

      ColumnData = \[TextPointer Timestamp\] Data

      AllColumnData = \*ColumnData

ColumnData is repeated once for each non-null column of data.

NullBitmapBitCount is equal to the number of columns in COLMETADATA.

NullBitmapByteCount is equal to the smallest number of bytes needed to
hold \'NullBitmapBitCount\' bits.

The server can decide to send either a NBCROW token or a ROW token. For
example, the server might choose to send a ROW token if there is no byte
savings if the result set has no [**nullable
columns**](#gt_16dd540d-3913-48e5-9a93-a769e85570d0), or if a particular
row in a result set has no null values. This implies that NBCROW and ROW
tokens can be intermixed in the same result set.

When determining whether or not a specific column is null, consider all
the columns from left to right ordered using a zero-based index from 0
to 65534 as they occur in the ColumnData section of the COLMETADATA
token. The null bitmap indicates that a column is null using a zero bit
at the following byte and bit layout:

1266. Byte 1 Byte 2 Byte 3

      \-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\--
      \-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\--
      \-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\--

      07 06 05 04 03 02 01 00 15 14 13 12 11 10 09 08 23 22 21 20 19 18
      17 16

Hence the first byte contains flags for columns 0 through 7, with the
least significant (or rightmost) bit within the byte indicating the
zeroth column and the most significant (or leftmost) bit within the byte
indicating the seventh column. For example, column index 8 would be in
the second byte as the least significant bit. If the null bitmap bit is
set, the column is null and no null token value for the column follows
in the row. If the null bitmap bit is clear, the column is not null and
the value for the column follows in the row.

**Token Stream Definition:**

1269. NBCROW = TokenType

      NullBitmap

      AllColumnData

**Token Stream Parameter Details**

  ----------------------------------------------------------------------------
  Parameter     Description
  ------------- --------------------------------------------------------------
  TokenType     NBCROW_TOKEN (0xD2)

  TextPointer   The length of the text pointer and the text pointer for Data.

  Timestamp     The time stamp of a text/image column.

  Data          The actual data for the column. The TYPE_INFO information
                describing the data type of this data is given in the
                preceding COLMETADATA_TOKEN.
  ----------------------------------------------------------------------------

#### OFFSET

**Token Stream Name:**

1272. OFFSET

**Token Stream Function:**

Used to inform the client where in the client\'s SQL text buffer a
particular keyword occurs.

**Token Stream Comments:**

-   The token value is 0x78.

-   The token was removed in TDS 7.2.

**Token Stream-Specific Rules:**

1273. TokenType = BYTE

      Identifier = USHORT

      OffSetLen = USHORT

**Token Stream Definition:**

1276. OFFSET = TokenType ; (removed in TDS 7.2)

      Identifier

      OffSetLen

**Token Stream Parameter Details**

  ---------------------------------------------------------------------------
  Parameter    Description
  ------------ --------------------------------------------------------------
  TokenType    OFFSET_TOKEN

  Identifier   The keyword to which OffSetLen refers.

  OffsetLen    The offset in the SQL text buffer received by the server of
               the identifier. The SQL text buffer begins with an OffSetLen
               value of 0 (MOD 64 kilobytes if value of OffSet is larger than
               64 kilobytes).
  ---------------------------------------------------------------------------

#### ORDER

**Token Stream Name:**

1279. ORDER

**Token Stream Function:**

Used to inform the client by which columns the data is ordered.

**Token Stream Comments**

-   The token value is 0xA9.

-   This token is sent only in the event that an ORDER BY clause is
    executed.

**Token Stream-Specific Rules:**

1280. TokenType = BYTE

      Length = USHORT

      ColNum = \*USHORT

The **ColNum** element is repeated once for each column within the ORDER
BY clause.

**Token Stream Definition:**

1283. ORDER = TokenType

      Length

      ColNum

**Token Stream Parameter Details**

  -----------------------------------------------------------------------
  Parameter         Description
  ----------------- -----------------------------------------------------
  TokenType         ORDER_TOKEN

  Length            The total length of the ORDER data stream.

  ColNum            The column number in the [**result
                    set**](#gt_c8a27238-8ccc-442b-9604-75f74d3e6b3d).
  -----------------------------------------------------------------------

#### RETURNSTATUS

**Token Stream Name:**

1286. RETURNSTATUS

**Token Stream Function:**

Used to send the status value of an RPC (section
[2.2.1.6](#Section_26327437aa3c4e969bba73a6e862ba21)) to the client. The
server also uses this token to send the result status value of a T-SQL
EXEC query.

**Token Stream Comments:**

-   The token value is 0x79.

-   This token MUST be returned to the client when an RPC is executed by
    the server.

**Token Stream-Specific Rules:**

1287. TokenType = BYTE

      Value = LONG

**Token Stream Definition:**

1289. RETURNSTATUS = TokenType

      Value

**Token Stream Parameter Details**

  -------------------------------------------------------------------------
  Parameter   Description
  ----------- -------------------------------------------------------------
  TokenType   RETURNSTATUS_TOKEN

  Value       The return status value determined by the remote procedure.
              Return status MUST NOT be NULL.
  -------------------------------------------------------------------------

#### RETURNVALUE

**Token Stream Name:**

1291. RETURNVALUE

**Token Stream Function:**

Used to send the return value of an RPC to the client. When an RPC is
executed, the associated parameters might be defined as input or output
(or \"return\") parameters. This token is used to send a description of
the return parameter to the client. This token is also used to describe
the value returned by a UDF when executed as an RPC.

**Token Stream Comments:**

-   The token value is 0xAC.

-   Multiple return values can exist per RPC. There is a separate
    RETURNVALUE token sent for each parameter returned.

-   Large Object output parameters are reordered to appear at the end of
    the stream. First the group of small parameters is sent, followed by
    the group of large output parameters. There is no reordering within
    the groups.

-   A UDF cannot have return parameters. As such, if a UDF is executed
    as an RPC there is exactly one RETURNVALUE token sent to the client.

**Token Stream-Specific Rules:**

1292. TokenType = BYTE

      ParamName = B_VARCHAR

      ParamOrdinal = USHORT

      Status = BYTE

      UserType = USHORT/ULONG; (Changed to ULONG in TDS 7.2)

      fNullable = BIT

      fCaseSen = BIT

      usUpdateable = 2BIT ; 0 = ReadOnly

      ; 1 = Read/Write

      ; 2 = Unused

      fIdentity = BIT

      fComputed = BIT ; (introduced in TDS 7.2)

      usReservedODBC = 2BIT

      fFixedLenCLRType = BIT ; (introduced in TDS 7.2)

      usReserved = 7BIT

      usReserved2 = 2BIT

      fEncrypted = BIT ; (introduced in TDS 7.4)

      usReserved3 = 4BIT

      Flags = fNullable

      fCaseSen

      usUpdateable

      fIdentity

      (FRESERVEDBIT / fComputed)

      usReservedODBC

      (FRESERVEDBIT / fFixedLenCLRType)

      (usReserved / (usReserved2 fEncrypted usReserved3))

      ; (introduced in TDS 7.4)

      TypeInfo = TYPE_INFO

      Value = TYPE_VARBYTE

      BaseTypeInfo = TYPE_INFO ; (BaseTypeInfo introduced in TDS 7.4)

      EncryptionAlgo = BYTE ; (EncryptionAlgo introduced in TDS 7.4)

      AlgoName = B_VARCHAR ; (introduced in TDS 7.4)

      EncryptionAlgoType = BYTE ; (introduced in TDS 7.4)

      NormVersion = BYTE ; (introduced in TDS 7.4)

      CryptoMetaData = UserType ; (CryptoMetaData introduced in TDS 7.4)

      BaseTypeInfo

      EncryptionAlgo

      \[AlgoName\]

      EncryptionAlgoType

      NormVersion

**Token Stream Definition:**

1343. RETURNVALUE = TokenType

      ParamOrdinal

      ParamName

      Status

      UserType

      Flags

      TypeInfo

      CryptoMetadata

      Value

**Token Stream Parameter Details:**

+------------+---------------------------------------------------------+
| Parameter  | Description                                             |
+============+=========================================================+
| TokenType  | RETURNVALUE_TOKEN                                       |
+------------+---------------------------------------------------------+
| Pa         | Indicates the ordinal position of the output parameter  |
| ramOrdinal | in the original RPC call. Large Object output           |
|            | parameters are reordered to appear at the end of the    |
|            | stream. First the group of small parameters is sent,    |
|            | followed by the group of large output parameters. There |
|            | is no reordering within the groups.                     |
+------------+---------------------------------------------------------+
| ParamName  | The parameter name length and parameter name (within    |
|            | B_VARCHAR).                                             |
+------------+---------------------------------------------------------+
| Status     | 0x01: If ReturnValue corresponds to OUTPUT parameter of |
|            | a [**stored                                             |
|            | procedure**](#gt_324d32b3-f4f3-41c9-b695-78c498094fb7)  |
|            | invocation.                                             |
|            |                                                         |
|            | 0x02: If ReturnValue corresponds to return value of     |
|            | User Defined Function.                                  |
+------------+---------------------------------------------------------+
| UserType   | The user type ID of the data type of the column.        |
|            | Depending on the TDS version that is used, valid values |
|            | are 0x0000 or 0x00000000, with the exceptions of data   |
|            | type timestamp (0x0050 or 0x00000050) and alias types   |
|            | (greater than 0x00FF or 0x000000FF).                    |
+------------+---------------------------------------------------------+
| Flags      | These bit flags are described in [least significant bit |
|            | order](#Section_bbc22f15e1a04338a169c79819d39b1c). All  |
|            | of these bit flags SHOULD be set to zero. For a         |
|            | description of each bit flag, see section               |
|            | [2.2.7.4](#Section_58880b9f381c43b2bf8b0727a98c4f4c).   |
|            |                                                         |
|            | -   fNullable                                           |
|            |                                                         |
|            | -   fCaseSen                                            |
|            |                                                         |
|            | -   usUpdateable                                        |
|            |                                                         |
|            | -   fIdentity                                           |
|            |                                                         |
|            | -   fComputed                                           |
|            |                                                         |
|            | -   usReservedODBC                                      |
|            |                                                         |
|            | -   fFixedLengthCLRType                                 |
|            |                                                         |
|            | -   fEncrypted                                          |
+------------+---------------------------------------------------------+
| TypeInfo   | The TYPE_INFO for the message.                          |
+------------+---------------------------------------------------------+
| Ba         | TYPE_INFO for the unencrypted type.                     |
| seTypeInfo |                                                         |
+------------+---------------------------------------------------------+
| Encr       | A byte that describes the encryption algorithm that is  |
| yptionAlgo | used. AlgoName is populated with the name of the custom |
|            | encryption algorithm. For all EncryptionAlgo values     |
|            | other than 0, AlgoName MUST NOT be sent. If             |
|            | EncryptionAlgo is set to 1, the algorithm that is used  |
|            | is AEAD_AES_256_CBC_HMAC_SHA512, as described in        |
|            | [\[IETF-Auth                                            |
|            | Encr\]](https://go.microsoft.com/fwlink/?LinkId=524322) |
|            | section 5.4.                                            |
+------------+---------------------------------------------------------+
| AlgoName   | Algorithm name literal that is used to encrypt the      |
|            | plaintext value.                                        |
+------------+---------------------------------------------------------+
| Encrypti   | A field that describes the encryption algorithm type.   |
| onAlgoType | Available values are defined as follows:                |
|            |                                                         |
|            | 1 = Deterministic encryption.                           |
|            |                                                         |
|            | 2 = Randomized encryption.                              |
+------------+---------------------------------------------------------+
| N          | The normalization version to which plaintext data MUST  |
| ormVersion | be normalized. Version numbering starts at 0x01.        |
+------------+---------------------------------------------------------+
| Cryp       | This describes the encryption metadata for a column. It |
| toMetaData | contains the UserType, the TYPE_INFO (BaseTypeInfo) for |
|            | the plaintext value, the encryption algorithm that is   |
|            | used, the algorithm name literal, the encryption        |
|            | algorithm type, and the normalization version.          |
+------------+---------------------------------------------------------+
| Value      | The type-dependent data for the parameter (within       |
|            | TYPE_VARBYTE).                                          |
+------------+---------------------------------------------------------+

#### ROW

**Token Stream Name:**

1352. ROW

**Token Stream Function:**

Used to send a complete row, as defined by the COLMETADATA token
(section [2.2.7.4](#Section_58880b9f381c43b2bf8b0727a98c4f4c)), to the
client.

**Token Stream Comments:**

-   The token value is 0xD1.

**Token Stream-Specific Rules:**

1353. TokenType = BYTE

      TextPointer = B_VARBYTE

      Timestamp = 8BYTE

      Data = TYPE_VARBYTE

      ColumnData = \[TextPointer Timestamp\]

      Data

      AllColumnData = \*ColumnData

The **ColumnData** element is repeated once for each column of data.

TextPointer and Timestamp MUST NOT be specified if the instance of type
text/ntext/image is a NULL instance (GEN_NULL).

**Token Stream Definition:**

1363. ROW = TokenType

      AllColumnData

**Token Stream Parameter Details:**

  ----------------------------------------------------------------------------
  Parameter     Description
  ------------- --------------------------------------------------------------
  TokenType     ROW_TOKEN

  TextPointer   The length of the text pointer and the text pointer for data.

  Timestamp     The time stamp of a text/image column. This is not present if
                the value of data is CHARBIN_NULL or GEN_NULL.

  Data          The actual data for the column. The TYPE_INFO information
                describing the data type of this data is given in the
                preceding COLMETADATA_TOKEN, ALTMETADATA_TOKEN or
                OFFSET_TOKEN.
  ----------------------------------------------------------------------------

#### SESSIONSTATE

**Token Stream Name:**

1365. SESSIONSTATE

**Token Stream Function:**

Used to send session state data to the client. The data format defined
here can also be used to send session state data for session recovery
during login and login response.

**Token Stream Comments:**

-   The token value is 0xE4.

-   This token stream MUST NOT be sent if the SESSIONRECOVERY feature is
    not negotiated on the connection.

-   When this token stream is sent, the next token MUST be DONE (section
    [2.2.7.6](#Section_3c06f11098bd4d5bb836b1ba66452cb7)) or DONEPROC
    (section [2.2.7.8](#Section_65e24140edea46e5b710209af2016195)) with
    DONE_FINAL.

-   If the SESSIONRECOVERY feature is negotiated on the connection, the
    server SHOULD send this token to the client to inform any session
    state update.

**Token Stream-Specific Rules:**

1366. fRecoverable = BIT

      TokenType = BYTE

      Length = DWORD

      SeqNo = DWORD

      Status = fRecoverable 7FRESERVEDBIT

      StateId = BYTE

      StateLen = BYTE ; 0-%xFE

      /

      (%xFF DWORD) ; %xFF - %xFFFF

      SessionStateData = StateId

      StateLen

      StateValue

      SessionStateDataSet = 1\*SessionStateData

**Token Stream Definition:**

1384. SESSIONSTATE = TokenType

      Length

      SeqNo

      Status

      SessionStateDataSet

**Token Stream Parameter Details**

+-------+--------------------------------------------------------------+
| Para  | Description                                                  |
| meter |                                                              |
+=======+==============================================================+
| Toke  | SESSIONSTATE_TOKEN                                           |
| nType |                                                              |
+-------+--------------------------------------------------------------+
| L     | The length, in bytes, of the token stream (excluding         |
| ength | TokenType and Length).                                       |
+-------+--------------------------------------------------------------+
| SeqNo | The sequence number of the SESSIONSTATE token in the         |
|       | connection. This number, which starts at 0 and increases by  |
|       | one each time, can be used to track the order of             |
|       | SESSIONSTATE tokens sent during the course of a connection.  |
|       | The SeqNo applies to all StateIds in the token. If the SeqNo |
|       | for any StateId reaches %xFFFFFFFF, both client and server   |
|       | MUST consider that the SESSIONRECOVERY feature is            |
|       | permanently disabled on the connection. The server SHOULD    |
|       | send a token with fRecoverable set to FALSE to disable       |
|       | SESSIONRECOVERY for this session. The client SHOULD NOT set  |
|       | either ResetConn bit (RESETCONNECTION or                     |
|       | RESETCONNECTIONSKIPTRAN) on the connection once it receives  |
|       | any SeqNo of %xFFFFFFFF because ResetConn could reset a      |
|       | connection back to an initial recoverable state and          |
|       | SESSIONRECOVERY needs to be permanently disabled on the      |
|       | connection in this case. If the server does receive          |
|       | ResetConn after SeqNo reaches %xFFFFFFFF, it SHOULD reuse    |
|       | this same SeqNo to disable SESSIONRECOVERY.                  |
|       |                                                              |
|       | The client SHOULD track SeqNo for each StateId and keep the  |
|       | latest data for session recovery.                            |
+-------+--------------------------------------------------------------+
| S     | Status of the session StateId in this token.                 |
| tatus |                                                              |
|       | fRecoverable: TRUE means all session StateIds in this token  |
|       | are recoverable.                                             |
|       |                                                              |
|       | The client SHOULD track Status for each StateId and keep the |
|       | latest data for session recovery. A client MUST NOT try to   |
|       | recover a dead connection unless fRecoverable is TRUE for    |
|       | all session StateIds received from server.                   |
+-------+--------------------------------------------------------------+
| St    | The identification number of the session state. %xFF is      |
| ateId | reserved.                                                    |
+-------+--------------------------------------------------------------+
| Sta   | The length, in bytes, of the corresponding StateValue. If    |
| teLen | the length is 254 bytes or smaller, one BYTE is used to      |
|       | represent the field. If the length is 255 bytes or larger,   |
|       | %xFF followed by a DWORD is used to represent the field. If  |
|       | this field is 0, client SHOULD skip sending SessionStateData |
|       | for the StateId during session recovery.                     |
+-------+--------------------------------------------------------------+
| State | The value of the session state. This can be any arbitrary    |
| Value | data as long as the server understands it.                   |
+-------+--------------------------------------------------------------+

#### SSPI

**Token Stream Name:**

1389. SSPI

**Token Stream Function:**

The SSPI token returned during the login process.

**Token Stream Comments:**

-   The token value is 0xED.

**Token Stream-Specific Rules:**

1390. TokenType = BYTE

      SSPIBuffer = US_VARBYTE

**Token Stream Definition:**

1392. SSPI = TokenType

      SSPIBuffer

**Token Stream Parameter Details:**

  ------------------------------------------------------------------------
  Parameter    Description
  ------------ -----------------------------------------------------------
  TokenType    SSPI_TOKEN

  SSPIBuffer   The length of the SSPIBuffer and the SSPI buffer using
               B_VARBYTE format.
  ------------------------------------------------------------------------

#### TABNAME

**Token Stream Name:**

1394. TABNAME

**Token Stream Function:**

Used to send the table name to the client only when in browser mode or
from sp_cursoropen.

**Token Stream Comments:**

-   The token value is 0xA4.

**Token Stream-Specific Rules:**

1395. TokenType = BYTE

      Length = USHORT

      NumParts = BYTE ; (introduced in TDS 7.1 Revision 1)

      PartName = US_VARCHAR ; (introduced in TDS 7.1 Revision 1)

      TableName = US_VARCHAR ; (removed in TDS 7.1 Revision 1)

      /

      (NumParts

      1\*PartName) ; (introduced in TDS 7.1 Revision 1)

      AllTableNames = TableName

The **TableName** element is repeated once for each table name in the
query.

**Token Stream Definition:**

1407. TABNAME = TokenType

      Length

      AllTableNames

**Token Stream Parameter Details**

  -------------------------------------------------------------------------
  Parameter   Description
  ----------- -------------------------------------------------------------
  TokenType   TABNAME_TOKEN

  Length      The actual data length, in bytes, of the TABNAME token
              stream. The length does not include token type and length
              field.

  TableName   The name of the base table referenced in the query statement.
  -------------------------------------------------------------------------

#### TVP_ROW

**Token Stream Name:**

1410. TVP_ROW

**Token Stream Function:**

Used to send a complete table valued parameter (TVP) row, as defined by
the TVP_COLMETADATA token from client to server.

**Token Stream Comments:**

-   The token value is 0x01/1.

**Token Stream-Specific Rules:**

1411. TokenType = BYTE

      TvpColumnData = TYPE_VARBYTE

      AllColumnData = \*TvpColumnData

TvpColumnData is repeated once for each column of data with a few
exceptions. For details about when certain TvpColumnData items are
required to be omitted, see the Flags description of the TVP_COLMETADATA
definition (section
[2.2.5.5.5.1](#Section_0dfc5367a3884c929ba44d28e775acbc)).

Note that unlike the ROW token, TVP_ROW does not use TextPointer +
Timestamp prefix with TEXT, NTEXT and IMAGE types.

**Token Stream Definition:**

1414. TVP_ROW = TokenType

      AllColumnData

**Token Stream Parameter Details:**

  ---------------------------------------------------------------------------
  Parameter       Description
  --------------- -----------------------------------------------------------
  TokenType       TVP_ROW_TOKEN

  TvpColumnData   The actual data for the TVP column. The TYPE_INFO
                  information describing the data type of this data is given
                  in the preceding TVP_COLMETADATA token.
  ---------------------------------------------------------------------------

