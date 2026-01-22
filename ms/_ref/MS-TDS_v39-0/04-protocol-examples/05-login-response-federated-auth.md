## Login Response with Federated Authentication Feature Extension Acknowledgement

Login response from server to client, including the Feature Extension
Acknowledgement token that contains a federated authentication feature
extension option:

2396. 04 01 01 BC 01 4A 01 00 E3 1B 00 01 06 6D 00 61

      00 73 00 74 00 65 00 72 00 06 6D 00 61 00 73 00

      74 00 65 00 72 00 AB 62 00 45 16 00 00 02 00 25

      00 43 00 68 00 61 00 6E 00 67 00 65 00 64 00 20

      00 64 00 61 00 74 00 61 00 62 00 61 00 73 00 65

      00 20 00 63 00 6F 00 6E 00 74 00 65 00 78 00 74

      00 20 00 74 00 6F 00 20 00 27 00 6D 00 61 00 73

      00 74 00 65 00 72 00 27 00 2E 00 05 63 00 6C 00

      6F 00 75 00 64 00 00 01 00 00 00 E3 08 00 07 05

      09 04 D0 00 34 00 E3 17 00 02 0A 75 00 73 00 5F

      00 65 00 6E 00 67 00 6C 00 69 00 73 00 68 00 00

      AB 66 00 47 16 00 00 01 00 27 00 43 00 68 00 61

      00 6E 00 67 00 65 00 64 00 20 00 6C 00 61 00 6E

      00 67 00 75 00 61 00 67 00 65 00 20 00 73 00 65

      00 74 00 74 00 69 00 6E 00 67 00 20 00 74 00 6F

      00 20 00 75 00 73 00 5F 00 65 00 6E 00 67 00 6C

      00 69 00 73 00 68 00 2E 00 05 63 00 6C 00 6F 00

      75 00 64 00 00 01 00 00 00 AD 36 00 01 74 00 00

      04 16 4D 00 69 00 63 00 72 00 6F 00 73 00 6F 00

      66 00 74 00 20 00 53 00 51 00 4C 00 20 00 53 00

      65 00 72 00 76 00 65 00 72 00 00 00 00 00 0B 00

      08 CB E3 13 00 04 04 34 00 30 00 39 00 36 00 04

      34 00 30 00 39 00 36 00 AE 02 40 00 00 00 C9 08

      46 4E 58 49 0C 71 80 72 CD 69 F0 EC 3D E2 F6 ED

      75 8C 77 7D 9C B8 BB 87 4A 9C 90 80 A4 EE 40 B6

      07 71 0E A8 3C 0E D7 DE 14 DE F3 8B 65 C2 06 8C

      F9 51 D3 BC 32 55 15 A5 E4 A2 45 62 78 80 FF FD

      00 00 00 00 00 00 00 00 00 00 00 00

      \<PacketHeader\>

      \<Type\>

      \<BYTE\>04 \</BYTE\>

      \</Type\>

      \<Status\>

      \<BYTE\>01 \</BYTE\>

      \</Status\>

      \<Length\>

      \<BYTE\>01 \</BYTE\>

      \<BYTE\>BC \</BYTE\>

      \</Length\>

      \<SPID\>

      \<BYTE\>01 \</BYTE\>

      \<BYTE\>4A \</BYTE\>

      \</SPID\>

      \<PacketID\>

      \<BYTE\>01 \</BYTE\>

      \</PacketID\>

      \<Window\>

      \<BYTE\>00 \</BYTE\>

      \</Window\>

      \</PacketHeader\>

      \<PacketData\>

      \<TableResponse\>

      \<ENVCHANGE\>

      \<TokenType\>

      \<BYTE\>E3 \</BYTE\>

      \</TokenType\>

      \<Length\>

      \<USHORT\>1B 00 \</USHORT\>

      \</Length\>

      \<EnvValueData\>

      \<Type\>

      \<BYTE\>01 \</BYTE\>

      \</Type\>

      \<NewValue\>

      \<B_VARCHAR\>

      \<BYTELEN\>

      \<BYTE\>06 \</BYTE\>

      \</BYTELEN\>

      \<BYTES ascii=\"m.a.s.t.e.r.\"\>6D 00 61 00 73 00 74 00 65 00 72
      00 \</BYTES\>

      \</B_VARCHAR\>

      \</NewValue\>

      \<OldValue\>

      \<B_VARCHAR\>

      \<BYTELEN\>

      \<BYTE\>06 \</BYTE\>

      \</BYTELEN\>

      \<BYTES ascii=\"m.a.s.t.e.r.\"\>6D 00 61 00 73 00 74 00 65 00 72
      00 \</BYTES\>

      \</B_VARCHAR\>

      \</OldValue\>

      \</EnvValueData\>

      \</ENVCHANGE\>

      \<INFO\>

      \<TokenType\>

      \<BYTE\>AB \</BYTE\>

      \</TokenType\>

      \<Length\>

      \<USHORT\>62 00 \</USHORT\>

      \</Length\>

      \<Number\>

      \<LONG\>45 16 00 00 \</LONG\>

      \</Number\>

      \<State\>

      \<BYTE\>02 \</BYTE\>

      \</State\>

      \<Class\>

      \<BYTE\>00 \</BYTE\>

      \</Class\>

      \<MsgText\>

      \<US_VARCHAR\>

      \<USHORTLEN\>

      \<USHORT\>25 00 \</USHORT\>

      \</USHORTLEN\>

      \<BYTES ascii=\"C.h.a.n.g.e.d. .d.a.t.a.b.a.s.e. .c.o.n.t.e.x.t.

      .t.o. .\'.m.a.s.t.e.r.\'\...\"\>

      43 00 68 00 61 00 6E 00 67 00 65 00 64 00 20 00

      64 00 61 00 74 00 61 00 62 00 61 00 73 00 65 00

      20 00 63 00 6F 00 6E 00 74 00 65 00 78 00 74 00

      20 00 74 00 6F 00 20 00 27 00 6D 00 61 00 73 00

      74 00 65 00 72 00 27 00 2E 00 \</BYTES\>

      \</US_VARCHAR\>

      \</MsgText\>

      \<ServerName\>

      \<B_VARCHAR\>

      \<BYTELEN\>

      \<BYTE\>05 \</BYTE\>

      \</BYTELEN\>

      \<BYTES ascii=\"c.l.o.u.d.\"\>

      63 00 6C 00 6F 00 75 00 64 00 \</BYTES\>

      \</B_VARCHAR\>

      \</ServerName\>

      \<ProcName\>

      \<B_VARCHAR\>

      \<BYTELEN\>

      \<BYTE\>00 \</BYTE\>

      \</BYTELEN\>

      \<BYTES ascii=\"\"\>

      \</BYTES\>

      \</B_VARCHAR\>

      \</ProcName\>

      \<LineNumber\>

      \<LONG\>01 00 00 00 \</LONG\>

      \</LineNumber\>

      \</INFO\>

      \<ENVCHANGE\>

      \<TokenType\>

      \<BYTE\>E3 \</BYTE\>

      \</TokenType\>

      \<Length\>

      \<USHORT\>08 00 \</USHORT\>

      \</Length\>

      \<EnvValueData\>

      \<Type\>

      \<BYTE\>07 \</BYTE\>

      \</Type\>

      \<NewValue\>

      \<B_VARBYTE\>

      \<BYTELEN\>

      \<BYTE\>05 \</BYTE\>

      \</BYTELEN\>

      \<BYTES\>09 04 D0 00 34 \</BYTES\>

      \</B_VARBYTE\>

      \</NewValue\>

      \<OldValue\>

      \<B_VARCHAR\>

      \<BYTELEN\>

      \<BYTE\>00 \</BYTE\>

      \</BYTELEN\>

      \<BYTES ascii=\"\"\>

      \</BYTES\>

      \</B_VARCHAR\>

      \</OldValue\>

      \</EnvValueData\>

      \</ENVCHANGE\>

      \<ENVCHANGE\>

      \<TokenType\>

      \<BYTE\>E3 \</BYTE\>

      \</TokenType\>

      \<Length\>

      \<USHORT\>17 00 \</USHORT\>

      \</Length\>

      \<EnvValueData\>

      \<Type\>

      \<BYTE\>02 \</BYTE\>

      \</Type\>

      \<NewValue\>

      \<B_VARCHAR\>

      \<BYTELEN\>

      \<BYTE\>0A \</BYTE\>

      \</BYTELEN\>

      \<BYTES ascii=\"u.s.\_.e.n.g.l.i.s.h.\"\>75 00 73 00 5F 00 65 00
      6E 00 67 00 6C 00 69 00 73 00 68 00 \</BYTES\>

      \</B_VARCHAR\>

      \</NewValue\>

      \<OldValue\>

      \<B_VARCHAR\>

      \<BYTELEN\>

      \<BYTE\>00 \</BYTE\>

      \</BYTELEN\>

      \<BYTES ascii=\"\"\>

      \</BYTES\>

      \</B_VARCHAR\>

      \</OldValue\>

      \</EnvValueData\>

      \</ENVCHANGE\>

      \<INFO\>

      \<TokenType\>

      \<BYTE\>AB \</BYTE\>

      \</TokenType\>

      \<Length\>

      \<USHORT\>66 00 \</USHORT\>

      \</Length\>

      \<Number\>

      \<LONG\>47 16 00 00 \</LONG\>

      \</Number\>

      \<State\>

      \<BYTE\>01 \</BYTE\>

      \</State\>

      \<Class\>

      \<BYTE\>00 \</BYTE\>

      \</Class\>

      \<MsgText\>

      \<US_VARCHAR\>

      \<USHORTLEN\>

      \<USHORT\>27 00 \</USHORT\>

      \</USHORTLEN\>

      \<BYTES ascii=\"C.h.a.n.g.e.d. .l.a.n.g.u.a.g.e. .s.e.t.t.i.n.g.

      .t.o. .u.s.\_.e.n.g.l.i.s.h\...\"\>

      43 00 68 00 61 00 6E 00 67 00 65 00 64 00 20 00

      6C 00 61 00 6E 00 67 00 75 00 61 00 67 00 65 00

      20 00 73 00 65 00 74 00 74 00 69 00 6E 00 67 00

      20 00 74 00 6F 00 20 00 75 00 73 00 5F 00 65 00

      6E 00 67 00 6C 00 69 00 73 00 68 00 2E 00 \</BYTES\>

      \</US_VARCHAR\>

      \</MsgText\>

      \<ServerName\>

      \<B_VARCHAR\>

      \<BYTELEN\>

      \<BYTE\>05 \</BYTE\>

      \</BYTELEN\>

      \<BYTES ascii=\"c.l.o.u.d.\"\>

      63 00 6C 00 6F 00 75 00 64 00 \</BYTES\>

      \</B_VARCHAR\>

      \</ServerName\>

      \<ProcName\>

      \<B_VARCHAR\>

      \<BYTELEN\>

      \<BYTE\>00 \</BYTE\>

      \</BYTELEN\>

      \<BYTES ascii=\"\"\>

      \</BYTES\>

      \</B_VARCHAR\>

      \</ProcName\>

      \<LineNumber\>

      \<LONG\>01 00 00 00 \</LONG\>

      \</LineNumber\>

      \</INFO\>

      \<LOGINACK\>

      \<TokenType\>

      \<BYTE\>AD \</BYTE\>

      \</TokenType\>

      \<Length\>

      \<USHORT\>36 00 \</USHORT\>

      \</Length\>

      \<Interface\>

      \<BYTE\>01 \</BYTE\>

      \</Interface\>

      \<TDSVersion\>

      \<DWORD\>74 00 00 04 \</DWORD\>

      \</TDSVersion\>

      \<ProgName\>

      \<B_VARCHAR\>

      \<BYTELEN\>

      \<BYTE\>16 \</BYTE\>

      \</BYTELEN\>

      \<BYTES ascii=\"M.i.c.r.o.s.o.f.t. .S.Q.L.
      .S.e.r.v.e.r\.....\"\>4D

      00 69 00 63 00 72 00 6F 00 73 00 6F 00 66 00 74 00 20 00 53 00 51
      00 4C 00

      20 00 53 00 65 00 72 00 76 00 65 00 72 00 00 00 00 00 \</BYTES\>

      \</B_VARCHAR\>

      \</ProgName\>

      \<ProgVersion\>

      \<DWORD\>0B 00 08 CB \</DWORD\>

      \</ProgVersion\>

      \</LOGINACK\>

      \<ENVCHANGE\>

      \<TokenType\>

      \<BYTE\>E3 \</BYTE\>

      \</TokenType\>

      \<Length\>

      \<USHORT\>13 00 \</USHORT\>

      \</Length\>

      \<EnvValueData\>

      \<Type\>

      \<BYTE\>04 \</BYTE\>

      \</Type\>

      \<NewValue\>

      \<B_VARCHAR\>

      \<BYTELEN\>

      \<BYTE\>04 \</BYTE\>

      \</BYTELEN\>

      \<BYTES ascii=\"4.0.9.6\"\>34 00 30 00 39 00 36 00 \</BYTES\>

      \</B_VARCHAR\>

      \</NewValue\>

      \<OldValue\>

      \<B_VARCHAR\>

      \<BYTELEN\>

      \<BYTE\>04 \</BYTE\>

      \</BYTELEN\>

      \<BYTES ascii=\"4.0.9.6\"\>34 00 30 00 39 00 36 00 \</BYTES\>

      \</B_VARCHAR\>

      \</OldValue\>

      \</EnvValueData\>

      \</ENVCHANGE\>

      \<FEATUREEXTACK\>

      \<TokenType\>

      \<BYTE\>AE \</BYTE\>

      \</TokenType\>

      \<FeatureAckOpt\>

      \<FeatureId\>

      \<BYTE\>02 \</BYTE\>

      \</FeatureId\>

      \<FeatureAckDataLen\>

      \<DWORD\>40 00 00 00 \</DWORD\>

      \</FeatureAckDataLen\>

      \<FeatureAckData\>

      \<Nonce\>

      \<BYTES\>

      C9 08 46 4E 58 49 0C 71 80 72 CD 69 F0 EC 3D E2

      F6 ED 75 8C 77 7D 9C B8 BB 87 4A 9C 90 80 A4 EE \</BYTES\>

      \</Nonce\>

      \<Signature\>

      \<BYTES\>

      40 B6 07 71 0E A8 3C 0E D7 DE 14 DE F3 8B 65 C2

      06 8C F9 51 D3 BC 32 55 15 A5 E4 A2 45 62 78 80 \</BYTES\>

      \</Signature\>

      \</FeatureAckData\>

      \</FeatureAckOpt\>

      \<FeatureAckOpt\>

      \<TERMINATOR\>

      \<BYTE\>FF \</BYTE\>

      \</TERMINATOR\>

      \</FeatureAckOpt\>

      \</FEATUREEXTACK\>

      \<DONE\>

      \<TokenType\>

      \<BYTE\>FD \</BYTE\>

      \</TokenType\>

      \<Status\>

      \<USHORT\>00 00 \</USHORT\>

      \</Status\>

      \<CurCmd\>

      \<USHORT\>00 00 \</USHORT\>

      \</CurCmd\>

      \<DoneRowCount\>

      \<LONGLONG\>00 00 00 00 00 00 00 00 \</LONGLONG\>

      \</DoneRowCount\>

      \</DONE\>

      \</TableResponse\>

      \</PacketData\>

