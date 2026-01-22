## Transport

The TDS protocol does not prescribe a specific underlying transport
protocol to use on the Internet or on other networks. TDS only presumes
a reliable transport that guarantees in-sequence delivery of data.

The chosen transport can be either stream-oriented or message-oriented.
If a message-oriented transport is used, any TDS packet sent from a TDS
client to a TDS server MUST be contained within a single transport data
unit. Any additional mapping of TDS data onto the transport data units
of the protocol in question is outside the scope of this specification.

The current version of the TDS protocol has implementations over the
following transports, except as indicated:[\<2\>](\l)

-   TCP [\[RFC793\]](https://go.microsoft.com/fwlink/?LinkId=150872).

-   A reliable transport over the [**Virtual Interface Architecture
    (VIA)**](#gt_c35909bd-185e-4b60-be82-995a0318873e) [**interface**](#gt_95913fbd-3262-47ae-b5eb-18e6806824b9) \[VIA2002\]
    can be used in only TDS 7.0, TDS 7.1, TDS 7.2, and TDS
    7.3.[\<3\>](\l)

-   Named Pipes
    [\[MSDN-NP\]](https://go.microsoft.com/fwlink/?LinkId=90247).

-   Shared memory
    [\[MSDN-TDSENDPT\]](https://go.microsoft.com/fwlink/?linkid=865399).

-   Optionally, the TDS protocol has implementations for the following
    protocols on top of the preceding transports:

    -   Transport Layer Security (TLS)/Secure Socket Layer (SSL)
        [\[RFC2246\]](https://go.microsoft.com/fwlink/?LinkId=90324)
        [\[RFC5246\]](https://go.microsoft.com/fwlink/?LinkId=129803)
        [\[RFC6101\]](https://go.microsoft.com/fwlink/?LinkId=509953),
        in case TLS/SSL encryption is negotiated in TDS 7.x.

    -   TLS
        [\[RFC8446\]](https://go.microsoft.com/fwlink/?linkid=2147431),
        in case TLS encryption is established in TDS 8.0.

    -   [**Session Multiplex Protocol
        (SMP)**](#gt_f70f98cc-c555-4a40-9509-bc1da4021211)
        [\[MC-SMP\]](%5bMC-SMP%5d.pdf#Section_04c8edde371d4af5bb33a39b3948f0af),
        in case the [**Multiple Active Result Sets
        (MARS)**](#gt_762fe1e3-0979-4402-b963-1e9150de133d) feature
        [\[MSDN-MARS\]](https://go.microsoft.com/fwlink/?LinkId=98459)
        is requested.

