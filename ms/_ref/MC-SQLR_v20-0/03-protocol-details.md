# Protocol Details

This section describes the important elements of the client software and
the server software necessary to support the SQL Server Resolution
Protocol.

As described in section
[1.3](#Section_5d3c0525bcfb44ad85b3143cbeb9494f), the SQL Server
Resolution Protocol is an application-level protocol that is used to
query information regarding database instances on one or more servers.
The protocol defines a limited set of messages through which the client
can make a request to the server or servers. The SQL Server Resolution
Protocol involves a single request and a single response from one or
more servers. The response contains instance-specific values provided by
the higher layer.

## Server Details

The following state machine diagram describes the server side of the SQL
Server Resolution Protocol.

![SQL Server Resolution Protocol server state
machine](media/image4.bin "SQL Server Resolution Protocol server state machine"){width="5.554166666666666in"
height="2.2111111111111112in"}

Figure 4: SQL Server Resolution Protocol server state machine

### Abstract Data Model

None.

### Timers

None.

### Initialization

The SQL Server Resolution Protocol does not perform any initialization
on the server side. A UDP socket that is listening on port 1434 is
assumed to have been created by the higher layer.

### Higher-Layer Triggered Events

None.

### Message Processing Events and Sequencing Rules

Because the SQL Server Resolution Protocol provides a single response
per server for each client request, no sequencing issues occur with this
protocol.

#### Initial State

In the \"Initial State\", the initialization found in section
[3.1.3](#Section_62234af4161547b9951d4414996a41f9) is assumed to have
taken place. Upon success, the server MUST immediately enter the
\"Waiting For Request From Client\" state.

#### Waiting For Request From Client

In the \"Waiting For Request From Client\" state, the server listens on
UDP port 1434 for an incoming request. If the request is valid and
understood, the server immediately sends an
[SVR_RESP](#Section_2e1560c9509740239f5e72b9ff1ec3b1) response back to
the client. The data content of the response depends on the request
type.

-   For [CLNT_BCAST_EX](#Section_a3035afac2684699b8fd4f351e5c8e9e) and
    [CLNT_UCAST_EX](#Section_ee0e41b0204f4a95b8bd5783a7c72cb2), the
    server returns information about all available instances. The
    information about all available instances is provided by the higher
    layer.

-   For [CLNT_UCAST_INST](#Section_c97b04b5d80f4d3e919583bbfe246639),
    the server returns information about the specified instance only (if
    available). The information about the specified instance is provided
    by the higher layer.

-   For [CLNT_UCAST_DAC](#Section_20ebabbf46644f36bee04e3676a7aecd), the
    server returns information about the [**dedicated administrator
    connection (DAC)**](#gt_d50a91b6-9599-4d29-bad9-83fd1f6e6bf6) only.

The response SHOULD include information for a particular protocol as
long as the aggregate information for the instance fits within the 1,024
bytes limit. If the information for a protocol would cause the total
information for all protocols to exceed 1,024 bytes---for example,
trying to add a 500-byte pipe name when 800 bytes of response data have
already been collected---the information for this protocol SHOULD not be
sent. The information for the next protocol (if any) SHOULD be included
in the response (assuming that it does not cause the response to exceed
the 1,024 bytes limit). Furthermore, the server SHOULD NOT include a
protocol and its information if no valid information is available. For
example, if the [**TCP**](#gt_b08d36f6-b5c6-4ce4-8d2d-6f2ab75ea4cb) port
is invalid, TCP would not be included in the response. The SQL Server
Resolution Protocol SHOULD NOT verify the length or content of the
**PIPENAME** field, which is provided by the higher layer. It is the
upper layer\'s responsibility to ensure that **PIPENAME** conforms to
the specification of a valid pipe name
[\[MSDN-NP\]](https://go.microsoft.com/fwlink/?LinkId=90247).

If the request is received on an
[**IPv4**](#gt_0f25c9b5-dc73-4c3e-9433-f09d1f62ea8e) socket, the
response provides the instance\'s port that is associated with an IPv4
address, and likewise for
[**IPv6**](#gt_64c29bb6-c8b2-4281-9f3a-c1eb5d2288aa).

If the request is not valid, not understood, or if there is no instance
for which it can send back
[**endpoint**](#gt_b91c1e27-e8e0-499b-8c65-738006af72ee) information,
the server MUST ignore the request. The server MUST then enter the
\"Waiting For Request From Client\" state.

### Timer Events

None.

### Other Local Events

None.

## Client Details

The following state machine diagram describes the client side of the SQL
Server Resolution Protocol.

![SQL Server Resolution Protocol client state
machine](media/image5.bin "SQL Server Resolution Protocol client state machine"){width="5.547916666666667in"
height="4.175in"}

Figure 5: SQL Server Resolution Protocol client state machine

### Abstract Data Model

This section describes a conceptual model of possible data organization
that an implementation maintains to participate in this protocol. The
described organization is provided to facilitate the explanation of how
the protocol behaves. This document does not mandate that
implementations adhere to this model provided that their external
behavior is consistent with that described in this document.

A SQL Server Resolution Protocol client does not need to maintain any
state data except for the knowledge of the request sent to the server.

### Timers

The SQL Server Resolution Protocol client MUST implement a timer for the
amount of time to wait for an
[SVR_RESP](#Section_2e1560c9509740239f5e72b9ff1ec3b1) message from the
server when a
[CLNT_UCAST_INST](#Section_c97b04b5d80f4d3e919583bbfe246639) or
[CLNT_UCAST_DAC](#Section_20ebabbf46644f36bee04e3676a7aecd) request is
sent. The timer mechanism that is used is implementation-specific but
SHOULD[\<2\>](\l) have a time-out value of 1 second.

The SQL Server Resolution Protocol client MUST implement a timer for the
amount of time to wait for SVR_RESP messages from servers in the network
after a [CLNT_UCAST_EX](#Section_ee0e41b0204f4a95b8bd5783a7c72cb2) or
[CLNT_BCAST_EX](#Section_a3035afac2684699b8fd4f351e5c8e9e) request is
sent. The timer mechanism that is used is
implementation-specific.[\<3\>](\l)

### Initialization

The SQL Server Resolution Protocol does not perform any initialization
on the client side. A UDP socket is assumed to be created prior to
requesting that a SQL Server Resolution Protocol request be sent.

### Higher-Layer Triggered Events

The SQL Server Resolution Protocol client implementation MUST support
the following event from the higher layer:

-   Send client request to server or servers. The type of request
    dictates if the message is sent to a specific machine or
    [**broadcast**](#gt_7f275cc2-a1c5-47d7-83ae-9a84178f2481)/[**multicast**](#gt_70b74a6e-db1d-4648-bedd-5a524dfe6396)
    to the network. The higher layer has to have already created a UDP
    socket prior to triggering the SQL Server Resolution Protocol client
    to send a request message. After the message is sent, the timer MUST
    be started.

The higher layer is responsible for closing the UDP socket after the
response is received or a time-out situation has occurred.

### Message Processing Events and Sequencing Rules

Because the SQL Server Resolution Protocol provides a single response
per server for each client request, no sequencing issues occur with this
protocol.

#### Begin

In the \"Begin\" state, the client awaits a request from the higher
layer. After a request from the higher layer is made, the client sends a
request to the server or servers. The request type determines what state
the client enters next.

If the client sends a
[CLNT_UCAST_INST](#Section_c97b04b5d80f4d3e919583bbfe246639) or
[CLNT_UCAST_DAC](#Section_20ebabbf46644f36bee04e3676a7aecd) request to
the server, the client MUST then enter the \"Client Waits For Response
From Server\" state. If the client sends a
[CLNT_UCAST_EX](#Section_ee0e41b0204f4a95b8bd5783a7c72cb2) or
[CLNT_BCAST_EX](#Section_a3035afac2684699b8fd4f351e5c8e9e) request to a
server or servers, the client MUST then enter the \"Client Waits For
Response From Server(s)\" state.

#### Client Waits For Response From Server

In the \"Client Waits For Response From Server\" state, the client waits
either for a time-out to occur or for the results of a request to
return. As soon as either occurs, the client MUST enter the \"Waiting
Completed\" state.

The details of the timer are outlined in section
[3.2.2](#Section_eb2f6036e02b4df59c04e7bd64ecba15).

#### Client Waits For Response From Server(s)

In the \"Client Waits For Response From Server(s)\" state, the client
waits for responses up until the time-out expires. If the client
receives an invalid message, it MUST ignore the message and continue
listening for additional responses until the time-out period elapses.
The client then MUST enter the \"Waiting Completed\" state.

For purposes of this section, invalid messages are defined as messages
that do not follow the prescribed message format that is outlined in
section [2](#Section_6e1a7cdfb5254830ba1af2fa15e8283b) or defined as an
unexpected [SVR_RESP](#Section_2e1560c9509740239f5e72b9ff1ec3b1)
messages type.

The details of the timer are outlined in section
[3.2.2](#Section_eb2f6036e02b4df59c04e7bd64ecba15).

#### Waiting Completed

The client\'s actions upon entering the \"Waiting Completed\" state are
determined by the client message type to which the server is responding.

[CLNT_UCAST_DAC](#Section_20ebabbf46644f36bee04e3676a7aecd): The client
MUST notify the higher layer of the valid and properly formatted
[SVR_RESP](#Section_2e1560c9509740239f5e72b9ff1ec3b1) (DAC) messages or
notify the higher layer if it received an invalid message. After this,
the client MUST enter the \"End\" state.

[CLNT_BCAST_EX](#Section_a3035afac2684699b8fd4f351e5c8e9e): The client
MUST notify the higher layer of the valid and properly formatted
SVR_RESP messages. The client SHOULD buffer all responses until the
timer has timed out. It MUST then pass the information to the higher
layer. The client MUST ignore the invalid messages and does not notify
the higher layer regarding these messages. After this, the client MUST
enter the \"End\" state.

[CLNT_UCAST_EX](#Section_ee0e41b0204f4a95b8bd5783a7c72cb2): The client
MUST notify the higher layer of the valid and properly formatted
SVR_RESP messages or notify the higher layer if it received an invalid
message. After this, the client MUST enter the \"End\" state. Although
the server's maximum RESP_DATA size for a SVR_RESP message is 65,535
bytes, the client MAY consider a SVR_RESP message improperly formatted
if the RESP_DATA field exceeds a value set by the client that is smaller
than 65,535 bytes.[\<4\>](\l)

[CLNT_UCAST_INST](#Section_c97b04b5d80f4d3e919583bbfe246639): The client
MUST notify the higher layer of the valid and properly formatted
SVR_RESP messages or notify the higher layer if it received an invalid
message. After this, the client MUST enter the \"End\" state. A SRV_RESP
message SHOULD NOT be considered properly formatted if the cumulative
length of the parameters of any transport protocol included in the
response is more than 255 bytes (in other words, a SRV_RESP message
SHOULD NOT be considered properly formatted if it contains an
NP_PARAMETERS, TCP_PARAMETERS, VIA_PARAMETERS, RPC_PARAMETERS,
SPX_PARAMETERS, ADSP_PARAMETERS, or BV_PARAMETERS token whose length is
more than 255 bytes).

#### End

The client has completed the request.

### Timer Events

When the timer for the response to a
[**broadcast**](#gt_7f275cc2-a1c5-47d7-83ae-9a84178f2481) or
[**multicast**](#gt_70b74a6e-db1d-4648-bedd-5a524dfe6396) request
expires, the client MUST enter the \"Waiting Completed\" state.

### Other Local Events

None.

