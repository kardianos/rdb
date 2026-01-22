# Index

A

Abstract data model

client ([section 3.1.1](#abstract-data-model) 127, [section
3.2.1](#abstract-data-model-1) 136)

server ([section 3.1.1](#abstract-data-model) 127, [section
3.3.1](#abstract-data-model-2) 146)

ALL_HEADERS rule definition

[overview](#packet-data-stream-headers---all_headers-rule-definition) 35

[Query Notifications header](#query-notifications-header) 36

[Transaction Descriptor header](#transaction-descriptor-header) 37

[Applicability](#applicability-statement) 17

[Attention message](#attention) 21

[Attention request example](#attention-request) 179

[Attention tokens](#done-and-attention-tokens) 29

C

[Capability negotiation](#versioning-and-capability-negotiation) 17

[Change tracking](#change-tracking) 227

Client

abstract data model ([section 3.1.1](#abstract-data-model) 127, [section
3.2.1](#abstract-data-model-1) 136)

higher-layer triggered events ([section
3.1.4](#higher-layer-triggered-events) 127, [section
3.2.4](#higher-layer-triggered-events-1) 138)

initialization ([section 3.1.3](#initialization) 127, [section
3.2.3](#initialization-1) 137)

[local events](#other-local-events) 134

message processing ([section
3.1.5](#message-processing-events-and-sequencing-rules) 127, [section
3.2.5](#message-processing-events-and-sequencing-rules-1) 139)

[Final state](#final-state) 143

[Logged In state](#logged-in-state) 142

[Sent Attention state](#sent-attention-state) 143

[Sent Client Request state](#sent-client-request-state) 143

[Sent Initial PRELOGIN Packet
state](#sent-initial-prelogin-packet-state) 139

[Sent LOGIN7 Record with SPNEGO Packet
state](#sent-login7-record-with-spnego-packet-state) 142

[Sent LOGIN7 Record with Standard Login
state](#sent-login7-record-with-complete-authentication-token-state) 141

[Sent TLS/SSL Negotiation Packet
state](#sent-tlsssl-negotiation-packet-state) 140

messages

[Attention](#attention) 21

[login](#login) 20

[overview](#client-messages) 19

[pre-login](#pre-login) 20

[remote procedure call](#remote-procedure-call) 21

[SQL command](#sql-batch) 20

[SQL command with binary data](#bulk-load) 20

[transaction manager request](#transaction-manager-request) 21

[other local events](#other-local-events-1) 144

overview ([section 3.1](#common-details) 127, [section
3.2](#client-details) 135)

sequencing rules ([section
3.1.5](#message-processing-events-and-sequencing-rules) 127, [section
3.2.5](#message-processing-events-and-sequencing-rules-1) 139)

[Final state](#final-state) 143

[Logged In state](#logged-in-state) 142

[Sent Attention state](#sent-attention-state) 143

[Sent Client Request state](#sent-client-request-state) 143

[Sent Initial PRELOGIN Packet
state](#sent-initial-prelogin-packet-state) 139

[Sent LOGIN7 Record with SPNEGO Packet
state](#sent-login7-record-with-spnego-packet-state) 142

[Sent LOGIN7 Record with Standard Login
state](#sent-login7-record-with-complete-authentication-token-state) 141

[Sent TLS/SSL Negotiation Packet
state](#sent-tlsssl-negotiation-packet-state) 140

timer events ([section 3.1.6](#timer-events) 134, [section
3.2.6](#timer-events-1) 144)

timers ([section 3.1.2](#timers) 127, [section 3.2.2](#timers-1) 137)

[Client Messages message](#client-messages) 19

[Client Request Execution state](#client-request-execution-state) 152

D

Data model - abstract

client ([section 3.1.1](#abstract-data-model) 127, [section
3.2.1](#abstract-data-model-1) 136)

server ([section 3.1.1](#abstract-data-model) 127, [section
3.3.1](#abstract-data-model-2) 146)

Data stream types

[data type dependent data streams](#data-type-dependent-data-streams) 34

[unknown-length data streams](#unknown-length-data-streams) 33

[variable-length data streams](#variable-length-data-streams) 33

Data type definitions

[fixed-length data types](#fixed-length-data-types) 38

[overview](#data-type-definitions) 37

[partially length-prefixed data
types](#partially-length-prefixed-data-types) 41

[SQL_VARIANT](#sql_variant-values) 45

Table Valued Parameter

[metadata](#metadata) 46

[optional metadata tokens](#optional-metadata-tokens) 48

[overview](#table-valued-parameter-tvp-values) 46

[TDS type restrictions](#tds-type-restrictions) 50

[UDT Assembly Information](#common-language-runtime-clr-instances) 44

[variable-length data types](#variable-length-data-types) 38

[XML data type](#xml-values) 44

[DONE tokens](#done-and-attention-tokens) 29

E

[Error messages](#error-and-info) 23

Examples

[attention request](#attention-request) 179

[FeatureExt with AZURESQLSUPPORT Feature
Data](#featureext-with-azuresqlsupport-feature-data) 204

[FeatureExt with SESSIONRECOVERY feature
data](#featureext-with-sessionrecovery-feature-data) 190

[FeatureExtAck with AZURESQLSUPPORT Feature
Data](#featureextack-with-azuresqlsupport-feature-data) 207

[FeatureExtAck with SESSIONRECOVERY feature
data](#featureextack-with-sessionrecovery-feature-data) 195

[login request](#login-request) 155

[login request with federated
authentication](#login-request-with-federated-authentication) 157

[login response](#login-response-1) 164

[login response with federated
authentication](#login-response-with-federated-authentication-feature-extension-acknowledgement)
168

[overview](#protocol-examples) 154

[pre-login request](#pre-login-request) 154

[RPC client request](#rpc-client-request) 176

[RPC server response](#rpc-server-response) 178

[SparseColumn select statement](#sparsecolumn-select-statement) 185

[SQL batch client request](#sql-batch-client-request) 174

[SQL batch server response](#sql-batch-server-response) 175

[SQL command with binary data](#bulk-load-1) 180

[SSPI message](#sspi-message-1) 179

[Table response with SESSIONSTATE token
data](#table-response-with-sessionstate-token-data) 201

[token data stream](#token-stream-communication) 203

[attention signal - out-of-band](#out-of-band-attention-signal) 203

[sending an SQL batch](#sending-a-sql-batch) 203

[transaction manager request](#transaction-manager-request-2) 181

[TVP insert statement](#tvp-insert-statement) 182

F

[FeatureExt with AZURESQLSUPPORT Feature Data
example](#featureext-with-azuresqlsupport-feature-data) 204

[FeatureExt with SESSIONRECOVERY feature data
example](#featureext-with-sessionrecovery-feature-data) 190

[FeatureExtAck with AZURESQLSUPPORT Feature Data
example](#featureextack-with-azuresqlsupport-feature-data) 207

[FeatureExtAck with SESSIONRECOVERY feature data
example](#featureextack-with-sessionrecovery-feature-data) 195

[Fields - vendor-extensible](#vendor-extensible-fields) 18

Final state ([section 3.2.5.11](#final-state) 143, [section
3.3.5.11](#final-state-1) 153)

[Fixed-length token](#fixed-length-tokenxx11xxxx) 28

G

[Glossary](#glossary) 9

Grammar definition - token description

[data packet stream tokens](#data-packet-stream-tokens) 54

data stream types

[data type dependent data streams](#data-type-dependent-data-streams) 34

[unknown-length data streams](#unknown-length-data-streams) 33

[variable-length data streams](#variable-length-data-streams) 33

data type definitions

[fixed-length data types](#fixed-length-data-types) 38

[overview](#data-type-definitions) 37

[partially length-prefixed data
types](#partially-length-prefixed-data-types) 41

[SQL_VARIANT](#sql_variant-values) 45

[Table Valued Parameter](#table-valued-parameter-tvp-values) 46

[UDT Assembly Information](#common-language-runtime-clr-instances) 44

[variable-length data types](#variable-length-data-types) 38

[XML data type](#xml-values) 44

general rules

[collation rule definition](#collation-rule-definition) 32

[least significant bit order](#least-significant-bit-order) 32

[overview](#general-rules) 30

packet data stream headers

[overview](#packet-data-stream-headers---all_headers-rule-definition) 35

[Query Notifications header](#query-notifications-header) 36

[Transaction Descriptor header](#transaction-descriptor-header) 37

[TYPE_INFO rule definition](#type-info-rule-definition) 52

[Grammar Definition for Token Description
message](#grammar-definition-for-token-description) 30

H

Higher-layer triggered events

client ([section 3.1.4](#higher-layer-triggered-events) 127, [section
3.2.4](#higher-layer-triggered-events-1) 138)

server ([section 3.1.4](#higher-layer-triggered-events) 127, [section
3.3.4](#higher-layer-triggered-events-2) 147)

I

[Implementer - security
considerations](#security-considerations-for-implementers) 217

[Index of security parameters](#index-of-security-parameters) 217

[Informational messages](#error-and-info) 23

[Informative references](#informative-references) 12

[Initial state](#initial-state) 148

Initialization

client ([section 3.1.3](#initialization) 127, [section
3.2.3](#initialization-1) 137)

server ([section 3.1.3](#initialization) 127, [section
3.3.3](#initialization-2) 147)

[Introduction](#introduction) 9

L

Local events

client ([section 3.1.7](#other-local-events) 134, [section
3.2.7](#other-local-events-1) 144)

[server](#other-local-events) 134

Logged In state ([section 3.2.5.7](#logged-in-state) 142, [section
3.3.5.8](#logged-in-state-1) 152)

[Login Ready state](#login-ready-state) 149

[Login request example](#login-request) 155

[Login request with federated authentication
example](#login-request-with-federated-authentication) 157

[Login response example](#login-response-1) 164

[Login response with federated authentication
example](#login-response-with-federated-authentication-feature-extension-acknowledgement)
168

M

Message processing

client ([section 3.1.5](#message-processing-events-and-sequencing-rules)
127, [section 3.2.5](#message-processing-events-and-sequencing-rules-1)
139)

[Final state](#final-state) 143

[Logged In state](#logged-in-state) 142

[Sent Attention state](#sent-attention-state) 143

[Sent Client Request state](#sent-client-request-state) 143

[Sent Initial PRELOGIN Packet
state](#sent-initial-prelogin-packet-state) 139

[Sent LOGIN7 Record with SPNEGO Packet
state](#sent-login7-record-with-spnego-packet-state) 142

[Sent LOGIN7 Record with Standard Login
state](#sent-login7-record-with-complete-authentication-token-state) 141

[Sent TLS/SSL Negotiation Packet
state](#sent-tlsssl-negotiation-packet-state) 140

server ([section 3.1.5](#message-processing-events-and-sequencing-rules)
127, [section 3.3.5](#message-processing-events-and-sequencing-rules-2)
147)

[Client Request Execution state](#client-request-execution-state) 152

[Final state](#final-state-1) 153

[Initial state](#initial-state) 148

[Logged In state](#logged-in-state-1) 152

[Login Ready state](#login-ready-state) 149

[SPNEGO Negotiation state](#spnego-negotiation-state) 151

[TLS/SSL Negotiation state](#tls-negotiation-state) 148

Messages

[Client Messages](#client-messages) 19

[Grammar Definition for Token
Description](#grammar-definition-for-token-description) 30

[overview](#messages) 19

[Packet Data Token and Tokenless Data
Streams](#packet-data-token-and-tokenless-data-streams) 27

[Packet Data Token Stream
Definition](#packet-data-token-stream-definition) 85

[Packets](#packets) 23

[Server Messages](#server-messages) 21

syntax

[client messages](#client-messages) 19

[grammar definition for token
description](#grammar-definition-for-token-description) 30

[overview](#message-syntax) 19

[packet data token and tokenless data
streams](#packet-data-token-and-tokenless-data-streams) 27

[packet data token stream
definition](#packet-data-token-stream-definition) 85

[packet header message type - stream
definition](#packet-header-message-type-stream-definition) 55

[packets](#packets) 23

[server messages](#server-messages) 21

[transport](#transport) 19

N

[Normative references](#normative-references) 11

O

Other local events

[client](#other-local-events-1) 144

[server](#other-local-events-2) 153

[Overview (synopsis)](#overview) 14

P

Packet data - token stream definition

[ALTMETADATA](#altmetadata) 85

[ALTROW](#altrow) 87

[COLINFO](#colinfo) 88

[COLMETADATA](#colmetadata) 89

[DONE](#done) 95

[DONEINPROC](#doneinproc) 96

[DONEPROC](#doneproc) 97

[ENVCHANGE](#envchange) 99

[ERROR](#error) 103

[FEATUREEXTACK](#featureextack) 105

[INFO](#info) 112

[LOGINACK](#loginack) 113

[NBCROW](#nbcrow) 114

[OFFSET](#offset) 116

[ORDER](#order) 117

[RETURNSTATUS](#returnstatus) 117

[RETURNVALUE](#returnvalue) 118

[ROW](#row) 121

[SESSIONSTATE](#sessionstate) 122

[SSPI](#sspi) 123

[Table Valued Parameter row](#tvp_row) 125

[TABNAME](#tabname) 124

Packet data stream headers

[overview](#packet-data-stream-headers---all_headers-rule-definition) 35

[Query Notifications header](#query-notifications-header) 36

[Transaction Descriptor header](#transaction-descriptor-header) 37

[Packet Data Token and Tokenless Data Streams
message](#packet-data-token-and-tokenless-data-streams) 27

[Packet Data Token Stream Definition
message](#packet-data-token-stream-definition) 85

Packet header message type - stream definition

[BulkLoad - UpdateText/WriteText](#bulk-load-update-textwrite-text) 56

[BulkLoadBCP](#bulk-load-bcp) 55

[LOGIN7](#login7) 57

[PRELOGIN](#prelogin) 71

[RPCRequest](#rpc-request) 76

[SQLBatch](#sqlbatch) 79

[SSPI message](#sspi-message) 80

[transaction manager request](#transaction-manager-request-1) 81

Packets

[overview](#packets) 23

[packet data](#packet-data) 26

packet header

[Length](#length) 26

[overview](#packet-header) 24

[PacketID](#packetid) 26

[SPID](#spid) 26

[Status](#status) 25

[Type](#type) 24

[Window](#window) 26

[Packets message](#packets) 23

[Parameters - security index](#index-of-security-parameters) 217

[Preconditions](#prerequisitespreconditions) 16

[Pre-login request example](#pre-login-request) 154

[Prerequisites](#prerequisitespreconditions) 16

[Product behavior](#appendix-a-product-behavior) 219

Protocol Details

[overview](#protocol-details) 127

Q

[Query Notifications header](#query-notifications-header) 36

R

[References](#references) 11

[informative](#informative-references) 12

[normative](#normative-references) 11

[Relationship to other protocols](#relationship-to-other-protocols) 16

[Remote procedure call](#remote-procedure-call) 21

[RPC client request example](#rpc-client-request) 176

[RPC server response example](#rpc-server-response) 178

S

Security

[implementer considerations](#security-considerations-for-implementers)
217

[parameter index](#index-of-security-parameters) 217

[Sent Attention state](#sent-attention-state) 143

[Sent Client Request state](#sent-client-request-state) 143

[Sent Initial PRELOGIN Packet
state](#sent-initial-prelogin-packet-state) 139

[Sent LOGIN7 Record with SPNEGO Packet
state](#sent-login7-record-with-spnego-packet-state) 142

[Sent LOGIN7 Record with Standard Login
state](#sent-login7-record-with-complete-authentication-token-state) 141

[Sent TLS/SSL Negotiation Packet
state](#sent-tlsssl-negotiation-packet-state) 140

Sequencing rules

client ([section 3.1.5](#message-processing-events-and-sequencing-rules)
127, [section 3.2.5](#message-processing-events-and-sequencing-rules-1)
139)

[Final state](#final-state) 143

[Logged In state](#logged-in-state) 142

[Sent Attention state](#sent-attention-state) 143

[Sent Client Request state](#sent-client-request-state) 143

[Sent Initial PRELOGIN Packet
state](#sent-initial-prelogin-packet-state) 139

[Sent LOGIN7 Record with SPNEGO Packet
state](#sent-login7-record-with-spnego-packet-state) 142

[Sent LOGIN7 Record with Standard Login
state](#sent-login7-record-with-complete-authentication-token-state) 141

[Sent TLS/SSL Negotiation Packet
state](#sent-tlsssl-negotiation-packet-state) 140

server ([section 3.1.5](#message-processing-events-and-sequencing-rules)
127, [section 3.3.5](#message-processing-events-and-sequencing-rules-2)
147)

[Client Request Execution state](#client-request-execution-state) 152

[Final state](#final-state-1) 153

[Initial state](#initial-state) 148

[Logged In state](#logged-in-state-1) 152

[Login Ready state](#login-ready-state) 149

[SPNEGO Negotiation state](#spnego-negotiation-state) 151

[TLS/SSL Negotiation state](#tls-negotiation-state) 148

Server

abstract data model ([section 3.1.1](#abstract-data-model) 127, [section
3.3.1](#abstract-data-model-2) 146)

higher-layer triggered events ([section
3.1.4](#higher-layer-triggered-events) 127, [section
3.3.4](#higher-layer-triggered-events-2) 147)

initialization ([section 3.1.3](#initialization) 127, [section
3.3.3](#initialization-2) 147)

[local events](#other-local-events) 134

message processing ([section
3.1.5](#message-processing-events-and-sequencing-rules) 127, [section
3.3.5](#message-processing-events-and-sequencing-rules-2) 147)

[Client Request Execution state](#client-request-execution-state) 152

[Final state](#final-state-1) 153

[Initial state](#initial-state) 148

[Logged In state](#logged-in-state-1) 152

[Login Ready state](#login-ready-state) 149

[SPNEGO Negotiation state](#spnego-negotiation-state) 151

[TLS/SSL Negotiation state](#tls-negotiation-state) 148

messages

[attention acknowledgment](#attention-acknowledgment) 23

[error and informational messages](#error-and-info) 23

[login response](#login-response) 22

[overview](#server-messages) 21

[pre-login response](#pre-login-response) 22

[response completion](#response-completion) 23

[return parameters](#return-parameters) 23

[return status](#return-status) 22

[row data](#row-data) 22

[other local events](#other-local-events-2) 153

overview ([section 3.1](#common-details) 127, [section
3.3](#server-details) 144)

sequencing rules ([section
3.1.5](#message-processing-events-and-sequencing-rules) 127, [section
3.3.5](#message-processing-events-and-sequencing-rules-2) 147)

[Client Request Execution state](#client-request-execution-state) 152

[Final state](#final-state-1) 153

[Initial state](#initial-state) 148

[Logged In state](#logged-in-state-1) 152

[Login Ready state](#login-ready-state) 149

[SPNEGO Negotiation state](#spnego-negotiation-state) 151

[TLS/SSL Negotiation state](#tls-negotiation-state) 148

timer events ([section 3.1.6](#timer-events) 134, [section
3.3.6](#timer-events-2) 153)

timers ([section 3.1.2](#timers) 127, [section 3.3.2](#timers-2) 147)

[Server Messages message](#server-messages) 21

[SparseColumn select statement example](#sparsecolumn-select-statement)
185

[SPNEGO Negotiation state](#spnego-negotiation-state) 151

[SQL batch client request example](#sql-batch-client-request) 174

[SQL batch server response example](#sql-batch-server-response) 175

[SQL command](#sql-batch) 20

[SQL command with binary data](#bulk-load) 20

[SQL command with binary data example](#bulk-load-1) 180

[SSPI message example](#sspi-message-1) 179

[Standards assignments](#standards-assignments) 18

Syntax

client messages

[Attention](#attention) 21

[login](#login) 20

[overview](#client-messages) 19

[pre-login](#pre-login) 20

[remote procedure call](#remote-procedure-call) 21

[SQL command](#sql-batch) 20

[SQL command with binary data](#bulk-load) 20

[transaction manager request](#transaction-manager-request) 21

grammar definition for token description

[data packet stream tokens](#data-packet-stream-tokens) 54

[data stream types](#data-stream-types) 33

[data type definitions](#data-type-definitions) 37

[general rules](#general-rules) 30

[overview](#grammar-definition-for-token-description) 30

[packet data stream
headers](#packet-data-stream-headers---all_headers-rule-definition) 35

[TYPE_INFO rule definition](#type-info-rule-definition) 52

[overview](#message-syntax) 19

packet data token and tokenless data streams

[DONE and attention tokens](#done-and-attention-tokens) 29

[overview](#packet-data-token-and-tokenless-data-streams) 27

[token stream](#token-stream) 28

[tokenless stream](#tokenless-stream) 28

packet data token stream definition

[ALTMETADATA](#altmetadata) 85

[ALTROW](#altrow) 87

[COLINFO](#colinfo) 88

[COLMETADATA](#colmetadata) 89

[DONE](#done) 95

[DONEINPROC](#doneinproc) 96

[DONEPROC](#doneproc) 97

[ENVCHANGE](#envchange) 99

[ERROR](#error) 103

[FEATUREEXTACK](#featureextack) 105

[INFO](#info) 112

[LOGINACK](#loginack) 113

[NBCROW](#nbcrow) 114

[OFFSET](#offset) 116

[ORDER](#order) 117

[overview](#packet-data-token-stream-definition) 85

[RETURNSTATUS](#returnstatus) 117

[RETURNVALUE](#returnvalue) 118

[ROW](#row) 121

[SESSIONSTATE](#sessionstate) 122

[SSPI](#sspi) 123

[Table Valued Parameter row](#tvp_row) 125

[TABNAME](#tabname) 124

packet header message type - stream definition

[BulkLoad - UpdateText/WriteText](#bulk-load-update-textwrite-text) 56

[BulkLoadBCP](#bulk-load-bcp) 55

[LOGIN7](#login7) 57

[PRELOGIN](#prelogin) 71

[RPCRequest](#rpc-request) 76

[SQLBatch](#sqlbatch) 79

[SSPI message](#sspi-message) 80

[transaction manager request](#transaction-manager-request-1) 81

packets

[overview](#packets) 23

[packet data](#packet-data) 26

[packet header](#packet-header) 24

server messages

[attention acknowledgment](#attention-acknowledgment) 23

[error and informational messages](#error-and-info) 23

[login response](#login-response) 22

[overview](#server-messages) 21

[pre-login response](#pre-login-response) 22

[response completion](#response-completion) 23

[return parameters](#return-parameters) 23

[return status](#return-status) 22

[row data](#row-data) 22

T

[Table response with SESSIONSTATE token data
example](#table-response-with-sessionstate-token-data) 201

Timer events

client ([section 3.1.6](#timer-events) 134, [section
3.2.6](#timer-events-1) 144)

server ([section 3.1.6](#timer-events) 134, [section
3.3.6](#timer-events-2) 153)

Timers

client ([section 3.1.2](#timers) 127, [section 3.2.2](#timers-1) 137)

server ([section 3.1.2](#timers) 127, [section 3.3.2](#timers-2) 147)

[TLS/SSL Negotiation state](#tls-negotiation-state) 148

Token data stream

[overview](#token-stream) 28

token definition

[fixed-length token](#fixed-length-tokenxx11xxxx) 28

[overview](#token-definition) 28

[variable-count tokens](#variable-count-tokensxx00xxxx) 29

[variable-length tokens](#variable-length-tokensxx10xxxx) 29

[zero-length token](#zero-length-tokenxx01xxxx) 28

Token data stream definition

[ALTMETADATA](#altmetadata) 85

[ALTROW](#altrow) 87

[COLINFO](#colinfo) 88

[COLMETADATA](#colmetadata) 89

[DONE](#done) 95

[DONEINPROC](#doneinproc) 96

[DONEPROC](#doneproc) 97

[ENVCHANGE](#envchange) 99

[ERROR](#error) 103

[FEATUREEXTACK](#featureextack) 105

[INFO](#info) 112

[LOGINACK](#loginack) 113

[NBCROW](#nbcrow) 114

[OFFSET](#offset) 116

[ORDER](#order) 117

[overview](#packet-data-token-stream-definition) 85

[RETURNSTATUS](#returnstatus) 117

[RETURNVALUE](#returnvalue) 118

[ROW](#row) 121

[SESSIONSTATE](#sessionstate) 122

[SSPI](#sspi) 123

[Table Valued Parameter row](#tvp_row) 125

[TABNAME](#tabname) 124

Token data stream examples

[out-of-band attention signal](#out-of-band-attention-signal) 203

[overview](#token-stream-communication) 203

[sending an SQL batch](#sending-a-sql-batch) 203

Token description - grammar definition

[data packet stream tokens](#data-packet-stream-tokens) 54

data stream types

[data type dependent data streams](#data-type-dependent-data-streams) 34

[unknown-length data streams](#unknown-length-data-streams) 33

[variable-length data streams](#variable-length-data-streams) 33

data type definitions

[fixed-length data types](#fixed-length-data-types) 38

[overview](#data-type-definitions) 37

[partially length-prefixed data
types](#partially-length-prefixed-data-types) 41

[SQL_VARIANT](#sql_variant-values) 45

[Table Valued Parameter](#table-valued-parameter-tvp-values) 46

[UDT Assembly Information](#common-language-runtime-clr-instances) 44

[variable-length data types](#variable-length-data-types) 38

[XML data type](#xml-values) 44

general rules

[collation rule definition](#collation-rule-definition) 32

[least significant bit order](#least-significant-bit-order) 32

[overview](#general-rules) 30

[overview](#grammar-definition-for-token-description) 30

packet data stream headers

[overview](#packet-data-stream-headers---all_headers-rule-definition) 35

[Query Notifications header](#query-notifications-header) 36

[Transaction Descriptor header](#transaction-descriptor-header) 37

[TYPE_INFO rule definition](#type-info-rule-definition) 52

[Tokenless data stream](#tokenless-stream) 28

[Tracking changes](#change-tracking) 227

[Transaction Descriptor header](#transaction-descriptor-header) 37

[Transaction manager request](#transaction-manager-request) 21

[Transaction manager request example](#transaction-manager-request-2)
181

[Transport](#transport) 19

Triggered events - higher-layer

client ([section 3.1.4](#higher-layer-triggered-events) 127, [section
3.2.4](#higher-layer-triggered-events-1) 138)

server ([section 3.1.4](#higher-layer-triggered-events) 127, [section
3.3.4](#higher-layer-triggered-events-2) 147)

[TVP insert statement example](#tvp-insert-statement) 182

U

[Unknown-length data streams](#unknown-length-data-streams) 33

V

[Variable-count tokens](#variable-count-tokensxx00xxxx) 29

[Variable-length data streams](#variable-length-data-streams) 33

[Variable-length tokens](#variable-length-tokensxx10xxxx) 29

[Vendor-extensible fields](#vendor-extensible-fields) 18

[Versioning](#versioning-and-capability-negotiation) 17

Z

[Zero-length token](#zero-length-tokenxx01xxxx) 28
