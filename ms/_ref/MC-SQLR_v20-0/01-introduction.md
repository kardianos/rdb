# Introduction

The SQL Server Resolution Protocol is an application-layer
request/response protocol that facilitates connectivity to a database
server. This protocol provides for the following:

-   Communication
    [**endpoint**](#gt_b91c1e27-e8e0-499b-8c65-738006af72ee)
    information; for example, the TCP port for connecting to a
    particular instance of the database server on a machine.

-   Database instance enumeration.

Sections 1.5, 1.8, 1.9, 2, and 3 of this specification are normative.
All other sections and examples in this specification are informative.

## Glossary

This document uses the following terms:

> []{#gt_7f275cc2-a1c5-47d7-83ae-9a84178f2481 .anchor}**broadcast**: A
> style of resource location or data transmission in which a client
> makes a request to all parties on a network simultaneously (a
> one-to-many communication). Also, a mode of resource location that
> does not use a name service.
>
> []{#gt_fe1a7538-7138-4669-8522-9e53c5bf8fe7 .anchor}**database server
> discovery service**: A service that allows applications to discover
> the existence of database instances.
>
> []{#gt_d50a91b6-9599-4d29-bad9-83fd1f6e6bf6 .anchor}**dedicated
> administrator connection (DAC)**: A special TCP
> [**endpoint**](#gt_b91c1e27-e8e0-499b-8c65-738006af72ee) that was
> introduced in Microsoft SQL Server 2005. DAC provides a special
> diagnostic connection for administrators when standard connections to
> the server are not possible.
>
> []{#gt_b91c1e27-e8e0-499b-8c65-738006af72ee .anchor}**endpoint**: A
> client that is on a network and is requesting access to a network
> access server (NAS).
>
> []{#gt_0f25c9b5-dc73-4c3e-9433-f09d1f62ea8e .anchor}**Internet
> Protocol version 4 (IPv4)**: An Internet protocol that has 32-bit
> source and destination addresses. IPv4 is the predecessor of IPv6.
>
> []{#gt_64c29bb6-c8b2-4281-9f3a-c1eb5d2288aa .anchor}**Internet
> Protocol version 6 (IPv6)**: A revised version of the Internet
> Protocol (IP) designed to address growth on the Internet. Improvements
> include a 128-bit IP address size, expanded routing capabilities, and
> support for authentication and privacy.
>
> []{#gt_079478cb-f4c5-4ce5-b72b-2144da5d2ce7 .anchor}**little-endian**:
> Multiple-byte values that are byte-ordered with the least significant
> byte stored in the memory location with the lowest address.
>
> []{#gt_70b74a6e-db1d-4648-bedd-5a524dfe6396 .anchor}**multicast**: A
> style of resource location or a data transmission in which a client
> makes a request to specific parties on a network simultaneously.
>
> []{#gt_34f1dfa8-b1df-4d77-aa6e-d777422f9dca .anchor}**named pipe**: A
> named, one-way, or duplex pipe for communication between a pipe server
> and one or more pipe clients.
>
> []{#gt_b08d36f6-b5c6-4ce4-8d2d-6f2ab75ea4cb .anchor}**Transmission
> Control Protocol (TCP)**: A protocol used with the Internet Protocol
> (IP) to send data in the form of message units between computers over
> the Internet. TCP handles keeping track of the individual units of
> data (called packets) that a message is divided into for efficient
> routing through the Internet.
>
> []{#gt_e73c7149-240a-4fad-8a27-5c6b7fdc956a .anchor}**unicast**: A
> style of resource location or a data transmission in which a client
> makes a request to a single party.
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

\[MS-UCODEREF\] Microsoft Corporation, \"[Windows Protocols Unicode
Reference](%5bMS-UCODEREF%5d.pdf#Section_4a045e08fc294f22baf416f38c2825fb)\".

\[RFC2119\] Bradner, S., \"Key words for use in RFCs to Indicate
Requirement Levels\", BCP 14, RFC 2119, March 1997,
[https://www.rfc-editor.org/info/rfc2119](https://go.microsoft.com/fwlink/?LinkId=90317)

\[RFC2460\] Deering, S., and Hinden, R., \"Internet Protocol, Version 6
(IPv6) Specification\", RFC 2460, December 1998,
[https://www.rfc-editor.org/info/rfc2460](https://go.microsoft.com/fwlink/?LinkId=90357)

\[RFC4234\] Crocker, D., Ed., and Overell, P., \"Augmented BNF for
Syntax Specifications: ABNF\", RFC 4234, October 2005,
[https://www.rfc-editor.org/info/rfc4234](https://go.microsoft.com/fwlink/?LinkId=90462)

\[RFC768\] Postel, J., \"User Datagram Protocol\", STD 6, RFC 768,
August 1980,
[https://www.rfc-editor.org/info/rfc768](https://go.microsoft.com/fwlink/?LinkId=90490)

\[RFC791\] Postel, J., Ed., \"Internet Protocol: DARPA Internet Program
Protocol Specification\", RFC 791, September 1981,
[https://www.rfc-editor.org/info/rfc791](https://go.microsoft.com/fwlink/?LinkId=392659)

\[RFC793\] Postel, J., Ed., \"Transmission Control Protocol: DARPA
Internet Program Protocol Specification\", RFC 793, September 1981,
[https://www.rfc-editor.org/info/rfc793](https://go.microsoft.com/fwlink/?LinkId=150872)

\[VIA2002\] Cameron, D., and Regnier, G., \"The Virtual Interface
Architecture\", Intel Press, 2002, ISBN:0971288704.

### Informative References

\[MSDN-CS\] Microsoft Corporation, \"Character Sets\",
[https://learn.microsoft.com/en-us/windows/desktop/Intl/character-sets](https://go.microsoft.com/fwlink/?LinkId=90692)

\[MSDN-DAC\] Microsoft Corporation, \"Diagnostic Connection for Database
Administrators\",
[https://learn.microsoft.com/en-us/sql/database-engine/configure-windows/diagnostic-connection-for-database-administrators](https://go.microsoft.com/fwlink/?LinkId=95068)

\[MSDN-NamedPipes\] Microsoft Corporation, \"Creating a Valid Connection
String Using Named Pipes\",
[https://learn.microsoft.com/en-us/previous-versions/sql/sql-server-2008-r2/ms189307(v=sql.105)](https://go.microsoft.com/fwlink/?LinkId=127839)

\[MSDN-NP\] Microsoft Corporation, \"Named Pipes\",
[https://learn.microsoft.com/en-us/windows/desktop/ipc/named-pipes](https://go.microsoft.com/fwlink/?LinkId=90247)

## Overview

The SQL Server Resolution Protocol is a simple application-level
protocol that is used for the transfer of requests and responses between
clients and [**database server discovery
services**](#gt_fe1a7538-7138-4669-8522-9e53c5bf8fe7). In such a system,
the client either (i) sends a single request to a specific machine and
expects a single response, or (ii)
[**broadcasts**](#gt_7f275cc2-a1c5-47d7-83ae-9a84178f2481) or
[**multicasts**](#gt_70b74a6e-db1d-4648-bedd-5a524dfe6396) a request to
the network and expects zero or more responses from different discovery
services on the network. The first case is used for the purpose of
determining the communication
[**endpoint**](#gt_b91c1e27-e8e0-499b-8c65-738006af72ee) information of
a particular database instance, whereas the second case is used for
enumeration of database instances in the network and to obtain the
endpoint information of each instance.

The SQL Server Resolution Protocol does not include any facilities for
authentication, protection of data, or reliability. The SQL Server
Resolution Protocol is always implemented on top of the UDP Transport
Protocol [\[RFC768\]](https://go.microsoft.com/fwlink/?LinkId=90490).

In the case of endpoint determination for a single instance, the
following diagram depicts a typical flow of communication.

![Communication flow for single-instance endpoint
discovery](media/image1.bin "Communication flow for single-instance endpoint discovery"){width="5.547916666666667in"
height="2.2527777777777778in"}

Figure 1: Communication flow for single-instance endpoint discovery

Conversely, in the case of a broadcast/multicast request, the following
diagram applies.

![Communication flow for multiple-instance endpoint
discovery](media/image2.bin "Communication flow for multiple-instance endpoint discovery"){width="5.547916666666667in"
height="2.957638888888889in"}

Figure 2: Communication flow for multiple-instance endpoint discovery

In the case of a broadcast or multicast request, the client does not
necessarily know the number of responses that it can expect. As a
result, it is reasonable for the client to enforce a time limitation
during which it waits for responses. Because some servers might not
respond quickly enough or might not receive the request (highly
dependent on network topology), the broadcast/multicast request for
multiple-instance endpoint information is considered nondeterministic.

## Relationship to Other Protocols

The SQL Server Resolution Protocol (SSRP) depends on the UDP Transport
Protocol to communicate with the database server machine or to
[**broadcast**](#gt_7f275cc2-a1c5-47d7-83ae-9a84178f2481)/[**multicast**](#gt_70b74a6e-db1d-4648-bedd-5a524dfe6396)
its request to the network. The types of addresses used can differ based
on the underlying IP protocol version as described in section
[2.1](#Section_f5d99036b1b64b8490a1933aafac5a1c). For details about
[**IPv4**](#gt_0f25c9b5-dc73-4c3e-9433-f09d1f62ea8e), see
[\[RFC791\]](https://go.microsoft.com/fwlink/?LinkId=392659). For
details about [**IPv6**](#gt_64c29bb6-c8b2-4281-9f3a-c1eb5d2288aa), see
[\[RFC2460\]](https://go.microsoft.com/fwlink/?LinkId=90357).

![Protocol
relationship](media/image3.bin "Protocol relationship"){width="1.698611111111111in"
height="2.0298600174978128in"}

Figure 3: Protocol relationship

## Prerequisites/Preconditions

Unprohibited access to UDP port 1434 is required.

## Applicability Statement

The SQL Server Resolution Protocol is appropriate for use to facilitate
retrieval of database
[**endpoint**](#gt_b91c1e27-e8e0-499b-8c65-738006af72ee) information or
for database instance enumeration in all scenarios where network or
local connectivity is available.

## Versioning and Capability Negotiation

This document covers versioning issues in the following areas:

-   Supported transports: This protocol is implemented on top of UDP, as
    discussed in section
    [2.1](#Section_f5d99036b1b64b8490a1933aafac5a1c).

-   Protocol versions: The SQL Server Resolution Protocol supports the
    following explicit dialect: \"SSRP 1.0\", as defined in section
    [2.2](#Section_def8dec45c2b4bfd9531655de3f74954).

-   Security and authentication methods: The SQL Server Resolution
    Protocol does not provide or support any security or authentication
    methods.

-   Localization: Localization-dependent protocol behavior is specified
    in sections 2.2 and
    [3.2.5](#Section_4cacbe76f0374fd3856669e1537973a1).

-   Capability negotiation: The SQL Server Resolution Protocol does not
    support negotiation of the dialect to use. Instead, an
    implementation can be configured with the dialect to use, as
    described later in this section.

No version or capability negotiation is supported by the SQL Server
Resolution Protocol. For example, the client sends a request message to
the server with the expectation that the server will understand the
message and send back a response. If the server does not understand the
message, the server ignores the request and does not send a response
back to the client.

## Vendor-Extensible Fields

None.

## Standards Assignments

  -----------------------------------------------------------------------------------
  Parameter                 Value      Reference
  ------------------------- ---------- ----------------------------------------------
  Microsoft-SQL-Monitor     1434/UDP   http://www.iana.com/assignments/port-numbers
  (ms-sql-m)                           

  -----------------------------------------------------------------------------------

The client always sends its request to UDP port 1434 of the server or
servers.

