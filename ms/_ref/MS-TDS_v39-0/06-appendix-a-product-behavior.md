# Appendix A: Product Behavior

The information in this specification is applicable to the following
Microsoft products or supplemental software. References to product
versions include updates to those products.

This document specifies version-specific details in the Microsoft .NET
Framework. For information about which versions of .NET Framework are
available in each released Windows product or as supplemental software,
see
[\[MS-NETOD\]](%5bMS-NETOD%5d.pdf#Section_bcca8164da0843f2a983c34ed99171b0)
section 4.

-   Microsoft .NET Framework 1.1

-   Microsoft .NET Framework 2.0

-   Microsoft .NET Framework 4.0

-   Microsoft .NET Framework 4.5

-   Microsoft .NET Framework 4.6

-   Microsoft .NET Framework 4.7

-   Microsoft .NET Framework 4.8

```{=html}
<!-- -->
```
-   Microsoft SQL Server 7.0

-   Microsoft SQL Server 2000

-   Microsoft SQL Server 2005

-   Microsoft SQL Server 2008

-   Microsoft SQL Server 2008 R2

-   Microsoft SQL Server 2012

-   Microsoft SQL Server 2014

-   Microsoft SQL Server 2016

-   Microsoft SQL Server 2017

-   Microsoft SQL Server 2019

-   Microsoft SQL Server 2022

-   Microsoft SQL Server 2025

Exceptions, if any, are noted in this section. If an update version,
service pack or Knowledge Base (KB) number appears with a product name,
the behavior changed in that update. The new behavior also applies to
subsequent updates unless otherwise specified. If a product edition
appears with the product version, behavior is different in that product
edition.

Unless otherwise specified, any statement of optional behavior in this
specification that is prescribed using the terms \"SHOULD\" or \"SHOULD
NOT\" implies product behavior in accordance with the SHOULD or SHOULD
NOT prescription. Unless otherwise specified, the term \"MAY\" implies
that the product does not follow the prescription.

[\<1\> Section 1.3](\l): The following table describes the latest TDS
version that is supported by a particular version of Microsoft SQL
Server. To determine the earliest TDS version that is supported by a
particular SQL Server version, refer to the product documentation.

+-----------------------------+----------------------------------------+
| TDS version                 | SQL Server version                     |
+=============================+========================================+
| 7.0                         | SQL Server 7.0                         |
+-----------------------------+----------------------------------------+
| 7.1                         | SQL Server 2000                        |
+-----------------------------+----------------------------------------+
| 7.1 Revision 1              | SQL Server 2000 SP1                    |
+-----------------------------+----------------------------------------+
| 7.2                         | SQL Server 2005                        |
+-----------------------------+----------------------------------------+
| 7.3.A                       | SQL Server 2008                        |
+-----------------------------+----------------------------------------+
| 7.3.B                       | SQL Server 2008 R2                     |
+-----------------------------+----------------------------------------+
| 7.4                         | SQL Server 2012                        |
|                             |                                        |
|                             | SQL Server 2014                        |
|                             |                                        |
|                             | SQL Server 2016                        |
|                             |                                        |
|                             | SQL Server 2017                        |
|                             |                                        |
|                             | SQL Server 2019                        |
|                             |                                        |
|                             | SQL Server 2022                        |
|                             |                                        |
|                             | SQL Server 2025                        |
+-----------------------------+----------------------------------------+
| 8.0                         | SQL Server 2022                        |
|                             |                                        |
|                             | SQL Server 2025                        |
+-----------------------------+----------------------------------------+

The following table describes the TDS versions that are supported by
particular versions of the .NET Framework.

+------------------------------+---------------------------------------+
| TDS version                  | .NET Framework version                |
+==============================+=======================================+
| 7.0                          | .NET Framework 1.1                    |
+------------------------------+---------------------------------------+
| 7.1                          | .NET Framework 1.1                    |
+------------------------------+---------------------------------------+
| 7.1 Revision 1               | .NET Framework 1.1                    |
+------------------------------+---------------------------------------+
| 7.2                          | .NET Framework 2.0                    |
+------------------------------+---------------------------------------+
| 7.3.A                        | .NET Framework 2.0                    |
|                              |                                       |
|                              | .NET Framework 4.0                    |
+------------------------------+---------------------------------------+
| 7.3.B                        | .NET Framework 2.0                    |
|                              |                                       |
|                              | .NET Framework 4.0                    |
+------------------------------+---------------------------------------+
| 7.4                          | .NET Framework 4.5                    |
|                              |                                       |
|                              | .NET Framework 4.6                    |
|                              |                                       |
|                              | .NET Framework 4.7                    |
|                              |                                       |
|                              | .NET Framework 4.8                    |
+------------------------------+---------------------------------------+
| 8.0                          | Not applicable                        |
+------------------------------+---------------------------------------+

[\<2\> Section 2.1](\l): Microsoft Windows Named Pipes in message mode
[\[MSDN-NP\]](https://go.microsoft.com/fwlink/?LinkId=90247). Please see
[\[MSDN-NamedPipes\]](https://go.microsoft.com/fwlink/?LinkId=127839)
for additional information related to Microsoft-specific
implementations.

[\<3\> Section 2.1](\l): VIA is supported only by SQL Server 7.0, SQL
Server 2000, SQL Server 2005, SQL Server 2008, and SQL Server 2008 R2.
This means that VIA is never the underlying transport protocol if either
the server or the client can support TDS 7.4 or TDS 8.0.

[\<4\> Section 2.2.1.3](\l): Federated authentication is not supported
by SQL Server.

[\<5\> Section 2.2.3.1.1](\l): Only legacy clients that support SQL
Server versions that were released prior to SQL Server 7.0 can use
Pre-TDS7 Login.

[\<6\> Section 2.2.3.1.1](\l): Only clients that support SQL Server 7.0
or later can use TDS7 Login.

[\<7\> Section 2.2.3.1.5](\l): Depending on the message type and
provider, such as Microsoft SQL Server Native Client or Microsoft .NET
Framework Data Provider for SQL Server, PacketID values start with
either 0 or 1, which is an implementation choice. The .NET Framework
Data Provider for SQL Server uses 1.

\<8\> Section 2.2.4.3: Not all pre-SQL Server 7.0 servers support the
attention signal by using the message header. The older implementation
was for the client to send a 1-byte message (no header) containing \"A\"
by using the [**out-of-band**](#gt_26c1caf3-c889-4b99-a22b-9da056d397cf)
write.

[\<9\> Section 2.2.5.1.2](\l): The sorting styles that are used by SQL
Server are described in
[\[MSDN-ColSortSty\]](https://go.microsoft.com/fwlink/?LinkId=233328).

[\<10\> Section 2.2.5.1.2](\l): COLLATION represents a collation in SQL
Server, as described in
[\[MSDN-Collation\]](https://go.microsoft.com/fwlink/?LinkId=233327). It
can be either a SQL Server collation or a Windows collation.

Version can be of value 0, 1, 2, or 3. A value of 0 denotes collations
introduced in SQL Server 2000. A value of 1 denotes collations
introduced in SQL Server 2005. A value of 2 denotes collations
introduced in SQL Server 2008. A value of 3 denotes collations
introduced in SQL Server 2017.

The **GetLocaleInfo** Windows API can be used to retrieve information
about the locale. In particular, querying for the
LOCALE_IDEFAULTANSICODEPAGE locale information constant retrieves the
code page information for the given locale.

For either collation type, the different comparison flags map to those
defined as valid comparison flags for the **CompareString** Windows API.

However, for SQL collations with
non-[**Unicode**](#gt_c305d0ab-8b94-461a-bd76-13b40cb8c4d8) data, the
SortId is used to derive comparison information flags, such as whether,
for a given SortId, a lowercase \"a\" equals an uppercase \"A\".

\<11\> Section 2.2.5.3.1: Query notifications is not supported by SQL
Server 7.0 and SQL Server 2000.

[\<12\> Section 2.2.5.3.1](\l): SSBDeployment corresponds to the SQL
Server Service Broker deployment version.

[\<13\> Section 2.2.5.4.1](\l): NULLTYPE can be sent to SQL Server (for
example, in RPCRequest), but SQL Server never emits NULLTYPE data.

[\<14\> Section 2.2.5.5.3](\l): When a .NET Framework Data Provider for
SQL Server accesses an XML field, the returned data value is encoded in
binary XML format
[\[MS-BINXML\]](%5bMS-BINXML%5d.pdf#Section_11ab6e8d247244d1a9e6bddf000e12f6).
For other providers, the value is sent in Unicode text format.

[\<15\> Section 2.2.5.5.4](\l): Microsoft implementations return an
error if a client does send a raw collation within a sql_variant.

[\<16\> Section 2.2.6.3](\l): Federated Authentication and the FEDAUTH
token are not supported by SQL Server.

[\<17\> Section 2.2.6.4](\l): The version numbers used by clients are as
follows.

+--------------------------+-------------------------------------------+
| SQL Server version       | Version sent from client to server        |
+==========================+===========================================+
| SQL Server 7.0           | 0x00000070                                |
+--------------------------+-------------------------------------------+
| SQL Server 2000          | 0x00000071                                |
+--------------------------+-------------------------------------------+
| SQL Server 2000 SP1      | 0x01000071                                |
+--------------------------+-------------------------------------------+
| SQL Server 2005          | 0x02000972                                |
+--------------------------+-------------------------------------------+
| SQL Server 2008          | 0x03000A73                                |
+--------------------------+-------------------------------------------+
| SQL Server 2008 R2       | 0x03000B73                                |
+--------------------------+-------------------------------------------+
| SQL Server 2012          | 0x04000074                                |
|                          |                                           |
| SQL Server 2014          |                                           |
|                          |                                           |
| SQL Server 2016          |                                           |
|                          |                                           |
| SQL Server 2017          |                                           |
|                          |                                           |
| SQL Server 2019          |                                           |
|                          |                                           |
| SQL Server 2022\*        |                                           |
|                          |                                           |
| SQL Server 2025\*        |                                           |
+--------------------------+-------------------------------------------+

\*In TDS 7.x flow.

[\<18\> Section 2.2.6.4](\l): The value \"1\" for fByteOrder is
supported only by SQL Server 7.0.

[\<19\> Section 2.2.6.4](\l): SQL Server assumes fFloat to be
FLOAT_IEEE_754 and ignores the other settings.

[\<20\> Section 2.2.6.4](\l): For fODBC, SQL Server returns a value of
zero for ROWCOUNT.

[\<21\> Section 2.2.6.4](\l): For fOLEDB, SQL Server returns a value of
zero for ROWCOUNT.

[\<22\> Section 2.2.6.4](\l): SQL Server implementations do not inspect
the fSendYukonBinaryXML bit. When using the .NET Framework Data Provider
for SQL Server, the server sends binary XML if the TDS version is 7.2 or
later.

[\<23\> Section 2.2.6.4](\l): The FEDAUTH feature extension is not
supported by SQL Server.

[\<24\> Section 2.2.6.4](\l): The COLUMNENCRYPTION feature extension is
not supported by SQL Server 7.0, SQL Server 2000, SQL Server 2005, SQL
Server 2008, SQL Server 2008 R2, SQL Server 2012, and SQL Server 2014.

[\<25\> Section 2.2.6.4](\l): Enclave computations are not supported by
SQL Server 7.0, SQL Server 2000, SQL Server 2005, SQL Server 2008, SQL
Server 2008 R2, SQL Server 2012, SQL Server 2014, SQL Server 2016, and
SQL Server 2017. Support for this functionality was introduced in the
.NET Framework 4.7.2 and is not supported by the .NET Framework 1.1,
.NET Framework 2.0, .NET Framework 4.0, .NET Framework 4.5, .NET
Framework 4.6, .NET Framework 4.7, and .NET Framework 4.7.1.

[\<26\> Section 2.2.6.4](\l): Enclave computations with cached column
encryption keys are not supported by SQL Server 7.0, SQL Server 2000,
SQL Server 2005, SQL Server 2008, SQL Server 2008 R2, SQL Server 2012,
SQL Server 2014, SQL Server 2016, SQL Server 2017, and SQL Server 2019.

[\<27\> Section 2.2.6.4](\l): The GLOBALTRANSACTIONS feature extension
is not supported by SQL Server.

[\<28\> Section 2.2.6.4](\l): The AZURESQLSUPPORT feature extension is
not supported by SQL Server. This feature extension was introduced in
.NET Framework 4.7.2 and is not supported by the .NET Framework 1.1,
.NET Framework 2.0, .NET Framework 4.0, .NET Framework 4.5, .NET
Framework 4.6, .NET Framework 4.7, and .NET Framework 4.7.1.

[\<29\> Section 2.2.6.4](\l): The DATACLASSIFICATION feature extension
is not supported by SQL Server 7.0, SQL Server 2000, SQL Server 2005,
SQL Server 2008, SQL Server 2008 R2, SQL Server 2012, SQL Server 2014,
SQL Server 2016, and SQL Server 2017.

[\<30\> Section 2.2.6.4](\l): The UTF8_SUPPORT feature extension is not
supported by SQL Server 7.0, SQL Server 2000, SQL Server 2005, SQL
Server 2008, SQL Server 2008 R2, SQL Server 2012, SQL Server 2014, SQL
Server 2016, and SQL Server 2017.

[\<31\> Section 2.2.6.4](\l): The AZURESQLDNSCACHING feature extension
is not supported by SQL Server.

[\<32\> Section 2.2.6.4](): The JSONSUPPORT feature extension is not
supported by SQL Server 7.0, SQL Server 2000, SQL Server 2005, SQL
Server 2008, SQL Server 2008 R2, SQL Server 2012, SQL Server 2014, SQL
Server 2016, SQL Server 2017, SQL Server 2019, and SQL Server 2022.

[\<33\> Section 2.2.6.4](\l): The VECTORSUPPORT feature extension is not
supported by SQL Server 7.0, SQL Server 2000, SQL Server 2005, SQL
Server 2008, SQL Server 2008 R2, SQL Server 2012, SQL Server 2014, SQL
Server 2016, SQL Server 2017, SQL Server 2019, and SQL Server 2022.

[\<34\> Section 2.2.6.4](\l): The ENHANCEDROUTINGSUPPORT feature
extension is not supported by SQL Server.

\<35\> Section 2.2.6.5: The FEDAUTHREQUIRED payload option token is not
supported by SQL Server.

[\<36\> Section 2.2.6.5](\l): The ENCRYPT_EXT bit is not supported by
SQL Server and is ignored. When the client driver uses this bit to log
in to Azure SQL Database, the session could be disconnected.

[\<37\> Section 2.2.6.5](\l): The ENCRYPT_CLIENT_CERT setting is used
only when SQL Server is running on a Linux operating system and is not
supported by SQL Server 7.0, SQL Server 2000, SQL Server 2005, SQL
Server 2008, SQL Server 2008 R2, SQL Server 2012, SQL Server 2014, SQL
Server 2016, and SQL Server 2017.

[\<38\> Section 2.2.6.5](\l): Of the SQL Server products that are
applicable to this specification, with the exception of SQL Server 7.0,
SQL Server 2000, SQL Server 2005, SQL Server 2008, and SQL Server 2008
R2, the server always sends the value 0 for the INSTOPT option when the
string specified in the client\'s INSTOPT option is \"MSSQLServer\". The
reason for this is that \"MSSQLServer\" is the name of a default
instance, and \"MSSQLServer\" can be provided by the client even in the
absence of an explicit instance name. SQL Server 2000, SQL Server 2005,
SQL Server 2008, and SQL Server 2008 R2, which support the INSTOPT
field, always validate the client-specified string against the server\'s
instance name.

[\<39\> Section 2.2.6.6](\l): The fNoMetaData flag is supported only by
SQL Server 7.0, SQL Server 2000, SQL Server 2005, SQL Server 2008, SQL
Server 2008 R2, SQL Server 2012, and SQL Server 2014.

[\<40\> Section 2.2.6.6](\l): The EnclavePackage parameter is not
supported by SQL Server 7.0, SQL Server 2000, SQL Server 2005, SQL
Server 2008, SQL Server 2008 R2, SQL Server 2012, SQL Server 2014, SQL
Server 2016, and SQL Server 2017. This parameter was introduced in the
.NET Framework 4.7.2 and is not supported by .NET Framework 1.1, .NET
Framework 2.0, .NET Framework 4.0, .NET Framework 4.5, .NET Framework
4.6, .NET Framework 4.7, and .NET Framework 4.7.1.

[\<41\> Section 2.2.7.1](\l): ALTMETADATA_TOKEN is supported only by SQL
Server 7.0, SQL Server 2000, SQL Server 2005, SQL Server 2008, and SQL
Server 2008 R2.

[\<42\> Section 2.2.7.2](\l): ALTROW_TOKEN is supported only by SQL
Server 7.0, SQL Server 2000, SQL Server 2005, SQL Server 2008, and SQL
Server 2008 R2.

[\<43\> Section 2.2.7.4](\l): Of the SQL Server products that are
applicable to this specification, with the exception of SQL Server 7.0,
SQL Server 2000, SQL Server 2005, SQL Server 2008, SQL Server 2008 R2,
SQL Server 2012, and SQL Server 2014, SQL Server supports the fHidden
flag only through a many-to-many result and by connecting via ODBC.

[\<44\> Section 2.2.7.4](\l): The NoMetaData parameter is supported only
by SQL Server 7.0, SQL Server 2000, SQL Server 2005, SQL Server 2008,
SQL Server 2008 R2, SQL Server 2012, and SQL Server 2014.

[\<45\> Section 2.2.7.5](\l): The DATACLASSIFICATION token is not
supported by SQL Server 7.0, SQL Server 2000, SQL Server 2005, SQL
Server 2008, SQL Server 2008 R2, SQL Server 2012, SQL Server 2014, SQL
Server 2016, and SQL Server 2017.

[\<46\> Section 2.2.7.6](\l): The 0x4: DONE_INXACT bit is not set by SQL
Server and is reserved for future use.

[\<47\> Section 2.2.7.6](\l): The DONE token is usually sent after login
has succeeded. In this case, the negotiated TDS version is known, and
the client can determine whether DoneRowCount is LONG or ULONGLONG.

When login fails for any reason, SQL Server might also send an error
message followed by a [DONE](#Section_3c06f11098bd4d5bb836b1ba66452cb7)
token. In this case, the server has already completed TDS version
negotiation and has to send DoneRowCount as LONG or ULONGLONG based on
the negotiated TDS version.

However, sometimes the client cannot determine the server TDS version
and cannot determine whether LONG or ULONGLONG is expected for
DoneRowCount. If the client TDS level is 7.0 or 7.1, DoneRowCount is
always LONG. If the client TDS level is 7.2, 7.3.A, 7.3.B, 7.4, or 8.0,
the DoneRowCount can be LONG or ULONGLONG, depending on which version of
the server the client is connecting to.

[**SQL Server Native Client
(SNAC)**](#gt_f5b7b5b5-50e8-4f63-8e66-ad8a30b229f2)
[\[MSDN-SNAC\]](https://go.microsoft.com/fwlink/?LinkId=213738) and
SQLClient use the VERSION option in the Pre-Login Response message to
detect whether DoneRowCount is LONG or ULONGLONG. It is LONG if VERSION
in the Pre-Login Response message indicates that the server is SQL
Server 7.0 or SQL Server 2000. Otherwise, DoneRowCount is ULONGLONG.

A third-party implementation has its own logic to detect whether
DoneRowCount is LONG or ULONGLONG or to make the client able to handle
both LONG and ULONGLONG. In any implementation, before the client
performs this task, the server performs TDS version negotiation and
determines whether to send LONG or ULONGLONG.

[\<48\> Section 2.2.7.7](\l): The 0x4: DONE_INXACT bit is not set by SQL
Server and is reserved for future use.

[\<49\> Section 2.2.7.8](\l): The 0x4: DONE_INXACT bit is not set by SQL
Server and is reserved for future use.

[\<50\> Section 2.2.7.9](\l): Type 16: Transaction Manager Address is
not used by SQL Server.

[\<51\> Section 2.2.7.10](\l): Numbers less than 20001 are reserved by
SQL Server.

[\<52\> Section 2.2.7.10](\l): SQL Server does not raise system errors
with severities of 0 through 9.

[\<53\> Section 2.2.7.10](\l): For compatibility reasons, SQL Server
converts severity 10 to severity 0 before returning the error
information to the calling application.

[\<54\> Section 2.2.7.11](\l): The FEDAUTH feature extension is not
supported by SQL Server.

[\<55\> Section 2.2.7.11](\l): The COLUMNENCRYPTION feature extension is
not supported by SQL Server 7.0, SQL Server 2000, SQL Server 2005, SQL
Server 2008, SQL Server 2008 R2, SQL Server 2012, and SQL Server 2014.

[\<56\> Section 2.2.7.11](\l): Enclave computations are not supported by
SQL Server 7.0, SQL Server 2000, SQL Server 2005, SQL Server 2008, SQL
Server 2008 R2, SQL Server 2012, SQL Server 2014, SQL Server 2016, and
SQL Server 2017. Support for this feature was introduced in the .NET
Framework 4.7.2 and is not supported by the .NET Framework 1.1, .NET
Framework 2.0, .NET Framework 4.0, .NET Framework 4.5, .NET Framework
4.6, .NET Framework 4.7, and .NET Framework 4.7.1.

\<57\> Section 2.2.7.11: Enclave computations with cached column
encryption keys are not supported by SQL Server 7.0, SQL Server 2000,
SQL Server 2005, SQL Server 2008, SQL Server 2008 R2, SQL Server 2012,
SQL Server 2014, SQL Server 2016, SQL Server 2017, and SQL Server 2019.

[\<58\> Section 2.2.7.11](\l): The **EnclaveType** field is not
supported by SQL Server 7.0, SQL Server 2000, SQL Server 2005, SQL
Server 2008, SQL Server 2008 R2, SQL Server 2012, SQL Server 2014, SQL
Server 2016, and SQL Server 2017. This field was introduced in the .NET
Framework 4.7.2 and is not supported by the .NET Framework 1.1, .NET
Framework 2.0, .NET Framework 4.0, .NET Framework 4.5, .NET Framework
4.6, .NET Framework 4.7, and .NET Framework 4.7.1.

[\<59\> Section 2.2.7.11](\l): The GLOBALTRANSACTIONS feature extension
is not supported by SQL Server.

[\<60\> Section 2.2.7.11](\l): The AZURESQLSUPPORT feature extension is
not supported by SQL Server. This feature extension was introduced in
.NET Framework 4.7.2 and is not supported by the .NET Framework 1.1,
.NET Framework 2.0, .NET Framework 4.0, .NET Framework 4.5, .NET
Framework 4.6, .NET Framework 4.7, and .NET Framework 4.7.1.

[\<61\> Section 2.2.7.11](\l): The DATACLASSIFICATION feature extension
is not supported by SQL Server 7.0, SQL Server 2000, SQL Server 2005,
SQL Server 2008, SQL Server 2008 R2, SQL Server 2012, SQL Server 2014,
SQL Server 2016, and SQL Server 2017.

[\<62\> Section 2.2.7.11](\l): The UTF8_SUPPORT feature extension is not
supported by SQL Server 7.0, SQL Server 2000, SQL Server 2005, SQL
Server 2008, SQL Server 2008 R2, SQL Server 2012, SQL Server 2014, SQL
Server 2016, and SQL Server 2017.

[\<63\> Section 2.2.7.11](\l): The AZURESQLDNSCACHING feature extension
is not supported by SQL Server. The FeatureData value is always 0.

[\<64\> Section 2.2.7.11](\l): The JSONSUPPORT feature extension is not
supported by SQL Server 7.0, SQL Server 2000, SQL Server 2005, SQL
Server 2008, SQL Server 2008 R2, SQL Server 2012, SQL Server 2014, SQL
Server 2016, SQL Server 2017, SQL Server 2019, and SQL Server 2022.

[\<65\> Section 2.2.7.11](\l): The VECTORSUPPORT feature extension is
not supported by SQL Server 7.0, SQL Server 2000, SQL Server 2005, SQL
Server 2008, SQL Server 2008 R2, SQL Server 2012, SQL Server 2014, SQL
Server 2016, SQL Server 2017, SQL Server 2019, and SQL Server 2022.

[\<66\> Section 2.2.7.11](\l): The ENHANCEDROUTINGSUPPORT feature
extension is not supported by SQL Server.

[\<67\> Section 2.2.7.12](\l): The FEDAUTHINFO token is not supported by
SQL Server.

[\<68\> Section 2.2.7.13](\l): Numbers less than 20001 are reserved by
SQL Server.

[\<69\> Section 2.2.7.14](\l): The following table shows the values in
network transfer format.

+--------------------------+---------------------+---------------------+
| SQL Server               | Client to server    | Server to client    |
+==========================+=====================+=====================+
| SQL Server 7.0           | 0x00000070          | 0x07000000          |
+--------------------------+---------------------+---------------------+
| SQL Server 2000          | 0x00000071          | 0x07010000          |
+--------------------------+---------------------+---------------------+
| SQL Server 2000 SP1      | 0x01000071          | 0x71000001          |
+--------------------------+---------------------+---------------------+
| SQL Server 2005          | 0x02000972          | 0x72090002          |
+--------------------------+---------------------+---------------------+
| SQL Server 2008\*        | 0x03000A73          | 0x730A0003          |
+--------------------------+---------------------+---------------------+
| SQL Server 2008 R2       | 0x03000B73          | 0x730B0003          |
+--------------------------+---------------------+---------------------+
| SQL Server 2012          | 0x04000074          | 0x74000004          |
|                          |                     |                     |
| SQL Server 2014          |                     |                     |
|                          |                     |                     |
| SQL Server 2016          |                     |                     |
|                          |                     |                     |
| SQL Server 2017          |                     |                     |
|                          |                     |                     |
| SQL Server 2019          |                     |                     |
|                          |                     |                     |
| SQL Server 2022\*\*      |                     |                     |
|                          |                     |                     |
| SQL Server 2025\*\*      |                     |                     |
+--------------------------+---------------------+---------------------+

\*SQL Server 2008 TDS version 0x03000A73 does not include support for
NBCROW and fSparseColumnSet.

\*\*In TDS 7.x flow.

[\<70\> Section 3.2.2](\l): In Microsoft implementations, the default
value for the [**Microsoft/Windows Data Access Components
(MDAC/WDAC)**](#gt_4f587fbb-99d6-4d8c-aebd-b4b325ce64d8) and SNAC Client
Request Timers is zero, which is interpreted as no timeout. For a
SqlClient Client Request, the default value is 30 seconds. For a
description of the data access drivers, see
[\[MSDN-MDAC\]](https://go.microsoft.com/fwlink/?LinkId=213737).

[\<71\> Section 3.2.2](\l): In Microsoft implementations, the default
setting for MDAC/WDAC and SNAC Cancel Timer values is 120 seconds. For a
SqlClient Cancel Timer, the default value is 5 seconds. For a
description of the data access drivers, see \[MSDN-MDAC\].

