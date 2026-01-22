### Server Messages

Messages sent from the server to the client are the following:

-   [Pre-Login Response](#Section_63c37a5d459d40b98c4a47e13aff235d)

-   [Login Response](#Section_da5a73bbe49247ab9d9c23ac00292efc)

-   [Federated Authentication
    Information](#Section_1a22145b7657435d9c60318be82456e1)

-   [Row Data](#Section_cc5aaa57f2e04571910ac5f6a538f88f)

-   [Return Status](#Section_ed493f4cf63d41739476e64261400119)

-   [Return Parameters](#Section_43acbd448d754f2e8f5c543b59a9d517)

-   [Response Completion](#Section_6ee2125f34be43c6a38ce9ee7338500f)

-   [Error and Info](#Section_7fffdf0c78d140fd9dee1e6b4246d3aa)

-   [Attention
    Acknowledgement](#Section_9bcc4b79a79141d0bf7673a97e501b35)

These messages are briefly described in the sections that follow.
Detailed descriptions of message contents are in section
[2.2.6](#Section_c060af9c6db74360954cc28a132c8949) and section
[2.2.7](#Section_67b6113cd72242d1902c3f6e8de09173).

#### Pre-Login Response

The Pre-Login Response message is a tokenless packet [**data
stream**](#gt_151643ce-fb5e-460e-8bdf-dc10bbd1950e). The data stream
consists of the response to the information requested by the client\'s
Pre-Login message. For more details, see section
[2.2.6.5](#Section_60f5640801884cd58b9025c6f2423868).

#### Login Response

The Login Response message is a token stream that consists of
information about the server\'s characteristics, optional information
and error messages, and finally, a completion message.

The LOGINACK token [**data
stream**](#gt_151643ce-fb5e-460e-8bdf-dc10bbd1950e) includes information
about the server
[**interface**](#gt_95913fbd-3262-47ae-b5eb-18e6806824b9) and the
server\'s product code and name. For more details, see section
[2.2.7.14](#Section_490e563dcc6e4c86bb95ef0186b98032).

If there are any messages in the login response, an ERROR or INFO token
data stream is returned from the server to the client. For more details,
see sections [2.2.7.10](#Section_9805e9fa1f8b4cf88f788d2602228635) and
[2.2.7.13](#Section_284bb815d0834ed5b33abdc2492e322b), respectively.

The server can send, as part of the login response, one or more
ENVCHANGE token data streams if the login changed the environment and
the associated notification flag was set. An example of an environment
change includes the current database context and language setting. For
more details, see section
[2.2.7.9](#Section_2b3eb7e5d43d4d1bbf4d76b9e3afc791).

A done packet MUST be present as the final part of the login response,
and a DONE token data stream is the last thing sent in response to a
server login request. For more details, see section
[2.2.7.6](#Section_3c06f11098bd4d5bb836b1ba66452cb7).

#### Federated Authentication Information

After the server receives a Login message that states that the client
intends to use a [**federated
authentication**](#gt_5ae22a0e-5ff4-441b-80d4-224ef4dd4d19) token from a
specific client library that needs additional information from the
server to generate that token, if the server supports federated
authentication that uses that client library, the server responds to the
client with a message. This message contains a Federated Authentication
Information Token that provides the information necessary for the client
to generate a federated authentication token. If the server determines
that no information is required for this particular client library, the
server does not send the information token. For more details, see
section [2.2.7.12](#Section_0e4486d6d407496298030c1a4d4d87ce).

#### Row Data

If the server request results in data being returned, the data precedes
any other data streams returned from the server except warnings. Row
data MUST be preceded by a description of the column names and data
types. For more information about how the column names and data types
are described, see section
[2.2.7.4](#Section_58880b9f381c43b2bf8b0727a98c4f4c).

#### Return Status

When a [**stored procedure**](#gt_324d32b3-f4f3-41c9-b695-78c498094fb7)
is executed by the server, the server MUST return a status value. This
is a 4-byte integer and is sent via the RETURNSTATUS token. A stored
procedure execution is requested through either an RPC Batch or a SQL
Batch (section [2.2.1.4](#Section_5b7f9062d5144b4ea7c59a27a19596c4))
message. For more details about RETURNSTATUS, see section
[2.2.7.18](#Section_c719f199e71b418790b994f78bd1870e).

#### Return Parameters

The response format for execution of a [**stored
procedure**](#gt_324d32b3-f4f3-41c9-b695-78c498094fb7) is identical
regardless of whether the request was sent as SQL Batch (section
[2.2.1.4](#Section_5b7f9062d5144b4ea7c59a27a19596c4)) or RPC Batch. It
is always a tabular result-type message.

If the procedure explicitly sends any data, then the message starts with
a single token stream of rows, informational messages, and error
messages. This data is sent in the usual way.

When the RPC is invoked, some or all of its parameters are designated as
output parameters. All output parameters have values returned from the
server. For each output parameter, there is a corresponding return
value, sent via the RETURNVALUE token. The RETURNVALUE token [**data
stream**](#gt_151643ce-fb5e-460e-8bdf-dc10bbd1950e) is also used for
sending back the value returned by a user-defined function (UDF), if it
is called as an RPC. For more details about the RETURNVALUE token, see
section [2.2.7.19](#Section_7091f6f6b83d4ed2afebba5013dfb18f).

#### Response Completion

The client reads results in logical units and can tell when all results
have been received by examining the
[DONE](#Section_3c06f11098bd4d5bb836b1ba66452cb7) token [**data
stream**](#gt_151643ce-fb5e-460e-8bdf-dc10bbd1950e).

When executing a batch of [**SQL
statements**](#gt_dc5ca224-43ec-4b44-9dba-726d6fd6057d), the server MUST
return a DONE token data stream for each set of results. All but the
last DONE will have the DONE_MORE bit set in the **Status** field of the
DONE token data stream. Therefore, the client can always tell after
reading a DONE whether or not there are more results. For more details,
see section 2.2.7.6.

For [**stored procedures**](#gt_324d32b3-f4f3-41c9-b695-78c498094fb7),
completion of SQL statements in the stored procedure is indicated by a
[DONEINPROC](#Section_43e891c5f7a1432f8f9f233c4cd96afb) token data
stream for each SQL statement and a
[DONEPROC](#Section_65e24140edea46e5b710209af2016195) token data stream
for each completed stored procedure. For more details about DONEINPROC
and DONEPROC tokens, see section 2.2.7.7 and 2.2.7.8, respectively.

#### Error and Info

Besides returning descriptions of Row data and the data itself, TDS
provides a token [**data
stream**](#gt_151643ce-fb5e-460e-8bdf-dc10bbd1950e) type for the server
to send error and informational messages to the client. These are the
ERROR token data stream and the INFO token data stream. For more
details, see sections
[2.2.7.10](#Section_9805e9fa1f8b4cf88f788d2602228635) and
[2.2.7.13](#Section_284bb815d0834ed5b33abdc2492e322b), respectively.

#### Attention Acknowledgment

After a client has sent an interrupt signal to the server, the client
MUST read returning data until the interrupt has been acknowledged.
Attention messages are acknowledged in the DONE token [**data
stream**](#gt_151643ce-fb5e-460e-8bdf-dc10bbd1950e). For more details,
see section [2.2.7.6](#Section_3c06f11098bd4d5bb836b1ba66452cb7).

