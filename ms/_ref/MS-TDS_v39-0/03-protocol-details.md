# Protocol Details

This section describes the important elements of the client software and
the server software necessary to support the TDS protocol.

## Common Details

As described in section
[1.3](#Section_893fcc7e8a394b3c815a773b7b982c50), TDS is an
application-level protocol that is used for the transfer of requests and
responses between clients and database server systems. The protocol
defines a limited set of messages through which the client can make a
request to the server. The TDS server is message-oriented. Once a
connection has been established between the client and server, a
complete message is sent from client to server. Following this, a
complete response is sent from server to client (with the possible
exception of when the client aborts the request), and the server then
waits for the next request. Other than this Post-Login state, the other
states defined by the TDS protocol are (i) pre-authentication
(Pre-Login), (ii) authentication (Login), and (iii) when the client
sends an attention message (Attention). These are expanded upon in
subsequent sections.

### Abstract Data Model

See sections [3.2.1](#Section_0e7007a13da74323914828d89ec2ac48) and
[3.3.1](#Section_1457f2976d7d4b47a4bef35d273b8cae) for the abstract data
model of the client and server, respectively.

### Timers

See section [3.2.2](#Section_8313b90eeee347e1844173b3b5df1e39) for a
description of the client timer used and section
[3.3.2](#Section_25a86907cd7a44a39c923b9a9c0e127f) for a description of
the server timer used.

### Initialization

None.

### Higher-Layer Triggered Events

For information about higher-layer triggered events, see section
[3.2.4](#Section_e7af6058672941438815f82a475f9ba6) for a TDS client and
section [3.3.4](#Section_6af826c5a9574438a7c0b3bbd309c69a) for a TDS
server.

### Message Processing Events and Sequencing Rules

The following series of sequence diagrams illustrate the messages that
can be exchanged between client and server. See sections
[3.2.5](#Section_bd3a16dd75b64546933059c3ef44d50e) and
[3.3.5](#Section_8d3c50da023f4bc5959c60f38ed22825) for specific client
and server details regarding message processing events and sequencing
rules.

![Pre-login to post-login sequence that is used in TDS
7.x](media/image3.bin "Pre-login to post-login sequence that is used in TDS 7.x"){width="6.291666666666667in"
height="7.025in"}

Figure 3: Pre-login to post-login sequence that is used in TDS 7.x

![Pre-login to post-login sequence that is used in TDS
8.0](media/image4.bin "Pre-login to post-login sequence that is used in TDS 8.0"){width="5.416666666666667in"
height="5.566666666666666in"}

Figure 4: Pre-login to post-login sequence that is used in TDS 8.0

![Pre-login to post-login sequence with federated authentication that
uses a client library that requires additional information from a server
to generate a federated authentication token that is used in TDS
7.x](media/image5.bin "Pre-login to post-login sequence with federated authentication that uses a client library that requires additional information from a server to generate a federated authentication token that is used in TDS 7.x"){width="6.383333333333334in"
height="7.533333333333333in"}

Figure 5: Pre-login to post-login sequence with federated authentication
that uses a client library that requires additional information from a
server to generate a federated authentication token that is used in TDS
7.x

![Pre-login to post-login sequence with federated authentication that
uses a client library that requires additional information from a server
to generate a federated authentication token that is used in TDS
8.0](media/image6.bin "Pre-login to post-login sequence with federated authentication that uses a client library that requires additional information from a server to generate a federated authentication token that is used in TDS 8.0"){width="6.366666666666666in"
height="6.541666666666667in"}

Figure 6: Pre-login to post-login sequence with federated authentication
that uses a client library that requires additional information from a
server to generate a federated authentication token that is used in TDS
8.0

![SQL command and RPC
sequence](media/image7.bin "SQL command and RPC sequence"){width="4.975in"
height="6.95in"}

Figure 7: SQL command and RPC sequence

![Transaction manager request
sequence](media/image8.bin "Transaction manager request sequence"){width="5.65in"
height="4.616666666666666in"}

Figure 8: Transaction manager request sequence

![Bulk insert
sequence](media/image9.bin "Bulk insert sequence"){width="5.541666666666667in"
height="6.658333333333333in"}

Figure 9: Bulk insert sequence

### Timer Events

See sections [3.2.6](#Section_030af85119fd4a9ca7df05b637f7d888) and
[3.3.6](#Section_c0d2d1a29c3d4eb68a460a7dab3b3f56) for the timer events
of the client and server, respectively.

### Other Local Events

A TDS session is tied to the underlying established network protocol
session. As such, loss or termination of a network connection is
equivalent to immediate termination of a TDS session.

See sections [3.2.7](#Section_97d920799ea14a7ba915fa7e8045d88b) and
[3.3.7](#Section_388d89cc00a34f57bc22624e251e68f3) for the other local
events of the client and server, respectively.

## Client Details

The following state machine diagrams describe TDS on the client side.

![TDS client state machine that is used in TDS
7.x](media/image10.bin "TDS client state machine that is used in TDS 7.x"){width="5.741666666666666in"
height="6.9in"}

Figure 10: TDS client state machine that is used in TDS 7.x

![TDS client state machine that is used in TDS
8.0](media/image11.bin "TDS client state machine that is used in TDS 8.0"){width="5.883333333333334in"
height="7.183333333333334in"}

Figure 11: TDS client state machine that is used in TDS 8.0

### Abstract Data Model

This section describes a conceptual model of data organization that an
implementation maintains to participate in this protocol. The described
organization is provided to facilitate the explanation of how the
protocol behaves. This document does not mandate that implementations
adhere to this model as long as their external behavior is consistent
with that described in this document.

A TDS client SHOULD maintain the following states:

-   [Sent Initial TLS Negotiation Packet
    State](#Section_fc8fa46974004e328d023d54f166e7fb) (applies to only
    TDS 8.0)

-   [Sent Initial PRELOGIN Packet
    State](#Section_cc823ca848674387819dcc5c19da5732)

-   [Sent TLS/SSL Negotiation Packet
    State](#Section_d62e225bd8654ccc8f73de1ef49e30d4) (applies to only
    TDS 7.x)

-   [Sent LOGIN7 Record with Complete Authentication Token
    State](#Section_ab7a745416a949ef8f208a0de2b31fbb)

-   [Sent LOGIN7 Record with SPNEGO Packet
    State](#Section_c373a90b113d4c5f826ad410365ee883)

-   [Sent LOGIN7 Record with Federated Authentication Information
    Request State](#Section_c272e054279f4b09ac41f9ccf1d1ffde)

-   [Logged In State](#Section_e2e7293f55cf43eeb0282d6dea625db6)

-   [Sent Client Request
    State](#Section_c89701263cbc446f9275cc36ba62a2f4)

-   [Sent Attention State](#Section_61fb4f9d95ec41718d4bdac4eb16149f)

-   [Routing Completed State](#Section_52851226121146c8ac506b7facd08fd9)

-   [Final State](#Section_07a0c71e1a7b44558da47e43a09589bc)

### Timers

A TDS client SHOULD implement the following three timers:

-   Connection Timer. Controls the maximum time spent during the
    establishment of a TDS connection. The default value SHOULD be 15
    seconds. The implementation SHOULD allow the upper layer to specify
    a nondefault value, including an infinite value (for example, no
    timeout).

-   Client Request Timer. Controls the maximum time spent waiting for a
    query response from the server for a client request sent after the
    connection has been established. The default value is
    implementation-dependent. The implementation SHOULD allow the upper
    layer to specify a non-default value, including an infinite value
    (for example, no timeout).[\<70\>](\l)

-   Cancel Timer. Controls the maximum time spent waiting for a query
    cancellation acknowledgement after an Attention request is sent to
    the server. The default value is implementation-dependent. The
    implementation SHOULD allow the upper layer to specify a nondefault
    value, including an infinite value (for example, no
    timeout).[\<71\>](\l)

For all three timers, a client can implement a minimum timeout value
that is as short as required. If a TDS client implementation implements
any of the timers, it MUST implement their behavior according to this
specification.

A TDS client SHOULD request the transport to detect and indicate a
broken connection if the transport provides such mechanism. If the
transport used is TCP, it SHOULD use the TCP Keep-Alives
[\[RFC1122\]](https://go.microsoft.com/fwlink/?LinkId=112180) in order
to detect a nonresponding server in case infinite connection timeout or
infinite client request timeout is used. The default values of the TCP
Keep-Alive values set by a TDS client are 30 seconds of no activity
until the first keep-alive packet is sent and 1 second between when
successive keep-alive packets are sent if no acknowledgement is
received. The implementation SHOULD allow the upper layer to specify
other TCP keep-alive values.

### Initialization

None.

### Higher-Layer Triggered Events

A TDS client MUST support the following events from the upper layer:

-   Connection Open Request to establish a new TDS connection to a TDS
    server.

-   Client Request to send a query to a TDS server on an already
    established TDS connection. The Client Request is a request for one
    of four types of queries to be sent: SQL Command, Bulk Load,
    Transaction Manager Request, or an RPC.

In addition, it SHOULD support the following event from the upper layer:

-   Cancel Request to cancel a client request while waiting for a server
    response. For example, this enables the upper layer to cancel a
    long-running client request if the user/upper layer is no longer
    seeking the result, thus freeing up thus client and server
    resources. If a client implementation of the TDS protocol supports
    the Cancel Request event, it MUST handle it as described in this
    specification.

The processing and actions triggered by these events is described in the
remaining parts of this section.

When a TDS client receives a Connection Open Request from the upper
layer in the \"Initial State\" state of a TDS connection, it performs
the following actions:

-   If the TDS client implements the Connection Timer, it MUST start the
    Connection Timer if the connection timeout value is not infinite.

-   If there is upper-layer request
    [**MARS**](#gt_762fe1e3-0979-4402-b963-1e9150de133d) support, it
    MUST set the B_MARS byte in the PRELOGIN message to 0x01.

-   It MUST send a PRELOGIN message to the server by using the
    underlying transport protocol.

-   If the transport does not report an error, it MUST enter the \"Sent
    Initial PRELOGIN Packet\" state.

When a TDS client receives a Connection Open Request from the upper
layer in any state other than the \"Initial State\" state of a TDS
connection, it MUST indicate an error to the upper layer.

When a TDS client receives a Client Request from the upper layer in the
\"Logged In\" state, it MUST perform the following actions:

-   If the TDS client implements the Query Timer, it MUST start the
    Client Request Timer if the client request timeout value is not
    infinite.

-   If MARS is enabled, the client MUST keep track whether there is an
    outstanding active request. If this is the case, then the client
    MUST initiate a new SMP session, or else an existing SMP session MAY
    be used.

-   Send either SQL Command, Bulk Load, Transaction Manager Request, or
    a RPC message to the server. The message and its content MUST match
    the requested message from the Client Request. If MARS is enabled,
    the TDS message MUST be passed through to the SMP layer.

-   If the transport does not report an error, then enter the \"Sent
    Client Request\" state.

When a TDS client supporting the Cancel Request receives a Cancel
Request from the upper layer in the \"Sent Client Request\" state, it
MUST perform the following actions:

-   If the TDS client implements the Cancel Timer, it MUST start the
    Cancel Timer if the Attention request timeout value is not infinite.

-   Send an Attention message to the server. This indicates to the
    server that the client intends to abort the executing request. If
    MARS is enabled, the Attention message MUST be passed through to the
    SMP layer.

-   Enter the \"Sent Attention\" state.

### Message Processing Events and Sequencing Rules

The processing of messages received from a TDS server depends on the
message type and the current state of the TDS client. The rest of this
section describes the processing and actions to take on them. The
message type is determined from the TDS packet type and the token stream
inside the TDS packet payload, as described in section
[2.2.3](#Section_e5ea85201ea34a75a2a9c17e63e9ee19).

Whenever the TDS client enters either the \"Logged In\" state or the
\"Final State\" state, it MUST stop the Connection Timer (if implemented
and running), the Client Request Timer (if implemented and running), and
the Cancel Timer (if implemented and running).

Whenever a TDS client receives a structurally invalid TDS message, it
MUST close the underlying transport connection, indicate an error to the
upper layer, and enter the \"Final State\" state.

When a TDS client receives a [**table
response**](#gt_71dd1dd2-c167-49a8-a5f5-6b0df5c8b48a) (TDS packet type
%x04) from the server, it MUST behave as follows, according to the state
of the TDS client.

The corresponding action is taken when the client is in the following
states. In the following processing and actions, aspects that do not
apply to both TDS 7.x and TDS 8.0 are explicitly identified in the text.

#### Sent Initial TLS Negotiation Packet State

***Applies to only TDS 8.0***

If the response received from the server contains a structurally valid
TLS response that indicates a success, the TDS client MUST send a
PRELOGIN message to the server and enter the \"Sent Initial PRELOGIN
Packet State\" state.

If the response received from the server does not contain a structurally
valid TLS response, or if it contains a structurally valid response that
indicates an error, the TDS client MUST close the underlying transport
connection, indicate an error to the upper layer, and enter the \"Final
State\" state.

#### Sent Initial PRELOGIN Packet State

If the response contains a structurally valid PRELOGIN response
indicating a success, the TDS client MUST take action according to the
Encryption option and Authentication scheme:

-   If TLS was not established before TDS begins to function, as
    required in TDS 8.0, the encryption option MUST be handled as
    described in the \"Encryption\" subsection of section
    [2.2.6.5](#Section_60f5640801884cd58b9025c6f2423868) in the PRELOGIN
    message description.

-   If encryption was negotiated in TDS 7.x, the TDS client MUST
    initiate a TLS/SSL handshake, send to the server a TLS/SSL message
    obtained from the TLS/SSL layer encapsulated in TDS packet(s) of
    type PRELOGIN (0x12), and enter the \"Sent TLS/SSL Negotiation
    Packet\" state.

-   If encryption was not negotiated and the upper layer did not request
    full encryption, the TDS client MUST send to the server a Login
    message that contains the authentication scheme that is specified by
    the user and MUST enter one of the following three states, depending
    on the message sent:

    -   \"Sent LOGIN7 Record with Complete Authentication Token\" state,
        if a login message that contains either of the following was
        sent.

        -   Standard authentication.

        -   FEDAUTH FeatureExt that indicates a client library that does
            not need any additional information from the server for
            authentication.

    -   \"Sent LOGIN7 Record with
        [**SPNEGO**](#gt_bc2f6b5e-e5c0-408b-8f55-0350c24b9838) Packet\"
        state, if a Login message with SPNEGO authentication was sent.

    -   \"Sent LOGIN7 Record with Federated Authentication Information
        Request\" state, if a Login message with FEDAUTH FeatureExt that
        indicates a client library that needs additional information
        from the server for authentication was sent.

> The TDS specification does not prescribe the authentication protocol
> if SSPI [\[SSPI\]](https://go.microsoft.com/fwlink/?LinkId=90536)
> authentication is used. The current implementation of SSPI supports
> NTLM [\[MSDN-NTLM\]](https://go.microsoft.com/fwlink/?LinkId=145227)
> and Kerberos
> [\[RFC4120\]](https://go.microsoft.com/fwlink/?LinkId=90458).

-   If encryption was not negotiated and the upper layer requested full
    encryption, then the TDS client MUST close the underlying transport
    connection, indicate an error to the upper layer, and enter the
    \"Final State\" state.

-   If the response received from the server does not contain a
    structurally valid PRELOGIN response or it contains a structurally
    valid PRELOGIN response indicating an error, the TDS client MUST
    close the underlying transport connection, indicate an error to the
    upper layer, and enter the \"Final State\" state.

-   If NONCEOPT is specified in both the client PRELOGIN message and the
    server PRELOGIN message, the TDS client MUST maintain a state
    variable that includes the value of the NONCE that is sent to the
    server and a state variable that includes the value of the NONCE
    that is contained in the server's response.

#### Sent TLS/SSL Negotiation Packet State

***Applies to only TDS 7.x***

In TDS 8.0, because encryption is already established, the TDS state
machine MUST NOT enter this state. Otherwise, the TDS server closes the
underlying transport connection, indicates an error to the upper layer,
and enters the \"Final State\" state.

If the response contains a structurally valid TLS/SSL response message
(TDS packet type 0x12), the TDS client MUST pass the TLS/SSL message
contained in it to the TLS/SSL layer and MUST proceed as follows:

-   If the TLS/SSL layer indicates that further handshake is needed, the
    TDS client MUST send to the server the TLS/SSL message obtained from
    the TLS/SSL layer encapsulated in TDS packet(s) of type PRELOGIN
    (0x12).

-   If the TLS/SSL layer indicates successful completion of the TLS/SSL
    handshake, the TDS client MUST send a Login message to the server
    that contains the authentication scheme that is specified by the
    user. The TDS client then enters one of the following three states,
    depending on the message sent:

    -   \"Sent LOGIN7 Record with Complete Authentication Token\" state,
        if a Login message that contains either of the following was
        sent:

        -   Standard authentication.

        -   FEDAUTH FeatureId that indicates a client library that does
            not need any additional information from the server for
            authentication.

    -   The \"Sent LOGIN7 Record with
        [**SPNEGO**](#gt_bc2f6b5e-e5c0-408b-8f55-0350c24b9838) Packet\"
        state, if a Login message with SPNEGO authentication was sent.

    -   \"Sent LOGIN7 Record with Federated Authentication Information
        Request\" state, if a Login message with FEDAUTH FeatureExt that
        indicates a client library that needs additional information
        from server for authentication was sent.

> The TDS specification does not prescribe the authentication protocol
> if SSPI [\[SSPI\]](https://go.microsoft.com/fwlink/?LinkId=90536)
> authentication or federated authentication is used. The current
> implementation of SSPI supports NTLM
> [\[MSDN-NTLM\]](https://go.microsoft.com/fwlink/?LinkId=145227) and
> Kerberos [\[RFC4120\]](https://go.microsoft.com/fwlink/?LinkId=90458).

-   If login-only encryption was negotiated as described in section
    [2.2](#Section_a734db771cb247e4b61a0c9dbd8f960f) in the PRELOGIN
    message description, then the first TDS packet of the Login message
    MUST be encrypted using TLS/SSL and encapsulated in a TLS/SSL
    message. All other TDS packets sent or received MUST be in
    plaintext.

-   If full encryption was negotiated as described in section 2.2 in the
    PRELOGIN message description, then all subsequent TDS packets sent
    or received from this point on MUST be encrypted using TLS/SSL and
    encapsulated in a TLS/SSL message.

-   If the TLS/SSL layer indicates an error, the TDS client MUST close
    the underlying transport connection, indicate an error to the upper
    layer, and enter the \"Final State\" state.

If the response received from the server does not contain a structurally
valid TLS/SSL response or it contains a structurally valid response
indicating an error, the TDS client MUST close the underlying transport
connection, indicate an error to the upper layer, and enter the \"Final
State\" state.

#### Sent LOGIN7 Record with Complete Authentication Token State

If the response received from the server contains a structurally valid
Login response that indicates a successful login, and if the client used
[**federated authentication**](#gt_5ae22a0e-5ff4-441b-80d4-224ef4dd4d19)
to authenticate to the server, the client MUST read the Login response
stream to find the FEATUREEXTACK token and find the FEDAUTH FeatureId.
If the FEDAUTH FeatureId is not present, the TDS client MUST close the
underlying transport connection, indicate an error to the upper layer,
and enter the \"Final State\" state. If the FEDAUTH FeatureId is
present, the client\'s action is based on the bFedAuthLibrary as
follows:

-   When the bFedAuthLibrary is Live ID Compact Token, the client MUST
    use the session key from its federated authentication token to
    compute the HMAC-SHA-256
    [\[RFC6234\]](https://go.microsoft.com/fwlink/?LinkId=328921) of the
    NONCE field in the FEDAUTH Feature Extension Acknowledgement, and
    the client MUST verify that the nonce matches the nonce sent by the
    client in its PRELOGIN request. If the signature field does not
    match the computed HMAC-SHA-256 or if the nonce does not match the
    nonce sent by the client in its PRELOGIN request, the TDS client
    MUST close the underlying transport connection, indicate an error to
    the upper layer, and enter the \"Final State\" state.

-   When the bFedAuthLibrary is Security Token or [**Azure Active
    Directory Authentication Library
    (ADAL)**](#gt_5a728127-a59d-4cf9-8ab5-4a4e0747cc51) \[that is,
    0x02\] and any of the following statements is true, the TDS client
    MUST close the underlying transport connection, indicate an error to
    the upper layer, and enter the \"Final State\" state:

    -   The client had sent a nonce in the PRELOGIN message and either
        the NONCE field in FEDAUTH Feature Extension Acknowledgement is
        not present or the NONCE field does not match the nonce sent by
        the client in its PRELOGIN request.

    -   The client had not sent a nonce in its PRELOGIN request, and
        there is a NONCE field present in the FEDAUTH Feature Extension
        Acknowledgement.

If the response received from the server contains a structurally valid
Login response indicating a successful login and no Routing response is
detected, the TDS client MUST indicate successful Login completion to
the upper layer and enter the \"Logged In\" state.

If the response received from the server contains a structurally valid
Login response indicating a successful login and also contains a routing
response (a Routing or Enhanced Routing ENVCHANGE token) after the
LOGINACK token, the TDS client MUST enter the \"Routing Completed\"
state.

If the response received from the server does not contain a structurally
valid Login response or it contains a structurally valid Login response
indicating login failure, the TDS client MUST close the underlying
transport connection, indicate an error to the upper layer, and enter
the \"Final State\" state.

#### Sent LOGIN7 Record with SPNEGO Packet State

If the response received from the server contains a structurally valid
Login response indicating a successful login and no Routing response is
detected, the TDS client MUST indicate successful Login completion to
the upper layer and enter the \"Logged In\" state.

If the response received from the server contains a structurally valid
Login response indicating a successful login and also contains a routing
response (a Routing or Enhanced Routing ENVCHANGE token) after the
LOGINACK token, the TDS client MUST enter the \"Routing Completed\"
state.

If the response received from the server contains a structurally valid
SSPI response message, the TDS client MUST send to the server a SSPI
message (TDS packet type %x11) containing the data obtained from the
applicable SSPI layer. The TDS client SHOULD wait for the response and
reenter this state when the response is received.

If the response received from the server does not contain a structurally
valid Login response or SSPI response, or if it contains a structurally
valid Login response indicating login failure, the TDS client MUST close
the underlying transport connection, indicate an error to the upper
layer, and enter the \"Final State\" state.

#### Sent LOGIN7 Record with Federated Authentication Information Request State

If the response received from the server contains a structurally valid
Login Response message that contains a Routing or Enhanced Routing
ENVCHANGE token in the response after the LOGINACK token, the TDS client
MUST enter the \"Routing Completed\" state.

If the response received from the server contains a structurally valid
Login Response message that contains a FEDAUTHINFO token, the TDS client
MUST generate a Federated Authentication message, send that Federated
Authentication message to the server, and enter the \"Sent LOGIN7 Record
with Complete Authentication Token\" state.

If the response received from the server does not contain a structurally
valid Login Response message that contains a routing response or a
structurally valid FEDAUTHINFO token, the TDS client MUST close the
underlying transport connection, indicate an error to the upper layer,
and enter the \"Final State\" state.

#### Logged In State

The TDS client waits for notification from the upper layer. If the upper
layer requests a query to be sent to the server, the TDS client MUST
send the appropriate request to the server and enter the \"Sent Client
Request\" state. If [**MARS**](#gt_762fe1e3-0979-4402-b963-1e9150de133d)
is enabled, the TDS client MUST send the appropriate request to the SMP
layer. If the upper layer requests a termination of the connection, the
TDS client MUST disconnect from the server and enter the \"Final State\"
state. If the TDS client detects a connection error from the transport
layer, the TDS client MUST disconnect from the server and enter the
\"Final State\" state.

#### Sent Client Request State

If the response received from the server contains a structurally valid
response, the TDS client MUST indicate the result of the request to the
upper layer and enter the \"Logged In\" state.

The client has the ability to return data/control to the upper layers
while remaining in the \"Sent Client Request\" state while the complete
response has not been received or processed.

If the TDS client supports Cancel Request and the upper layer requests a
Cancel Request to be sent to the server, the TDS client sends an
Attention message to the server, start the Cancel Timer, and enter the
\"Sent Attention\" state.

If the response received from the server does not contain a structurally
valid response, the TDS client MUST close the underlying transport
connection, indicate an error to the upper layer, and enter the \"Final
State\" state.

#### Sent Attention State

If the response is structurally valid and it does not acknowledge the
Attention as described in section
[2.2.1.7](#Section_dc28579f49b14a789c5f63fbda002d2e), then the TDS
client MUST discard any data contained in the response and remain in the
\"Sent Attention\" state.

If the response is structurally valid and it acknowledges the Attention
as described in section 2.2.1.7, then the TDS client MUST discard any
data contained in the response, indicate the completion of the query to
the upper layer together with the cause of the Attention (either an
upper-layer cancellation as described in section
[3.2.4](#Section_e7af6058672941438815f82a475f9ba6) or query timeout as
described in section
[3.2.2](#Section_8313b90eeee347e1844173b3b5df1e39)), and enter the
\"Logged In\" state.

If the response received from the server is not structurally valid, then
the TDS client MUST close the underlying transport connection, indicate
an error to the upper layer, and enter the \"Final State\" state.

#### Routing Completed State

The TDS client MUST:

-   Read the rest of the login response from the server, processing the
    remaining tokens until the final DONE token is read, as it does with
    a normal login response.

-   Discard all information read from the original login response except
    for the routing information supplied in the Routing or Enhanced
    Routing ENVCHANGE token.

    -   Any information in the original login response (for example, the
        language, collation, packet size, or database mirroring partner)
        does not apply to the subsequent connection established to the
        alternate server specified in the Routing or Enhanced Routing
        ENVCHANGE token.

    -   The alternate database specified in an Enhanced Routing
        ENVCHANGE token overrides any previous database and must be used
        when connecting to the alternate server specified in the token.

-   Close the original connection and enter the \"Final State\" state.
    The original connection cannot be used for any other purpose after
    the Routing or Enhanced Routing ENVCHANGE token is read and the
    response is drained.

#### Final State

The \"Final State\" state is achieved when the application layer has
finished the communication and the lower-layer connection is
disconnected. All resources for this connection are recycled by the TDS
server.

### Timer Events

If a TDS client implements the Connection Timer and the timer times out,
then the TDS client MUST close the underlying connection, indicate the
error to the upper layer, and enter the \"Final State\" state.

If a TDS client implements the Client Request Timer and the timer times
out, then the TDS client MUST send an Attention message to the server
and enter the \"Sent Attention\" state.

If a TDS client implements the Cancel Timer and the timer times out,
then the TDS client MUST close the underlying connection, indicate the
error to the upper layer, and enter the \"Final State\" state.

### Other Local Events

Whenever an indication of a connection error is received from the
underlying transport, the TDS client MUST close the transport
connection, indicate an error to the upper layer, stop any timers if
running, and enter the \"Final State\" state. If TCP is used as the
underlying transport, examples of events that can trigger such
action---dependent on the actual TCP implementation---might be media
sense loss, a TCP connection going down in the middle of communication,
or a TCP keep-alive failure.

## Server Details

The following state machine diagrams describe TDS on the server side.
Depending on the first bytes received, one of the following flows would
be initiated.

![TDS server state machine if the first packet received is
PRELOGIN](media/image12.bin "TDS server state machine if the first packet received is PRELOGIN"){width="5.686111111111111in"
height="6.557638888888889in"}

Figure 12: TDS server state machine if the first packet received is
PRELOGIN

![TDS server state machine if the first packet received is TLS
ClientHello](media/image13.bin "TDS server state machine if the first packet received is TLS ClientHello"){width="5.980555555555555in"
height="7.397221128608924in"}

Figure 13: TDS server state machine if the first packet received is TLS
ClientHello

### Abstract Data Model

This section describes a conceptual model of data organization that an
implementation maintains to participate in this protocol. The
organization is provided to explain how the protocol behaves. This
document does not mandate that implementations adhere to this model as
long as their external behavior is consistent with what is described in
this document.

The server SHOULD maintain the following states:

-   [Initial State](#Section_8bfcbcd21baf47928fff86a63a9e90fa)

-   [TLS/SSL Negotiation
    State](#Section_ef1c4791413f4ec7ad475810a514db94) (applies to only
    TDS 7.x)

-   [TLS Negotiation State](#Section_09e5bc81690b4a9e942a24067e87bd9e)
    (applies to only TDS 8.0)

-   [PRELOGIN Ready State](#Section_e78fd371ce074bfa8597fbca06bb1ed9)
    (applies to only TDS 8.0)

-   [Login Ready State](#Section_1d82a68b57c246e984baed41f7bd0c7c)

-   [SPNEGO Negotiation
    State](#Section_7b617d06e13845f3bcea35619b14ac72)

-   [Federated Authentication Ready
    State](#Section_f6374988cc5446dbadbeb7742b8600e6)

-   [Logged In State](#Section_7c37e9c36c6544dfbbd81e8fc11a618d)

-   [Client Request Execution
    State](#Section_21c9c608c6bc456fb7eb9f09a90bd282)

-   [Routing Completed State](#Section_5d9e2ee4b6f24333a9b8944109b78c80)

-   [Final State](#Section_f9157109df6b4f23ab9135a5ca43260d)

### Timers

The TDS protocol does not regulate any timer on a data stream. The TDS
server MAY implement a timer on any message found in section
[2](#Section_8d0d4f5624a145ffb0bacc7d17517e82).

### Initialization

The server MUST establish a listening endpoint based on one of the
transport protocols described in section
[2.1](#Section_fd30432f71b2488cb30f19737d76d970). The server can
establish additional listening endpoints.

When a client makes a connection request, the transport layer listening
endpoint initializes all resources required for this connection. The
server is ready to receive a Pre-Login message (section
[2.2.1.1](#Section_58886b79ec7542f3bd7be32f327366b6)).

### Higher-Layer Triggered Events

A higher layer can choose to terminate a TDS connection at any time. In
the current TDS implementation, the upper layer can kill a connection.
When this happens, the server MUST terminate the connection and recycle
all resources for this connection. No response is sent to the client.

### Message Processing Events and Sequencing Rules

The processing of messages received from a TDS client depends on the
message type and the current state of the TDS server. The rest of this
section describes the processing and actions to take on them. The
message type is determined from the TDS packet type and the token stream
inside the TDS packet payload, as described in section
[2.2](#Section_a734db771cb247e4b61a0c9dbd8f960f).

The corresponding action is taken when the server is in the following
states. In the following processing and actions, aspects that do not
apply to both TDS 7.x and TDS 8.0 are explicitly identified in the text.

#### Initial State

The \"Initial State\" state is a prerequisite for application-layer
communication, and a lower-layer channel that can provide reliable
communication MUST be established. The TDS server enters the \"Initial
State\" state when the first packet is received from the client. The
packet SHOULD be a PRELOGIN packet to set up context for login or a TLS
ClientHello to set up context for the TLS handshake. A Pre-Login message
is indicated by the PRELOGIN (0x12) message type described in section
[2](#Section_8d0d4f5624a145ffb0bacc7d17517e82). A TLS ClientHello is
indicated when the first byte is equal to 0x16.

If the first packet is not a structurally correct PRELOGIN packet, or if
the PRELOGIN packet does not contain the client version as the first
option token, or if the first packet is not indicated to be a TLS
ClientHello, the TDS server MUST close the underlying transport
connection, indicate an error to the upper layer, and enter the \"Final
State\" state.

Otherwise, the TDS server MUST do one of the following.

-   If the first packet is a PRELOGIN packet:

    -   Return to the client a PRELOGIN structure wrapped in a [**table
        response**](#gt_71dd1dd2-c167-49a8-a5f5-6b0df5c8b48a) (0x04)
        packet and enter the \"TLS/SSL Negotiation\" state if encryption
        is negotiated.

    -   Return to the client a PRELOGIN structure wrapped in a table
        response (0x04) packet and enter unencrypted \"Login Ready\"
        state if encryption is not negotiated.

-   If the first packet is a TLS ClientHello, the TDS server MUST enter
    the \"TLS Negotiation\" state.

If a FEDAUTHREQUIRED option is contained in the PRELOGIN structure sent
by the server to the client, the TDS server MUST maintain the value of
the FEDAUTHREQUIRED option in a state variable to validate the LOGIN7
message with FEDAUTH FeatureId when the message arrives, as described in
section [3.3.5.5](#Section_1d82a68b57c246e984baed41f7bd0c7c).

If no FEDAUTHREQUIRED option is contained in the PRELOGIN structure sent
by the server to the client, or if the value of B_FEDAUTHREQUIRED = 0,
the TDS client can treat both events as equivalent and MUST remember the
event in a state variable. Either state is treated the same when the
state variables are examined in the \"Login Ready\" state (see section
3.3.5.5 for further details).

If NONCEOPT is specified in both the client PRELOGIN message and the
server PRELOGIN message, the TDS server MUST maintain a state variable
that includes the values of both the NONCE it sent to the client and the
NONCE the client sent to it during the PRELOGIN exchange.

#### TLS/SSL Negotiation State

***Applies to only TDS 7.x***

If the next packet from the TDS client is not a TLS/SSL negotiation
packet or the packet is not structurally correct, the TDS server MUST
close the underlying transport connection, indicate an error to the
upper layer, and enter the \"Final State\" state.

A TLS/SSL negotiation packet is a PRELOGIN (0x12) packet header
encapsulated with TLS/SSL payload. The TDS server MUST exchange a
TLS/SSL negotiation packet with the client and reenter this state until
the TLS/SSL negotiation is successfully completed. In this case, the TDS
server enters the \"Login Ready\" state.

#### TLS Negotiation State

***Applies to only TDS 8.0***

If the next packet from the TDS client is not a TLS negotiation packet,
or if the packet is not structurally correct, the TDS server closes the
underlying transport connection, indicates an error to the upper layer,
and enters the \"Final State\" state. A TLS negotiation packet is a
standard TLS packet. The TDS server MUST exchange the TLS negotiation
packet with the client and reenter this state until the TLS negotiation
is successfully completed. In this case, the TDS server enters the
"PRELOGIN Ready" state.

#### PRELOGIN Ready State

***Applies to only TDS 8.0***

If the packet is not a structurally correct PRELOGIN packet or if the
PRELOGIN packet does not contain the client version as the first option
token, the TDS server closes the underlying transport connection,
indicates an error to the upper layer, and enters the \"Final State\"
state. Otherwise, the TDS server MUST return to the client a PRELOGIN
structure wrapped in a [**table
response**](#gt_71dd1dd2-c167-49a8-a5f5-6b0df5c8b48a) (0x04) packet and
enter the \"Login Ready\" state if encryption is negotiated.

If a FEDAUTHREQUIRED option is contained in the PRELOGIN structure sent
by the server to the client, the TDS server MUST maintain the value of
the FEDAUTHREQUIRED option in a state variable to validate the LOGIN7
message with the FEDAUTH FeatureId when the message arrives, as
described in section
[3.3.5.5](#Section_1d82a68b57c246e984baed41f7bd0c7c).

If no FEDAUTHREQUIRED option is contained in the PRELOGIN structure sent
by the server to the client or if the value of B_FEDAUTHREQUIRED = 0,
the TDS client can treat both events as equivalent and MUST remember the
event in a state variable. Either state is treated the same when the
state variables are examined in the \"Login Ready\" state (see section
3.3.5.5 for further details).

If NONCEOPT is specified in both the client PRELOGIN message and the
server PRELOGIN message, the TDS server MUST maintain a state variable
that includes the values of both the NONCE it sent to the client and the
NONCE the client sent to it during the PRELOGIN exchange.

#### Login Ready State

If the TDS server receives a valid LOGIN7 message with the FEDAUTH
FeatureId from the client, the server MUST validate that one of the
following is true:

-   The TDS server\'s PRELOGIN structure contained a FEDAUTHREQUIRED
    option with the value 0x00, or the TDS server's PRELOGIN structure
    did not contain a FEDAUTHREQUIRED option, and the value of
    fFedAuthEcho is 0.

-   The TDS server\'s PRELOGIN structure contained a FEDAUTHREQUIRED
    option with the value 0x01, and the value of fFedAuthEcho is 1.

If the TDS server receives a valid LOGIN7 message with the FEDAUTH
FeatureId from the client but neither of the above statements is true,
the server MUST send an ERROR packet, described in section
[2](#Section_8d0d4f5624a145ffb0bacc7d17517e82), to the client. The TDS
server MUST then close the underlying transport connection, indicate an
error to the upper layer, and enter the \"Final State\" state.
Otherwise, the TDS server MUST process the FedAuthToken embedded in the
packet in a way appropriate for the value of bFedAuthLibrary.

When the bFedAuthLibrary is a Live ID Compact token, the TDS Server MUST
respond as follows:

-   If no NONCEOPT was specified in the client's PRELOGIN message, the
    TDS server MUST send a \"Login failed\" ERROR token to the client,
    close the connection, and enter the \"Final State\" state.

-   If a NONCEOPT was specified in the client\'s PRELOGIN message, the
    federated authentication library layer responds with one of two
    results, and the TDS server continues processing according to the
    response as follows:

    -   Success:

        -   The TDS server MUST use the session key from the federated
            authentication token to compute the HMAC-SHA-256
            [\[RFC6234\]](https://go.microsoft.com/fwlink/?LinkId=328921)
            of the data sent by the client. If the Signature field does
            not match the computed HMAC-SHA-256, or if the nonce does
            not match the nonce sent by the server in its PRELOGIN
            response, then the TDS server MUST send a \"Login failed\"
            ERROR token to the client, close the connection, and enter
            the \"Final State\" state.

        -   If a ChannelBindingToken is present, the server MUST compare
            the ChannelBindingToken against the channel binding token
            calculated from the underlying TLS/SSL channel. If the two
            values do not match, then the TDS server MUST send a \"Login
            failed\" ERROR token to the client, close the connection,
            and enter the \"Final State\" state.

        -   If both the channel binding token and the nonce match the
            expected values, the server MUST send the security token to
            the upper layer (an application that provides database
            management functions) for authorization. If the upper layer
            approves the security token, the TDS server MUST send a
            LOGINACK message that includes a FEATUREEXTACK token with
            the FEDAUTH FeatureId and immediately enter the \"Logged
            In\" state or enter the \"Routing Completed\" state if the
            server decides to route. If the upper layer rejects the
            security token, the TDS server MUST send a \"Login failed\"
            ERROR token to the client, close the connection, and enter
            the \"Final State\" state.

    -   Error: The server then MUST close the underlying transport
        connection, indicate an error to the upper layer, and enter the
        \"Final State\" state.

When the bFedAuthLibrary is Security Token, the TDS server MUST respond
as follows:

-   If the server's PRELOGIN response contained a NONCEOPT, the TDS
    Server MUST validate to see whether the client\'s LOGIN7 packet has
    the same nonce echoed back as part of FEDAUTH Feature SignedData. If
    the NONCE field is not present or if the nonce does not match, the
    TDS server MUST send a \"Login failed\" ERROR token to the client,
    close the connection, and enter the \"Final State\" state.

-   If the server's PRELOGIN response did not contain a NONCEOPT, the
    TDS Server MUST verify that there is NO NONCE as part LOGIN7 FEDAUTH
    Feature SignedData. If a NONCE field is present, the TDS server MUST
    send a \"Login failed\" ERROR token back to the client, close the
    connection, and enter the \"Final State\" state.

    -   Success:

        -   The server MUST send the security token to the upper layer
            (an application that provides database management functions)
            for authorization. If the upper layer approves the security
            token, the TDS server MUST send a LOGINACK message that
            includes a FEATUREEXTACK token with the FEDAUTH FeatureId
            and immediately enter the \"Logged In\" state or enter the
            \"Routing Completed\" state if the server decides to route.
            If the upper layer rejects the security token, the TDS
            server MUST send a \"Login failed\" ERROR token to the
            client, close the connection, and enter the \"Final State\"
            state.

    -   Error: The server then MUST close the underlying transport
        connection, indicate an error to the upper layer, and enter the
        \"Final State\" state.

When bFedAuthLibrary is [**Azure Active Directory Authentication Library
(ADAL)**](#gt_5a728127-a59d-4cf9-8ab5-4a4e0747cc51) \[that is, 0x02\],
the TDS server MUST validate that no other data was sent as part of the
feature extension, that is, that FeatureExt is structurally valid for
this library type. Then the TDS server MUST send a FEDAUTHINFO token
with data for FedAuthInfoIDs of STSURL and SPN and enter the \"Federated
Authentication Ready\" state. This FEDAUTHINFO Token message SHOULD be
used by the client to generate a federated authentication token.

If the TDS server receives a valid LOGIN7 packet with standard login,
the TDS server MUST respond to the TDS client with a LOGINACK (0xAD)
described in section 2 indicating login succeed. The TDS server MUST
enter the \"Logged in\" state or enter the \"Routing Completed\" state
if the server decides to route.

If the TDS server receives a LOGIN7 packet with SSPI Negotiation packet,
the TDS server MUST enter the \"SPNEGO Negotiation\" state.

If the TDS server receives a LOGIN7 packet with standard login packet,
but the login is invalid, the TDS server MUST send an ERROR packet,
described in section 2, to the client. The TDS server MUST close the
underlying transport connection, indicate an error to the upper layer,
and enter the \"Final State\" state.

If the packet received is not a structurally valid LOGIN7 packet, the
TDS server does not send any response to the client. The TDS server MUST
close the underlying transport connection, indicate an error to the
upper layer, and enter the \"Final State\" state.

#### SPNEGO Negotiation State

This state is used to negotiate the security scheme between the client
and server. The TDS server processes the packet received according to
the following rules.

-   If the packet received is a structurally valid
    [**SPNEGO**](#gt_bc2f6b5e-e5c0-408b-8f55-0350c24b9838)
    [\[RFC4178\]](https://go.microsoft.com/fwlink/?LinkId=90461)
    negotiation packet, the TDS server delegates processing of the
    security token embedded in the packet to the SPNEGO layer. The
    SPNEGO layer responds with one of three results, and the TDS server
    continues processing according to the response as follows:

    -   Complete: The TDS server then sends the security token to the
        upper layer (an application that provides database management
        functions) for authorization. If the upper layer approves the
        security token, the TDS server returns the security token to the
        client within a LOGINACK message and immediately enters the
        \"Logged In\" state or enters the \"Routing Completed\" state if
        the server decides to route. If the upper layer rejects the
        security token, then a \"Login failed\" ERROR token is sent back
        to the client, and the TDS server closes the connection and
        enters the \"Final State\" state.

    -   Continue: The TDS server sends a SPNEGO \[RFC4178\] negotiation
        response to the client, embedding the new security token
        returned by SPNEGO as part of the Continue response. The server
        then waits for a message from the client and re-renters the
        \"SPNEGO Negotiation\" state when such a packet is received.

    -   Error: The server then MUST close the underlying transport
        connection, indicate an error to the upper layer, and enter the
        \"Final State\" state.

-   If the packet received is not a structurally valid SPNEGO
    \[RFC4178\] negotiation packet, the TDS server sends no response to
    the client. The TDS server MUST close the underlying transport
    connection, indicate an error to the upper layer, and enter the
    \"Final State\" state.

#### Federated Authentication Ready State

This state is used to process the [**federated
authentication**](#gt_5ae22a0e-5ff4-441b-80d4-224ef4dd4d19) token that
is obtained from the client. The TDS server processes the packet that is
received according to the following rules:

-   If the packet that is received is a structurally valid Federated
    Authentication Token message, the TDS server MUST delegate
    processing of the security token embedded in the packet to the
    federated authentication layer, using the library that is indicated
    by the state variable that maintains the value of the
    bFedAuthLibrary field of the login packet's FEDAUTH FeatureExt. The
    federated authentication layer responds with one of two results, and
    the TDS server continues processing according to the response as
    follows:

    -   SUCCESS: The TDS Server MUST send the Federated Authentication
        Token to the upper layer (an application that provides database
        management functions) for authorization. If the upper layer
        approves the token, the TDS server MUST send a LoginACK message
        that includes a FEATUREEXTACK token that contains FEDAUTH
        FeatureId and immediately enter the \"Logged In\" state or enter
        the \"Routing Completed\" state if the server decides to route.
        If the upper layer rejects the token, then a \"Login Failed\"
        ERROR token MUST be sent back to the client, and the TDS server
        MUST close the connection and enter the \"Final State\" state.

    -   ERROR: The server MUST close the underlying transport
        connection, indicate an error to the upper layer, and enter the
        \"Final State\" state.

-   If the packet that is received is not a structurally valid Federated
    Authentication Token message, the TDS server SHOULD send no response
    to the client. The TDS server MUST close the underlying transport
    connection, indicate an error to the upper layer, and enter the
    \"Final State\" state.

#### Logged In State

If a TDS of type 1, 3, 7, or 14 (see section
[2.2.3.1.1](#Section_9b4a463c26344a4bac35bebfff2fb0f7)) arrives, then
the TDS server begins processing by raising an event to the upper layer
containing the data of the client request and entering the \"Client
Request Execution\" state. If any other TDS types arrive, the server
MUST enter the \"Final State\" state. The TDS server MUST continue to
listen for messages from the client while awaiting notification of
client request completion from the upper layer.

#### Client Request Execution State

The TDS server MUST continue to listen for messages from the client
while awaiting notification of client request for completion from the
upper layer. The TDS server MUST also do one of the following:

-   If the upper layer notifies TDS that the client request has finished
    successfully, the TDS server MUST send the results in the formats
    described in section [2](#Section_8d0d4f5624a145ffb0bacc7d17517e82)
    to the TDS client and enter the \"Logged In\" state.

-   If the upper layer notifies TDS that an error has been encountered
    during client request, the TDS server MUST send an error message
    (described in section 2) to the TDS client and enter the \"Logged
    In\" state.

-   If an attention packet (described in section 2) is received during
    the execution of the current client request, it MUST deliver a
    cancel indication to the upper layer. If an attention packet
    (described in section 2) is received after the execution of the
    current client request, it SHOULD NOT deliver a cancel indication to
    the upper layer because there is no existing execution to cancel.
    The TDS server MUST send an attention acknowledgment to the TDS
    client and enter the \"Logged In\" state.

-   If another client request packet is received during the execution of
    the current client request, the TDS server SHOULD queue the new
    client request, and continue processing the client request already
    in progress according to the preceding rules. When this operation is
    complete, the TDS server re-enters the \"Client Request Execution\"
    state and processes the newly arrived message.

-   If [**MARS**](#gt_762fe1e3-0979-4402-b963-1e9150de133d) is enabled,
    all TDS server responses to client request messages MUST be passed
    through to the SMP layer.

-   If any other message type arrives, the server MUST close the
    connection and enter the \"Final State\" state.

#### Routing Completed State

The TDS server SHOULD wait for connection closure initiated by the
client and enter the \"Final State\" state. If any request is received
from the client in this state, the server SHOULD close the connection
with no response and enter the \"Final State\" state.

#### Final State

The \"Final State\" state is achieved when the application layer has
finished the communication, and the lower-layer connection is
disconnected. All resources for this connection are recycled by the TDS
server.

### Timer Events

None.

### Other Local Events

When there is a failure in under-layers, the server SHOULD terminate the
[**TDS session**](#gt_276cd76b-c0a2-4f7c-8529-ad0d60aa9592) without
sending any response to the client. The under-layer failure could be
triggered by network failure. It can also be triggered by the
termination action from the client, which could be communicated to the
server stack by under-layers.

