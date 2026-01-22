# Security

## Security Considerations for Implementers

As previously described in this specification, the TDS protocol provides
facilities for authentication and channel encryption negotiation. If
SSPI authentication is requested by the client application, the exact
choice of security mechanisms is determined by the SSPI layer. Likewise,
although the decision as to whether channel encryption is used is
negotiated in the TDS layer, the exact choice of cipher suite is
negotiated by the TLS/SSL layer. Likewise, although the decision as to
whether federated authentication or SSPI authentication is used can
optionally be negotiated in the TDS layer, the exact choice of
authentication mechanism is determined by either the SSPI layer or the
federated authentication layer.

The TDS protocol also includes a mechanism to provide information about
the sensitivity of a [**result
set**](#gt_c8a27238-8ccc-442b-9604-75f74d3e6b3d) through [**data
classification**](#gt_acdeafb0-9b24-420e-b712-9284ad49eb56). Clients can
utilize this information to further control access or annotate the
sensitive data within an application.

## Index of Security Parameters

The following table lists the sections in this document in which the
available Tabular Data Stream (TDS) security parameters are mentioned.

+----------------+-----------------------------------------------------+
| Security       | Section                                             |
| parameter      |                                                     |
+================+=====================================================+
| TLS            | [2.1](#Section_fd30432f71b2488cb30f19737d76d970)    |
| Negotiation    | Transport                                           |
|                |                                                     |
|                | [                                                   |
|                | 3.2.5.1](#Section_fc8fa46974004e328d023d54f166e7fb) |
|                | Sent Initial TLS Negotiation Packet State           |
|                |                                                     |
|                | [                                                   |
|                | 3.2.5.2](#Section_cc823ca848674387819dcc5c19da5732) |
|                | Sent Initial PRELOGIN Packet State                  |
|                |                                                     |
|                | [                                                   |
|                | 3.2.5.3](#Section_d62e225bd8654ccc8f73de1ef49e30d4) |
|                | Sent TLS/SSL Negotiation Packet State               |
|                |                                                     |
|                | [                                                   |
|                | 3.3.5.1](#Section_8bfcbcd21baf47928fff86a63a9e90fa) |
|                | Initial State                                       |
|                |                                                     |
|                | [                                                   |
|                | 3.3.5.2](#Section_09e5bc81690b4a9e942a24067e87bd9e) |
|                | TLS/SSL Negotiation State                           |
|                |                                                     |
|                | [                                                   |
|                | 3.3.5.3](#Section_ef1c4791413f4ec7ad475810a514db94) |
|                | TLS Negotiation State                               |
|                |                                                     |
|                | [                                                   |
|                | 3.3.5.4](#Section_e78fd371ce074bfa8597fbca06bb1ed9) |
|                | PRELOGIN Ready State                                |
|                |                                                     |
|                | [                                                   |
|                | 3.3.5.5](#Section_1d82a68b57c246e984baed41f7bd0c7c) |
|                | Login Ready State                                   |
+----------------+-----------------------------------------------------+
| SSPI           | [                                                   |
| Authentication | 2.2.1.2](#Section_f159ec89ef1a4fcd9ffd8ee701fccbb0) |
|                | Login                                               |
|                |                                                     |
|                | [2.                                                 |
|                | 2.3.1.1](#Section_9b4a463c26344a4bac35bebfff2fb0f7) |
|                | Type                                                |
|                |                                                     |
|                | [                                                   |
|                | 2.2.5.8](#Section_f79bb5b85919439aa69648064b78b091) |
|                | Data Packet Stream Tokens                           |
|                |                                                     |
|                | [                                                   |
|                | 2.2.6.4](#Section_773a62b6ee894c029e5e344882630aac) |
|                | LOGIN7                                              |
|                |                                                     |
|                | [                                                   |
|                | 2.2.6.5](#Section_60f5640801884cd58b9025c6f2423868) |
|                | PRELOGIN                                            |
|                |                                                     |
|                | [                                                   |
|                | 2.2.6.8](#Section_1dc6c197f606410badd9e6fe1dff0e8b) |
|                | SSPI Message                                        |
|                |                                                     |
|                | [2                                                  |
|                | .2.7.22](#Section_07e2bb7b8ba6445f89b1cc76d8bfa9c6) |
|                | SSPI                                                |
|                |                                                     |
|                | 3.2.5.2 Sent Initial PRELOGIN Packet State          |
|                |                                                     |
|                | 3.2.5.3 Sent TLS/SSL Negotiation Packet State       |
|                |                                                     |
|                | [                                                   |
|                | 3.2.5.5](#Section_c373a90b113d4c5f826ad410365ee883) |
|                | Sent LOGIN7 Record with SPNEGO Packet State         |
|                |                                                     |
|                | 3.3.5.5 Login Ready State                           |
|                |                                                     |
|                | [4.2](#Section_ce5ad23f6bf84fa594266b0d36e14da2)    |
|                | Login Request                                       |
|                |                                                     |
|                | [4.3](#Section_f88b63bbb47949e1a87bdeda521da508)    |
|                | Login Request with Federated Authentication         |
|                |                                                     |
|                | [4.11](#Section_dc57840ad13b43a1aad35425afadbc0c)   |
|                | SSPI Message                                        |
|                |                                                     |
|                | [4.16](#Section_b238ee1d249d478086a8a777f66623f2)   |
|                | FeatureExt with SESSIONRECOVERY Feature Data        |
|                |                                                     |
|                | [4.20](#Section_81d084b0ea234a9cbe4e7aadbc5a88c3)   |
|                | FeatureExt with AZURESQLSUPPORT Feature Data        |
+----------------+-----------------------------------------------------+
| Federated      | 2.2.1.2 Login                                       |
| Authentication |                                                     |
|                | [                                                   |
|                | 2.2.1.3](#Section_f40cca2929b843ecb9ff2f8682486c29) |
|                | Federated Authentication Token                      |
|                |                                                     |
|                | [                                                   |
|                | 2.2.2.3](#Section_1a22145b7657435d9c60318be82456e1) |
|                | Federated Authentication Information                |
|                |                                                     |
|                | 2.2.3.1.1 Type                                      |
|                |                                                     |
|                | [2.2.4](#Section_dc3a08548230482fbbb9d94a3b905a26)  |
|                | Packet Data Token and Tokenless Data Streams        |
|                |                                                     |
|                | [                                                   |
|                | 2.2.6.3](#Section_827d963229574d54b9ea384530ae79d0) |
|                | Federated Authentication Token                      |
|                |                                                     |
|                | 2.2.6.4 LOGIN7                                      |
|                |                                                     |
|                | 2.2.6.5 PRELOGIN                                    |
|                |                                                     |
|                | [2                                                  |
|                | .2.7.11](#Section_2eb82f8e11f046dcb42d27302fa4701a) |
|                | FEATUREEXTACK                                       |
|                |                                                     |
|                | [2                                                  |
|                | .2.7.12](#Section_0e4486d6d407496298030c1a4d4d87ce) |
|                | FEDAUTHINFO                                         |
|                |                                                     |
|                | [3.1.5](#Section_5a9d49b8d1c44ff58f6576dc924f5c4d)  |
|                | Message Processing Events and Sequencing Rules      |
|                |                                                     |
|                | 3.2.5.2 Sent Initial PRELOGIN Packet State          |
|                |                                                     |
|                | 3.2.5.3 Sent TLS/SSL Negotiation Packet State       |
|                |                                                     |
|                | [                                                   |
|                | 3.2.5.4](#Section_ab7a745416a949ef8f208a0de2b31fbb) |
|                | Sent LOGIN7 Record with Complete Authentication     |
|                | Token State                                         |
|                |                                                     |
|                | [                                                   |
|                | 3.2.5.6](#Section_c272e054279f4b09ac41f9ccf1d1ffde) |
|                | Sent LOGIN7 Record with Federated Authentication    |
|                | Information Request State                           |
|                |                                                     |
|                | 3.3.5.1 Initial State                               |
|                |                                                     |
|                | 3.3.5.5 Login Ready State                           |
|                |                                                     |
|                | [                                                   |
|                | 3.3.5.7](#Section_f6374988cc5446dbadbeb7742b8600e6) |
|                | Federated Authentication Ready State                |
|                |                                                     |
|                | 4.3 Login Request with Federated Authentication     |
|                |                                                     |
|                | [4.5](#Section_1582e9753662411e9b27c23cc70a8b4b)    |
|                | Login Response with Federated Authentication        |
|                | Feature Extension Acknowledgement                   |
+----------------+-----------------------------------------------------+
| Data           | [2.2.                                               |
| classification | 4.2.1.3](#Section_d3edea23be1f416098f50de233ffeebc) |
|                | Variable Length Tokens(xx10xxxx)                    |
|                |                                                     |
|                | 2.2.5.8 Data Packet Stream Tokens                   |
|                |                                                     |
|                | 2.2.6.4 LOGIN7                                      |
|                |                                                     |
|                | [                                                   |
|                | 2.2.7.5](#Section_813b88bc0a324e7ebc92d98f62cb8981) |
|                | DATACLASSIFICATION                                  |
|                |                                                     |
|                | 2.2.7.11 FEATUREEXTACK                              |
+----------------+-----------------------------------------------------+

