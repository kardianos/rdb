# Introduction

The Tabular Data Stream (TDS) protocol versions 7 and 8 is an
application layer request/response protocol that facilitates interaction
with a database server and provides for the following:

-   Authentication and channel encryption.

-   Specification of requests in SQL (including Bulk Insert).

-   Invocation of a [**stored
    procedure**](#gt_324d32b3-f4f3-41c9-b695-78c498094fb7) or
    user-defined function, also known as a [**remote procedure call
    (RPC)**](#gt_8a7f6700-8311-45bc-af10-82e10accd331).

-   The return of data.

-   [**Transaction manager**](#gt_4553803e-9d8d-407c-ad7d-9e65e01d6eb3)
    requests.

Sections 1.5, 1.8, 1.9, 2, and 3 of this specification are normative.
All other sections and examples in this specification are informative.

## Glossary

This document uses the following terms:

> []{#gt_5a728127-a59d-4cf9-8ab5-4a4e0747cc51 .anchor}**Azure Active
> Directory Authentication Library (ADAL)**: A tool in Microsoft .NET
> Framework that allows application developers to authenticate users
> either to the cloud or to a deployed on-premises Active Directory and
> to then obtain tokens for secure access to API calls.
>
> []{#gt_6f6f9e8e-5966-4727-8527-7e02fb864e7e .anchor}**big-endian**:
> Multiple-byte values that are byte-ordered with the most significant
> byte stored in the memory location with the lowest address.
>
> []{#gt_81e2b338-0f2a-4980-9824-9c27bb6d1341 .anchor}**bulk insert**: A
> method for efficiently populating the rows of a table from the client
> to the server.
>
> []{#gt_5f1d976e-cd4b-4a78-a6a1-d0bdb0aa0360 .anchor}**common language
> runtime user-defined type (CLR UDT)**: A data type that is created and
> defined by the user on a database server that supports SQL by using a
> Microsoft .NET Framework common language runtime assembly.
>
> []{#gt_acdeafb0-9b24-420e-b712-9284ad49eb56 .anchor}**data
> classification**: An information protection framework that includes
> sensitivity information about the data that is being returned from a
> query. The sensitivity information includes labels and information
> types and their identifiers.
>
> []{#gt_151643ce-fb5e-460e-8bdf-dc10bbd1950e .anchor}**data stream**: A
> stream of data that corresponds to specific Tabular Data Stream (TDS)
> semantics. A single data stream can represent an entire TDS message or
> only a specific, well-defined portion of a TDS message. A TDS data
> stream can span multiple network data packets.
>
> []{#gt_177a71e6-7cf0-48c1-b169-063da349648a .anchor}**Distributed
> Transaction Coordinator (DTC)**: A Windows service that coordinates
> transactions across multiple resource managers, including databases.
> For more information, see
> [\[MSDN-DTC\]](https://go.microsoft.com/fwlink/?LinkId=89994).
>
> []{#gt_ef41e9e0-3e5d-432f-96d6-39515bdc5340 .anchor}**enclave**: A
> protected region of memory that is used only on the server side. This
> region is within the address space of SQL Server, and it acts as a
> trusted execution environment. Only code that runs within the enclave
> can access data within that enclave. Neither the data nor the code
> inside the enclave can be viewed from the outside, even with a
> debugger.
>
> []{#gt_6fe5534f-5cd8-4ab6-aba4-637a4344eda0 .anchor}**enclave
> computations**: Locally enabled cryptographic operations and other
> operations in Transact-SQL queries on encrypted columns that are
> performed inside an enclave.
>
> []{#gt_5ae22a0e-5ff4-441b-80d4-224ef4dd4d19 .anchor}**federated
> authentication**: An authentication mechanism that allows a security
> token service (STS) in one trust domain to delegate user
> authentication to an identity provider in another trust domain, while
> generating a security token for the user, when there is a trust
> relationship between the two domains.
>
> []{#gt_57552b13-b14e-4601-9621-500ce3297d15 .anchor}**Global
> Transactions**: A feature that allows users to execute transactions
> across multiple databases that are hosted in a shared service, such as
> Microsoft Azure SQL Database.
>
> []{#gt_95913fbd-3262-47ae-b5eb-18e6806824b9 .anchor}**interface**: A
> group of related function prototypes in a specific order, analogous to
> a C++ virtual interface. Multiple objects, of different object
> classes, can implement the same interface. A derived interface can be
> created by adding methods after the end of an existing interface. In
> the Distributed Component Object Model (DCOM), all interfaces
> initially derive from IUnknown.
>
> []{#gt_079478cb-f4c5-4ce5-b72b-2144da5d2ce7 .anchor}**little-endian**:
> Multiple-byte values that are byte-ordered with the least significant
> byte stored in the memory location with the lowest address.
>
> []{#gt_4f587fbb-99d6-4d8c-aebd-b4b325ce64d8
> .anchor}**Microsoft/Windows Data Access Components (MDAC/WDAC)**: With
> Microsoft/Windows Data Access Components (MDAC/WDAC), developers can
> connect to and use data from a wide variety of relational and
> nonrelational data sources. You can connect to many different data
> sources using Open Database Connectivity (ODBC), ActiveX Data Objects
> (ADO), or OLE DB. You can do this through providers and drivers that
> are built and shipped by Microsoft, or that are developed by various
> third parties. For more information, see
> [\[MSDN-MDAC\]](https://go.microsoft.com/fwlink/?LinkId=213737).
>
> []{#gt_762fe1e3-0979-4402-b963-1e9150de133d .anchor}**Multiple Active
> Result Sets (MARS)**: A feature in Microsoft SQL Server that allows
> applications to have more than one pending request per connection. For
> more information, see
> [\[MSDN-MARS\]](https://go.microsoft.com/fwlink/?LinkId=98459).
>
> []{#gt_16dd540d-3913-48e5-9a93-a769e85570d0 .anchor}**nullable
> column**: A database table column that is allowed to contain no value
> for a given row.
>
> []{#gt_26c1caf3-c889-4b99-a22b-9da056d397cf .anchor}**out-of-band**: A
> type of event that happens outside of the standard sequence of events.
> For example, an out-of-band signal or message can be sent during an
> unexpected time and will not cause any protocol parsing issues.
>
> []{#gt_62a2f252-bd68-4d77-a751-d6ff27010678 .anchor}**query
> notification**: A feature in SQL Server that allows the client to
> register for notification on changes to a given query result. For more
> information, see
> [\[MSDN-QUERYNOTE\]](https://go.microsoft.com/fwlink/?LinkId=119984).
>
> []{#gt_8a7f6700-8311-45bc-af10-82e10accd331 .anchor}**remote procedure
> call (RPC)**: A communication protocol used primarily between client
> and server. The term has three definitions that are often used
> interchangeably: a runtime environment providing for communication
> facilities between computers (the RPC runtime); a set of
> request-and-response message exchanges between computers (the RPC
> exchange); and the single message from an RPC exchange (the RPC
> message). For more information, see
> [\[C706\]](https://go.microsoft.com/fwlink/?LinkId=89824).
>
> []{#gt_c8a27238-8ccc-442b-9604-75f74d3e6b3d .anchor}**result set**: A
> list of records that results from running a stored procedure or query,
> or applying a filter. The structure and content of the data in a
> result set varies according to the implementation.
>
> []{#gt_fb216516-748b-4873-8bdd-64c5f4da9920 .anchor}**Security Support
> Provider Interface (SSPI)**: An API that allows connected applications
> to call one of several security providers to establish authenticated
> connections and to exchange data securely over those connections. It
> is equivalent to Generic Security Services (GSS)-API, and the two are
> on-the-wire compatible.
>
> []{#gt_f70f98cc-c555-4a40-9509-bc1da4021211 .anchor}**Session
> Multiplex Protocol (SMP)**: A multiplexing protocol that enables
> multiple logical client connections to share a single transport
> connection to a server. Used by [**Multiple Active Result Sets
> (MARS)**](#gt_762fe1e3-0979-4402-b963-1e9150de133d). For more
> information, see
> [\[MC-SMP\]](%5bMC-SMP%5d.pdf#Section_04c8edde371d4af5bb33a39b3948f0af).
>
> []{#gt_bc2f6b5e-e5c0-408b-8f55-0350c24b9838 .anchor}**Simple and
> Protected GSS-API Negotiation Mechanism (SPNEGO)**: An authentication
> mechanism that allows Generic Security Services (GSS) peers to
> determine whether their credentials support a common set of GSS-API
> security mechanisms, to negotiate different options within a given
> security mechanism or different options from several security
> mechanisms, to select a service, and to establish a security context
> among themselves using that service.
> [**SPNEGO**](#gt_bc2f6b5e-e5c0-408b-8f55-0350c24b9838) is specified in
> [\[RFC4178\]](https://go.microsoft.com/fwlink/?LinkId=90461).
>
> []{#gt_1dfba466-175f-4050-a0e7-c1baf187d21d .anchor}**SQL batch**: A
> set of [**SQL statements**](#gt_dc5ca224-43ec-4b44-9dba-726d6fd6057d).
>
> []{#gt_f5b7b5b5-50e8-4f63-8e66-ad8a30b229f2 .anchor}**SQL Server
> Native Client (SNAC)**: SNAC contains the SQL Server ODBC driver and
> the SQL Server OLE DB provider in one native dynamic link library
> (DLL) supporting applications using native-code APIs (ODBC, OLE DB,
> and ADO) to Microsoft SQL Server. For more information, see
> [\[MSDN-SNAC\]](https://go.microsoft.com/fwlink/?LinkId=213738).
>
> []{#gt_1a186995-43e5-4e46-896a-bad208ec2551 .anchor}**SQL Server User
> Authentication (SQLAUTH)**: An authentication mechanism that is used
> to support user accounts on a database server that supports SQL. The
> username and password of the user account are transmitted as part of
> the login message that the client sends to the server.
>
> []{#gt_dc5ca224-43ec-4b44-9dba-726d6fd6057d .anchor}**SQL statement**:
> A character string expression in a language that the server
> understands.
>
> []{#gt_324d32b3-f4f3-41c9-b695-78c498094fb7 .anchor}**stored
> procedure**: A precompiled collection of SQL statements and,
> optionally, control-of-flow statements that are stored under a name
> and processed as a unit. They are stored in a SQL database and can be
> run with one call from an application. Stored procedures return an
> integer return code and can additionally return one or more result
> sets. Also referred to as sproc.
>
> []{#gt_71dd1dd2-c167-49a8-a5f5-6b0df5c8b48a .anchor}**table
> response**: A collection of data, all formatted in a specific manner,
> that is sent by the server to the client for the purpose of
> communicating the result of a client request. The server returns the
> result in a table response format for many types of client requests
> such as LOGIN7, SQL, and remote procedure call (RPC) requests.
>
> []{#gt_276cd76b-c0a2-4f7c-8529-ad0d60aa9592 .anchor}**TDS session**: A
> successfully established communication over a period of time between a
> client and a server on which the Tabular Data Stream (TDS) protocol is
> used for message exchange.
>
> []{#gt_4553803e-9d8d-407c-ad7d-9e65e01d6eb3 .anchor}**transaction
> manager**: The party that is responsible for managing and distributing
> the outcome of atomic transactions. A transaction manager is either a
> root transaction manager or a subordinate transaction manager for a
> specified transaction.
>
> []{#gt_c305d0ab-8b94-461a-bd76-13b40cb8c4d8 .anchor}**Unicode**: A
> character encoding standard developed by the Unicode Consortium that
> represents almost all of the written languages of the world. The
> [**Unicode**](#gt_c305d0ab-8b94-461a-bd76-13b40cb8c4d8) standard
> [\[UNICODE5.0.0/2007\]](https://go.microsoft.com/fwlink/?LinkId=154659)
> provides three forms (UTF-8, UTF-16, and UTF-32) and seven schemes
> (UTF-8, UTF-16, UTF-16 BE, UTF-16 LE, UTF-32, UTF-32 LE, and UTF-32
> BE).
>
> []{#gt_c35909bd-185e-4b60-be82-995a0318873e .anchor}**Virtual
> Interface Architecture (VIA)**: A high-speed interconnect that
> requires special hardware and drivers that are provided by third
> parties.
>
> **MAY, SHOULD, MUST, SHOULD NOT, MUST NOT:** These terms (in all caps)
> are used as defined in
> [\[RFC2119\]](https://go.microsoft.com/fwlink/?LinkId=90317). All
> statements of optional behavior use either MAY, SHOULD, or SHOULD NOT.

## References

Links to a document in the Microsoft Open Specifications library point
to the correct section in the most recently published version of the
referenced document. However, because individual documents in the
library are not updated at the same time, the section numbers in the
documents may not match. You can confirm the correct section numbering
by checking the
[Errata](https://go.microsoft.com/fwlink/?linkid=850906).

### Normative References

We conduct frequent surveys of the normative references to assure their
continued availability. If you have any issue with finding a normative
reference, please contact <dochelp@microsoft.com>. We will assist you in
finding the relevant information.

\[IANAPORT\] IANA, \"Service Name and Transport Protocol Port Number
Registry\",
[https://www.iana.org/assignments/service-names-port-numbers/service-names-port-numbers.xhtml](https://go.microsoft.com/fwlink/?LinkId=89888)

\[IEEE754\] IEEE, \"IEEE Standard for Binary Floating-Point
Arithmetic\", IEEE 754-1985, October 1985,
[http://ieeexplore.ieee.org/servlet/opac?punumber=2355](https://go.microsoft.com/fwlink/?LinkId=89903)

\[IETF-AuthEncr\] McGrew, D., Foley, J., and Paterson, K.,
\"Authenticated Encryption with AES-CBC and HMAC-SHA\", Network Working
Group Internet-Draft, July 2014,
[http://tools.ietf.org/html/draft-mcgrew-aead-aes-cbc-hmac-sha2-05](https://go.microsoft.com/fwlink/?LinkId=524322)

\[MS-BINXML\] Microsoft Corporation, \"[SQL Server Binary XML
Structure](%5bMS-BINXML%5d.pdf#Section_11ab6e8d247244d1a9e6bddf000e12f6)\".

\[MS-LCID\] Microsoft Corporation, \"[Windows Language Code Identifier
(LCID)
Reference](%5bMS-LCID%5d.pdf#Section_70feba9f294e491eb6eb56532684c37f)\".

\[RFC1122\] Braden, R., Ed., \"Requirements for Internet Hosts \--
Communication Layers\", STD 3, RFC 1122, October 1989,
[https://www.rfc-editor.org/rfc/rfc1122](https://go.microsoft.com/fwlink/?LinkId=112180)

\[RFC2119\] Bradner, S., \"Key words for use in RFCs to Indicate
Requirement Levels\", BCP 14, RFC 2119, March 1997,
[https://www.rfc-editor.org/info/rfc2119](https://go.microsoft.com/fwlink/?LinkId=90317)

\[RFC2246\] Dierks, T., and Allen, C., \"The TLS Protocol Version 1.0\",
RFC 2246, January 1999,
[https://www.rfc-editor.org/info/rfc2246](https://go.microsoft.com/fwlink/?LinkId=90324)

\[RFC4234\] Crocker, D., Ed., and Overell, P., \"Augmented BNF for
Syntax Specifications: ABNF\", RFC 4234, October 2005,
[https://www.rfc-editor.org/info/rfc4234](https://go.microsoft.com/fwlink/?LinkId=90462)

\[RFC5246\] Dierks, T., and Rescorla, E., \"The Transport Layer Security
(TLS) Protocol Version 1.2\", RFC 5246, August 2008,
[https://www.rfc-editor.org/info/rfc5246](https://go.microsoft.com/fwlink/?LinkId=129803)

\[RFC6101\] Freier, A., Karlton, P., and Kocher, P., \"The Secure
Sockets Layer (SSL) Protocol Version 3.0\", RFC 6101, August 2011,
[http://www.rfc-editor.org/rfc/rfc6101.txt](https://go.microsoft.com/fwlink/?LinkId=509953)

\[RFC6234\] Eastlake III, D., and Hansen, T., \"US Secure Hash
Algorithms (SHA and SHA-based HMAC and HKDF)\", RFC 6234, May 2011,
[http://www.rfc-editor.org/rfc/rfc6234.txt](https://go.microsoft.com/fwlink/?LinkId=328921)

\[RFC7301\] Friedl, S., Popov, A., Langley, A., and Stephan, E.,
\"Transport Layer Security (TLS) Application-Layer Protocol Negotiation
Extension\", RFC 7301, July 2014,
[https://www.rfc-editor.org/info/rfc7301](https://go.microsoft.com/fwlink/?LinkId=513846)

\[RFC793\] Postel, J., Ed., \"Transmission Control Protocol: DARPA
Internet Program Protocol Specification\", RFC 793, September 1981,
[https://www.rfc-editor.org/info/rfc793](https://go.microsoft.com/fwlink/?LinkId=150872)

\[RFC8259\] Bray, T., Ed., \"The JavaScript Object Notation (JSON) Data
Interchange Format\", RFC 8259, December 2017,
[https://www.rfc-editor.org/info/rfc8259](https://go.microsoft.com/fwlink/?linkid=867803)

\[RFC8446\] Rescorla, E., \"The Transport Layer Security (TLS) Protocol
Version 1.3\", RFC 8446, August 2018,
[https://www.rfc-editor.org/info/rfc8446](https://go.microsoft.com/fwlink/?linkid=2147431)

\[UNICODE\] The Unicode Consortium, \"The Unicode Consortium Home
Page\",
[http://www.unicode.org/](https://go.microsoft.com/fwlink/?LinkId=90550)

\[VIA2002\] Cameron, D., and Regnier, G., \"The Virtual Interface
Architecture\", Intel Press, 2002, ISBN:0971288704.

### Informative References

\[MC-SMP\] Microsoft Corporation, \"[Session Multiplex
Protocol](%5bMC-SMP%5d.pdf#Section_04c8edde371d4af5bb33a39b3948f0af)\".

\[MS-NETOD\] Microsoft Corporation, \"[Microsoft .NET Framework
Protocols
Overview](%5bMS-NETOD%5d.pdf#Section_bcca8164da0843f2a983c34ed99171b0)\".

\[MS-SSCLRT\] Microsoft Corporation, \"[Microsoft SQL Server CLR Types
Serialization
Formats](%5bMS-SSCLRT%5d.pdf#Section_77460aa98c2f4449a65e1d649ebd77fa)\".

\[MSDN-Autocommit\] Microsoft Corporation, \"Autocommit Transactions\",
[https://learn.microsoft.com/en-us/previous-versions/sql/sql-server-2008-r2/ms187878(v=sql.105)](https://go.microsoft.com/fwlink/?LinkId=145156)

\[MSDN-BEGIN\] Microsoft Corporation, \"BEGIN TRANSACTION (Transact
SQL)\",
[https://learn.microsoft.com/en-us/sql/t-sql/language-elements/begin-transaction-transact-sql](https://go.microsoft.com/fwlink/?LinkId=144544)

\[MSDN-BOUND\] Microsoft Corporation, \"Using Bound Sessions\",
[https://learn.microsoft.com/en-us/previous-versions/sql/sql-server-2008-r2/ms177480(v=sql.105)](https://go.microsoft.com/fwlink/?LinkId=144543)

\[MSDN-BROWSE\] Microsoft Corporation, \"Browse Mode\", in SQL Server
2000 Retired Technical documentation, p. 12261,
[https://www.microsoft.com/en-us/download/confirmation.aspx?id=51958](https://go.microsoft.com/fwlink/?LinkId=140931)

\[MSDN-Collation\] Microsoft Corporation, \"Collation and Unicode
Support\",
[https://learn.microsoft.com/en-us/sql/relational-databases/collations/collation-and-unicode-support](https://go.microsoft.com/fwlink/?LinkId=233327)

\[MSDN-ColSets\] Microsoft Corporation, \"Use Column Sets\",
[https://learn.microsoft.com/en-us/sql/relational-databases/tables/use-column-sets](https://go.microsoft.com/fwlink/?LinkId=128616)

\[MSDN-ColSortSty\] Microsoft Corporation, \"Windows Collation Sorting
Styles\",
[https://learn.microsoft.com/en-us/previous-versions/sql/sql-server-2008-r2/ms143515(v=sql.105)](https://go.microsoft.com/fwlink/?LinkId=233328)

\[MSDN-COMMIT\] Microsoft Corporation, \"COMMIT TRANSACTION
(Transact-SQL)\",
[https://learn.microsoft.com/en-us/sql/t-sql/language-elements/commit-transaction-transact-sql](https://go.microsoft.com/fwlink/?LinkId=144542)

\[MSDN-DTC\] Microsoft Corporation, \"Distributed Transaction
Coordinator\",
[https://learn.microsoft.com/en-us/previous-versions/windows/desktop/ms684146(v=vs.85)](https://go.microsoft.com/fwlink/?LinkId=89994)

\[MSDN-INSERT\] Microsoft Corporation, \"INSERT (Transact-SQL)\",
[https://learn.microsoft.com/en-us/sql/t-sql/statements/insert-transact-sql](https://go.microsoft.com/fwlink/?LinkId=154273)

\[MSDN-ITrans\] Microsoft Corporation,
\"ITransactionExport::GetTransactionCookie\",
[https://learn.microsoft.com/en-us/previous-versions/windows/desktop/ms679869(v=vs.85)](https://go.microsoft.com/fwlink/?LinkId=146594)

\[MSDN-MARS\] Microsoft Corporation, \"Using Multiple Active Result Sets
(MARS)\",
[https://learn.microsoft.com/en-us/sql/relational-databases/native-client/features/using-multiple-active-result-sets-mars](https://go.microsoft.com/fwlink/?LinkId=98459)

\[MSDN-MDAC\] Wilkes, R., Bunch, A., and Dove, D., \"Microsoft Data
Access Components (MDAC) Installation\", May 2005,
[https://learn.microsoft.com/en-us/previous-versions/ms810805(v=msdn.10)](https://go.microsoft.com/fwlink/?LinkId=213737)

\[MSDN-NamedPipes\] Microsoft Corporation, \"Creating a Valid Connection
String Using Named Pipes\",
[https://learn.microsoft.com/en-us/previous-versions/sql/sql-server-2008-r2/ms189307(v=sql.105)](https://go.microsoft.com/fwlink/?LinkId=127839)

\[MSDN-NP\] Microsoft Corporation, \"Named Pipes\",
[https://learn.microsoft.com/en-us/windows/desktop/ipc/named-pipes](https://go.microsoft.com/fwlink/?LinkId=90247)

\[MSDN-NTLM\] Microsoft Corporation, \"Microsoft NTLM\",
[https://learn.microsoft.com/en-us/windows/desktop/SecAuthN/microsoft-ntlm](https://go.microsoft.com/fwlink/?LinkId=145227)

\[MSDN-QUERYNOTE\] Microsoft Corporation, \"Using Query Notifications\",
[https://learn.microsoft.com/en-us/previous-versions/sql/sql-server-2008-r2/ms175110(v=sql.105)](https://go.microsoft.com/fwlink/?LinkId=119984)

\[MSDN-SNAC\] Microsoft Corporation, \"Microsoft SQL Server Native
Client and Microsoft SQL Server 2008 Native Client\",
[https://learn.microsoft.com/en-us/archive/blogs/sqlnativeclient/microsoft-sql-server-native-client-and-microsoft-sql-server-2008-native-client](https://go.microsoft.com/fwlink/?LinkId=213738)

\[MSDN-SQLCollation\] Microsoft Corporation, \"Selecting a SQL Server
Collation\",
[https://learn.microsoft.com/en-us/previous-versions/sql/sql-server-2008-r2/ms144250(v=sql.105)](https://go.microsoft.com/fwlink/?LinkId=119987)

\[MSDN-TDSENDPT\] Microsoft Corporation, \"Network Protocols and TDS
Endpoints\",
[https://learn.microsoft.com/en-us/previous-versions/sql/sql-server-2008-r2/ms191220(v=sql.105)](https://go.microsoft.com/fwlink/?linkid=865399)

\[MSDN-UPDATETEXT\] Microsoft Corporation, \"UPDATETEXT
(Transact-SQL)\",
[https://learn.microsoft.com/en-us/sql/t-sql/queries/updatetext-transact-sql](https://go.microsoft.com/fwlink/?LinkId=154272)

\[MSDN-WRITETEXT\] Microsoft Corporation, \"WRITETEXT (Transact-SQL)\",
[https://learn.microsoft.com/en-us/sql/t-sql/queries/writetext-transact-sql](https://go.microsoft.com/fwlink/?LinkId=154269)

\[MSDOCS-DBMirror\] Microsoft Corporation, \"Database Mirroring in SQL
Server\",
[https://learn.microsoft.com/en-us/dotnet/framework/data/adonet/sql/database-mirroring-in-sql-server](https://go.microsoft.com/fwlink/?linkid=874052)

\[RFC4120\] Neuman, C., Yu, T., Hartman, S., and Raeburn, K., \"The
Kerberos Network Authentication Service (V5)\", RFC 4120, July 2005,
[https://www.rfc-editor.org/rfc/rfc4120](https://go.microsoft.com/fwlink/?LinkId=90458)

\[RFC4178\] Zhu, L., Leach, P., Jaganathan, K., and Ingersoll, W., \"The
Simple and Protected Generic Security Service Application Program
Interface (GSS-API) Negotiation Mechanism\", RFC 4178, October 2005,
[https://www.rfc-editor.org/info/rfc4178](https://go.microsoft.com/fwlink/?LinkId=90461)

\[SSPI\] Microsoft Corporation, \"SSPI\",
[https://learn.microsoft.com/en-us/windows/desktop/SecAuthN/sspi](https://go.microsoft.com/fwlink/?LinkId=90536)

## Overview

The Tabular Data Stream (TDS) protocol versions 7 and 8, hereinafter
referred to as \"TDS\", is an application-level protocol used for the
transfer of requests and responses between clients and database server
systems. In such systems, the client establishes a long-lived connection
with the server. Once the connection is established using a
transport-level protocol, TDS messages are used to communicate between
the client and the server. A database server can also act as the client
if needed, in which case a separate TDS connection has to be
established. Note that the [**TDS
session**](#gt_276cd76b-c0a2-4f7c-8529-ad0d60aa9592) is directly tied to
the transport-level session, meaning that a TDS session is established
when the transport-level connection is established and the server
receives a request to establish a TDS connection. It persists until the
transport-level connection is terminated (for example, when a TCP socket
is closed). In addition, TDS does not make any assumption about the
transport protocol used, but it does assume the transport protocol
supports reliable, in-order delivery of the data.

TDS includes facilities for authentication and identification, channel
encryption negotiation, issuing of [**SQL
batches**](#gt_1dfba466-175f-4050-a0e7-c1baf187d21d), [**stored
procedure**](#gt_324d32b3-f4f3-41c9-b695-78c498094fb7) calls, returning
data, and [**transaction
manager**](#gt_4553803e-9d8d-407c-ad7d-9e65e01d6eb3) requests. Returned
data is self-describing and record-oriented. The [**data
streams**](#gt_151643ce-fb5e-460e-8bdf-dc10bbd1950e) describe the names,
types, and optional descriptions of the rows being returned.

The difference between the TDS 7.x version family and the TDS 8.0
version centers on where and how network channel encryption is
initiated:

-   In the TDS 7.x version family, encryption is optional and is
    negotiated and handled in the TDS layer.

-   The TDS 8.0 version introduces mandatory encryption that is handled
    in the lower layer before TDS begins functioning.

The following diagram depicts a (simplified) typical flow of
communication in the TDS protocol.

![Communication flow in the TDS
protocol.](media/image1.bin "Communication flow in the TDS protocol"){width="3.8472222222222223in"
height="3.9375in"}

Figure 1: Communication flow in the TDS protocol

The following example is a high-level description of the messages
exchanged between the client and the server to execute a simple client
request such as the execution of a [**SQL
statement**](#gt_dc5ca224-43ec-4b44-9dba-726d6fd6057d). It is assumed
that the client and the server have already established a connection and
authentication has succeeded.

1.  Client:SQL statement

The server executes the SQL statement and then sends back the results to
the client. The data columns being returned are first described by the
server (represented as column metadata or COLMETADATA (section
[2.2.7.4](#Section_58880b9f381c43b2bf8b0727a98c4f4c)) and then the rows
follow. A completion message is sent after all the row data has been
transferred.

2.  Server:COLMETADATAdata stream

    ROWdata stream

    .

    .

    ROWdata stream

    DONEdata stream

For more information about the correlation between data stream and TDS
packet, see section
[2.2.4](#Section_dc3a08548230482fbbb9d94a3b905a26).[\<1\>](\l)

Additional details about which SQL Server version corresponds to which
TDS version number are defined in LOGINACK (section
[2.2.7.14](#Section_490e563dcc6e4c86bb95ef0186b98032)).

## Relationship to Other Protocols

The Tabular Data Stream (TDS) protocol depends upon a network transport
connection being established prior to a TDS conversation occurring (the
choice of transport protocol is not important to TDS).

TDS depends on Transport Layer Security (TLS)/Secure Socket Layer (SSL)
for network channel encryption. In the TDS 7.x version family, TLS/SSL
is optional and the negotiation of the encryption setting between the
client and server and the initial TLS/SSL handshake are handled in the
TDS layer.

Introduced in the TDS 8.0 version, TLS is mandatory and is established
in the lower layer before TDS begins functioning.

If the [**Multiple Active Result Sets
(MARS)**](#gt_762fe1e3-0979-4402-b963-1e9150de133d) feature
[\[MSDN-MARS\]](https://go.microsoft.com/fwlink/?LinkId=98459) is
enabled, the [**Session Multiplex Protocol
(SMP)**](#gt_f70f98cc-c555-4a40-9509-bc1da4021211)
[\[MC-SMP\]](%5bMC-SMP%5d.pdf#Section_04c8edde371d4af5bb33a39b3948f0af)
is required.

This relationship is illustrated in the following figure.

![Protocol
relationship](media/image2.bin "Protocol relationship"){width="5.5in"
height="2.8583333333333334in"}

Figure 2: Protocol relationship

## Prerequisites/Preconditions

This protocol can be used after the client has discovered the server and
established a network transport connection for use with TDS.

In the TDS 7.x version family, no security association is assumed to
have been established at the lower layer before TDS begins functioning.
In the TDS 8.0 version, such a security association is assumed.

For [**Security Support Provider Interface
(SSPI)**](#gt_fb216516-748b-4873-8bdd-64c5f4da9920) [\[SSPI\]](https://go.microsoft.com/fwlink/?LinkId=90536)
authentication to be used, SSPI support needs to be available on both
the client and server machines. For channel encryption to be used,
TLS/SSL support needs to be present on both client and server machines,
and a certificate suitable for encryption has to be deployed on the
server machine.

For [**federated
authentication**](#gt_5ae22a0e-5ff4-441b-80d4-224ef4dd4d19) to be used,
a library that provides federated authentication support or an
equivalent needs to be present on the server, and the client needs to be
able to generate a token for federated authentication.

## Applicability Statement

The TDS protocol is appropriate for use to facilitate request/response
communications between an application and a database server in all
scenarios where network or local connectivity is available.

## Versioning and Capability Negotiation

This protocol includes versioning issues in the following areas.

-   **Supported Transports:** This protocol can be implemented on top of
    any network transport protocol as discussed in section
    [2.1](#Section_fd30432f71b2488cb30f19737d76d970).

-   **Protocol Versions:** The TDS protocol supports the TDS 7.x version
    family (which is composed of explicit versions TDS 7.0, TDS 7.1, TDS
    7.2, TDS 7.3, and TDS 7.4) and the TDS 8.0 explicit version.

> In TDS 7.x, the explicit version is negotiated as part of the LOGIN7
> message [**data stream**](#gt_151643ce-fb5e-460e-8bdf-dc10bbd1950e),
> as described in section
> [2.2.6.4](#Section_773a62b6ee894c029e5e344882630aac).
>
> In TDS 8.0, the explicit version has to be identified from the TLS
> handshake by using an Application-Layer Protocol Negotiation (ALPN)
> TLS extension
> [\[RFC7301\]](https://go.microsoft.com/fwlink/?LinkId=513846). If ALPN
> is not present, the server has to assume the TDS 8.0 version has been
> sent. Version information sent in the LOGIN7 message data that is
> later in the flow does not have to be specified and ought to be
> ignored by both the client and server.
>
> Aspects of later versions of the TDS protocol that do not apply to
> earlier versions are identified in the text.
>
> **Note**  After a protocol feature is introduced, subsequent versions
> of the TDS protocol support that feature until that feature is
> removed.

-   **Security and Authentication Methods:** The TDS protocol supports
    [**SQL Server User Authentication
    (SQLAUTH)**](#gt_1a186995-43e5-4e46-896a-bad208ec2551). The TDS
    protocol also supports SSPI authentication and indirectly supports
    any authentication mechanism that SSPI supports. The use of SSPI in
    TDS is defined in sections 2.2.6.4 and
    [3.2.5.2](#Section_cc823ca848674387819dcc5c19da5732). The TDS
    protocol also supports [**federated
    authentication**](#gt_5ae22a0e-5ff4-441b-80d4-224ef4dd4d19). The use
    of federated authentication in TDS is defined in sections 2.2.6.4
    and [3.2.5](#Section_bd3a16dd75b64546933059c3ef44d50e).

-   **Localization:** Localization-dependent protocol behavior is
    specified in sections
    [2.2.5.1.2](#Section_3d29e8dc218a42c69ba4947ebca9fd7e) and
    [2.2.5.6](#Section_cbe9c510eae64b1f9893a098944d430a).

-   **Capability Negotiation:** This protocol does explicit capability
    negotiation as specified in this section.

In general, the TDS protocol does not provide facilities for capability
negotiation because the complete set of supported features is fixed for
each version of the protocol. Certain features such as authentication
type are not usually negotiated but rather are requested by the client.
However, the protocol supports negotiation for the following two
features:

-   **Channel encryption:** In TDS 7.x, the encryption behavior that is
    used for the [**TDS
    session**](#gt_276cd76b-c0a2-4f7c-8529-ad0d60aa9592) is negotiated
    in the initial messages exchanged by the client and the server. In
    TDS 8.0, encryption is mandatory and is established prior to initial
    messaging by the client and the server.

-   **Authentication mechanism for integrated authentication
    identities:** The authentication mechanism that is used for the TDS
    session is negotiated in the initial messages exchanged by the
    client and the server.

For more details about encryption behavior in TDS 7.x and about how the
client and server negotiate between SSPI authentication and federated
authentication, see the PRELOGIN description in section
[2.2.6.5](#Section_60f5640801884cd58b9025c6f2423868).

Note that the cipher suite for TLS/SSL
[\[RFC2246\]](https://go.microsoft.com/fwlink/?LinkId=90324)
[\[RFC5246\]](https://go.microsoft.com/fwlink/?LinkId=129803)
[\[RFC8446\]](https://go.microsoft.com/fwlink/?linkid=2147431)
[\[RFC6101\]](https://go.microsoft.com/fwlink/?LinkId=509953), the
authentication mechanism for SSPI
[\[SSPI\]](https://go.microsoft.com/fwlink/?LinkId=90536), and federated
authentication are negotiated outside the influence of TDS.

## Vendor-Extensible Fields

None.

## Standards Assignments

The TDS 7.x and TDS 8.0 protocols use the following assignment.

  -------------------------------------------------------------------------------------------------------------------------
  Parameter                              TCP port value     Reference
  -------------------------------------- ------------------ ---------------------------------------------------------------
  Default SQL Server instance TCP port   1433               [\[IANAPORT\]](https://go.microsoft.com/fwlink/?LinkId=89888)

  -------------------------------------------------------------------------------------------------------------------------

TDS 8.0 also uses the following ALPN identification sequence to identify
the TDS protocol.

  -----------------------------------------------------------------------
  Parameter      Identification Sequence                   Reference
  -------------- ----------------------------------------- --------------
  tds/8.0        0x74 0x64 0x73 0x2f 0x38 0x2e 0x30        \[MS-TDS\]

  -----------------------------------------------------------------------

**Note** This identification sequence has been requested and is in the
process of being registered with the Internet Assigned Numbers Authority
(IANA). This note will be removed when registration is completed.

