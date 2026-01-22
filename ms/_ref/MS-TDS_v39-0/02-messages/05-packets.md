### Packets

A packet is the unit written or read at one time. A message can consist
of one or more packets. A packet always includes a packet header and is
usually followed by packet data that contains the message. Each new
message starts in a new packet.

In practice, both the client and server try to read a packet full of
data. They pick out the header to see how much more (or less) data there
is in the communication.

At login time, clients MAY specify a requested \"packet\" size as part
of the LOGIN7 (section
[2.2.6.4](#Section_773a62b6ee894c029e5e344882630aac)) message stream.
This identifies the size used to break large messages into different
\"packets\". Server acknowledgment of changes in the negotiated packet
size is transmitted back to the client via ENVCHANGE (section
[2.2.7.9](#Section_2b3eb7e5d43d4d1bbf4d76b9e3afc791)) token stream. The
negotiated packet size is the maximum value that can be specified in the
**Length** packet header field described in section
[2.2.3.1.3](#Section_c1cddd03b448470a946a9b1b908f27a7).

Starting with TDS 7.3, the following behavior MUST also be enforced. For
requests sent to the server larger than the current negotiated
\"packet\" size, the client MUST send all but the last packet with a
total number of bytes equal to the negotiated size. Only the last packet
in the request can contain an actual number of bytes smaller than the
negotiated packet size. If any of the preceding packets are sent with a
length less than the negotiated packet size, the server SHOULD
disconnect the client when the next network payload arrives.

#### Packet Header

To implement messages on top of existing, arbitrary transport layers, a
packet header is included as part of the packet. The packet header
precedes all data within the packet. It is always 8 bytes in length.
Most importantly, the packet header states the **Type** (section
[2.2.3.1.1](#Section_9b4a463c26344a4bac35bebfff2fb0f7)) and **Length**
(section [2.2.3.1.3](#Section_c1cddd03b448470a946a9b1b908f27a7)) of the
entire packet.

The following sections provide a detailed description of each item
within the packet header.

##### Type

**Type** defines the type of message. **Type** is a 1-byte unsigned
char. The following table describes the types that are available.

  ------------------------------------------------------------------------
  Value     Description                        Packet contains data?
  --------- ---------------------------------- ---------------------------
  1         SQL batch                          Yes

  2         Pre-TDS7 Login[\<5\>](\l)          Yes

  3         RPC                                Yes

  4         Tabular result                     Yes

  5         Unused                             

  6         Attention signal                   No

  7         Bulk load data                     Yes

  8         Federated Authentication Token     Yes

  9-13      Unused                             

  14        Transaction manager request        Yes

  15        Unused                             

  16        TDS7 Login[\<6\>](\l)              Yes

  17        SSPI                               Yes

  18        Pre-Login                          Yes
  ------------------------------------------------------------------------

If an unknown **Type** is specified, the message receiver SHOULD
disconnect the connection. If a valid **Type** is specified, but is
unexpected (per section [3](#Section_52b5da79b22248fdb408dd4c67464f9a)),
the message receiver SHOULD disconnect the connection. This applies to
both the client and the server. For example, the server could disconnect
the connection if the server receives a message with **Type** equal 16
when the connection is already logged in.

The following table highlights which messages, as described previously
in sections [2.2.1](#Section_7ea9ee1ab46141f29004141c0e712935) and
[2.2.2](#Section_342f4cbb2b4b489c8b63f99b12021a94), correspond to which
packet header type.

  -----------------------------------------------------------------------------------------------------------
  Message type                                                  Client or server    Packet header type
                                                                message             
  ------------------------------------------------------------- ------------------- -------------------------
  [Pre-Login](#Section_58886b79ec7542f3bd7be32f327366b6)        Client              18

  [Login](#Section_f159ec89ef1a4fcd9ffd8ee701fccbb0)            Client              16 + 17 (if Integrated
                                                                                    authentication)

  [Federated Authentication                                     Client              8
  Token](#Section_827d963229574d54b9ea384530ae79d0)                                 

  [SQL Batch](#Section_5b7f9062d5144b4ea7c59a27a19596c4)        Client              1

  [Bulk Load](#Section_88176081df754b24bcfb4c16ff03cbfa)        Client              7

  [RPC](#Section_26327437aa3c4e969bba73a6e862ba21)              Client              3

  [Attention](#Section_dc28579f49b14a789c5f63fbda002d2e)        Client              6

  [Transaction Manager                                          Client              14
  Request](#Section_b4b7856454404fc0b5efc9e1925aaefe)                               

  [FeatureExtAck](#Section_2eb82f8e11f046dcb42d27302fa4701a)    Server              4

  [Pre-Login                                                    Server              4
  Response](#Section_63c37a5d459d40b98c4a47e13aff235d)                              

  [Login Response](#Section_da5a73bbe49247ab9d9c23ac00292efc)   Server              4

  [Federated Authentication                                     Server              4
  Information](#Section_0e4486d6d407496298030c1a4d4d87ce)                           

  [Row Data](#Section_cc5aaa57f2e04571910ac5f6a538f88f)         Server              4

  [Return Status](#Section_ed493f4cf63d41739476e64261400119)    Server              4

  [Return                                                       Server              4
  Parameters](#Section_43acbd448d754f2e8f5c543b59a9d517)                            

  [Response                                                     Server              4
  Completion](#Section_6ee2125f34be43c6a38ce9ee7338500f)                            

  [Session State](#Section_626fbe19f3564599ba17c70f44005106)    Server              4

  [Error and Info](#Section_7fffdf0c78d140fd9dee1e6b4246d3aa)   Server              4

  [Attention                                                    Server              4
  Acknowledgement](#Section_9bcc4b79a79141d0bf7673a97e501b35)                       
  -----------------------------------------------------------------------------------------------------------

##### Status

**Status** is a bit field used to indicate the message state. **Status**
is a 1-byte unsigned char. The following **Status** bit flags are
defined.

+----+-----------------------------------------------------------------+
| V  | Description                                                     |
| al |                                                                 |
| ue |                                                                 |
+====+=================================================================+
| 0x | \"Normal\" message.                                             |
| 00 |                                                                 |
+----+-----------------------------------------------------------------+
| 0x | End of message (EOM). The packet is the last packet in the      |
| 01 | whole request.                                                  |
+----+-----------------------------------------------------------------+
| 0x | (From client to server) Ignore this event (0x01 MUST also be    |
| 02 | set).                                                           |
+----+-----------------------------------------------------------------+
| 0x | RESETCONNECTION                                                 |
| 08 |                                                                 |
|    | (Introduced in TDS 7.1)                                         |
|    |                                                                 |
|    | (From client to server) Reset this connection before processing |
|    | event. Only set for event types Batch, RPC, or Transaction      |
|    | Manager request. If clients want to set this bit, it MUST be    |
|    | part of the first packet of the message. This signals the       |
|    | server to clean up the environment state of the connection back |
|    | to the default environment setting, effectively simulating a    |
|    | logout and a subsequent login, and provides server support for  |
|    | connection pooling. This bit SHOULD be ignored if it is set in  |
|    | a packet that is not the first packet of the message.           |
|    |                                                                 |
|    | This status bit MUST NOT be set in conjunction with the         |
|    | RESETCONNECTIONSKIPTRAN bit. Distributed transactions and       |
|    | isolation levels are not reset.                                 |
+----+-----------------------------------------------------------------+
| 0x | RESETCONNECTIONSKIPTRAN                                         |
| 10 |                                                                 |
|    | (Introduced in TDS 7.3)                                         |
|    |                                                                 |
|    | (From client to server) Reset the connection before processing  |
|    | event but do not modify the transaction state (the state        |
|    | remains the same before and after the reset). The transaction   |
|    | in the session can be a local transaction that is started from  |
|    | the session or it can be a distributed transaction in which the |
|    | session is enlisted. This status bit MUST NOT be set in         |
|    | conjunction with the RESETCONNECTION bit. Otherwise identical   |
|    | to RESETCONNECTION.                                             |
+----+-----------------------------------------------------------------+

All other bits are not used and MUST be ignored.

##### Length

**Length** is the size of the packet including the 8 bytes in the packet
header. It is the number of bytes from the start of this header to the
start of the next packet header. **Length** is a 2-byte, unsigned short
and is represented in network byte order
([**big-endian**](#gt_6f6f9e8e-5966-4727-8527-7e02fb864e7e)).

The **Length** value MUST be greater than or equal to 512 bytes and
smaller than or equal to 32,767 bytes. The packet size MUST be smaller
than or equal to 4,096 bytes until **Length** is successfully
negotiated. The default value is 4,096 bytes.

Starting with TDS 7.3, **Length** MUST be the negotiated packet size
when sending a packet from client to server, unless it is the last
packet of a request (that is, the EOM bit in **Status** is ON) or the
client has not logged in.

##### SPID

**SPID** is the process ID on the server, corresponding to the current
connection. This information is sent by the server to the client and is
useful for identifying which thread on the server sent the TDS packet.
It is provided for debugging purposes. The client MAY send the **SPID**
value to the server. If the client does not, then a value of 0x0000
SHOULD be sent to the server. This is a 2-byte value and is represented
in network byte order
([**big-endian**](#gt_6f6f9e8e-5966-4727-8527-7e02fb864e7e)).

##### PacketID

**PacketID** is used for numbering message packets that contain data in
addition to the packet header. PacketID is a 1-byte, unsigned char. Each
time packet data is sent, the value of **PacketID** is incremented by 1,
modulo 256.\<7\> This allows the receiver to track the sequence of TDS
packets for a given message. This value is ignored.

##### Window

This 1 byte is not used. This byte SHOULD be set to 0x00 and SHOULD be
ignored by the receiver.

#### Packet Data

Packet data for a given message follows the packet header (see **Type**
in section [2.2.3.1.1](#Section_9b4a463c26344a4bac35bebfff2fb0f7) for
messages that contain packet data). As stated in section
[2.2.3](#Section_e5ea85201ea34a75a2a9c17e63e9ee19), a message can span
more than one packet. Because each new message MUST always begin within
a new packet, a message that spans more than one packet only occurs if
the data to be sent exceeds the maximum packet data size, which is
computed as (negotiated packet size - 8 bytes), where the 8 bytes
represents the size of the packet header.

If a stream spans more than one packet, then the EOM bit of the packet
header **Status** (section
[2.2.3.1.2](#Section_ce398f9a7d474ede8f369dd6fc21ca43)) code MUST be set
to 0 for every packet header. The EOM bit MUST be set to 1 in the last
packet to signal that the stream ends. In addition, the **PacketID**
field of subsequent packets MUST be incremented as defined in section
[2.2.3.1.5](#Section_ec9e8663191c4dd1baa848bbfba5ed7e).

