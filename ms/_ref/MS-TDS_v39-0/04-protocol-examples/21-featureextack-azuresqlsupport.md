## FeatureExtAck with AZURESQLSUPPORT Feature Data

A login response message that contains FeatureExtAck data for the
AZURESQLSUPPORT feature:

4888. 04 01 02 C3 00 77 01 00 FF 11 00 C1 00 01 00 00

      00 00 00 00 00 FF 11 00 C1 00 00 00 00 00 00 00

      00 00 FF 01 00 C0 00 00 00 00 00 00 00 00 00 FF

      11 00 C1 00 01 00 00 00 00 00 00 00 FF 11 00 C1

      00 00 00 00 00 00 00 00 00 FF 01 00 C0 00 00 00

      00 00 00 00 00 00 FF 11 00 C1 00 01 00 00 00 00

      00 00 00 FF 11 00 C1 00 01 00 00 00 00 00 00 00

      FF 11 00 C1 00 00 00 00 00 00 00 00 00 FF 01 00

      C0 00 00 00 00 00 00 00 00 00 FF 11 00 C1 00 01

      00 00 00 00 00 00 00 FF 11 00 C1 00 00 00 00 00

      00 00 00 00 FF 01 00 C0 00 00 00 00 00 00 00 00

      00 FF 11 00 C1 00 01 00 00 00 00 00 00 00 E3 1B

      00 01 06 74 00 65 00 73 00 74 00 64 00 62 00 06

      6D 00 61 00 73 00 74 00 65 00 72 00 AB 66 00 45

      16 00 00 02 00 25 00 43 00 68 00 61 00 6E 00 67

      00 65 00 64 00 20 00 64 00 61 00 74 00 61 00 62

      00 61 00 73 00 65 00 20 00 63 00 6F 00 6E 00 74

      00 65 00 78 00 74 00 20 00 74 00 6F 00 20 00 27

      00 74 00 65 00 73 00 74 00 64 00 62 00 27 00 2E

      00 07 74 00 65 00 73 00 74 00 73 00 76 00 72 00

      00 01 00 00 00 E3 08 00 07 05 09 04 D0 00 34 00

      E3 17 00 02 0A 75 00 73 00 5F 00 65 00 6E 00 67

      00 6C 00 69 00 73 00 68 00 00 AB 6A 00 47 16 00

      00 01 00 27 00 43 00 68 00 61 00 6E 00 67 00 65

      00 64 00 20 00 6C 00 61 00 6E 00 67 00 75 00 61

      00 67 00 65 00 20 00 73 00 65 00 74 00 74 00 69

      00 6E 00 67 00 20 00 74 00 6F 00 20 00 75 00 73

      00 5F 00 65 00 6E 00 67 00 6C 00 69 00 73 00 68

      00 2E 00 07 74 00 65 00 73 00 74 00 73 00 76 00

      72 00 00 01 00 00 00 AD 36 00 01 74 00 00 04 16

      4D 00 69 00 63 00 72 00 6F 00 73 00 6F 00 66 00

      74 00 20 00 53 00 51 00 4C 00 20 00 53 00 65 00

      72 00 76 00 65 00 72 00 00 00 00 00 0C 00 03 E8

      E3 13 00 04 04 38 00 30 00 30 00 30 00 04 34 00

      30 00 39 00 36 00 AE 01 77 00 00 00 00 09 00 60

      81 14 FF E7 FF FF 00 02 02 07 01 04 01 00 05 04

      FF FF FF FF 06 01 00 07 01 02 08 08 00 00 00 00

      00 00 00 00 09 04 FF FF FF FF 0B 47 35 00 44 00

      37 00 45 00 44 00 37 00 30 00 42 00 2D 00 42 00

      39 00 32 00 45 00 2D 00 34 00 31 00 32 00 42 00

      2D 00 42 00 33 00 32 00 46 00 2D 00 37 00 36 00

      30 00 43 00 44 00 37 00 34 00 44 00 42 00 39 00

      32 00 43 04 01 00 00 00 01 05 01 00 00 00 01 08

      01 00 00 00 01 FF FD 00 00 00 00 00 00 00 00 00

      00 00 00

      \<tds version=\"latest\"\>

      \<PacketHeader\>

      \<Type\>

      \<BYTE\>04 \</BYTE\>

      \</Type\>

      \<Status\>

      \<BYTE\>01 \</BYTE\>

      \</Status\>

      \<Length\>

      \<BYTE\>02 \</BYTE\>

      \<BYTE\>C3 \</BYTE\>

      \</Length\>

      \<SPID\>

      \<BYTE\>00 \</BYTE\>

      \<BYTE\>77 \</BYTE\>

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

      \<DONEINPROC\>

      \<TokenType\>

      \<BYTE\>FF \</BYTE\>

      \</TokenType\>

      \<Status\>

      \<USHORT\>11 00 \</USHORT\>

      \</Status\>

      \<CurCmd\>

      \<USHORT\>C1 00 \</USHORT\>

      \</CurCmd\>

      \<DoneRowCount\>

      \<LONGLONG\>01 00 00 00 00 00 00 00 \</LONGLONG\>

      \</DoneRowCount\>

      \</DONEINPROC\>

      \<DONEINPROC\>

      \<TokenType\>

      \<BYTE\>FF \</BYTE\>

      \</TokenType\>

      \<Status\>

      \<USHORT\>11 00 \</USHORT\>

      \</Status\>

      \<CurCmd\>

      \<USHORT\>C1 00 \</USHORT\>

      \</CurCmd\>

      \<DoneRowCount\>

      \<LONGLONG\>00 00 00 00 00 00 00 00 \</LONGLONG\>

      \</DoneRowCount\>

      \</DONEINPROC\>

      \<DONEINPROC\>

      \<TokenType\>

      \<BYTE\>FF \</BYTE\>

      \</TokenType\>

      \<Status\>

      \<USHORT\>01 00 \</USHORT\>

      \</Status\>

      \<CurCmd\>

      \<USHORT\>C0 00 \</USHORT\>

      \</CurCmd\>

      \<DoneRowCount\>

      \<LONGLONG\>00 00 00 00 00 00 00 00 \</LONGLONG\>

      \</DoneRowCount\>

      \</DONEINPROC\>

      \<DONEINPROC\>

      \<TokenType\>

      \<BYTE\>FF \</BYTE\>

      \</TokenType\>

      \<Status\>

      \<USHORT\>11 00 \</USHORT\>

      \</Status\>

      \<CurCmd\>

      \<USHORT\>C1 00 \</USHORT\>

      \</CurCmd\>

      \<DoneRowCount\>

      \<LONGLONG\>01 00 00 00 00 00 00 00 \</LONGLONG\>

      \</DoneRowCount\>

      \</DONEINPROC\>

      \<DONEINPROC\>

      \<TokenType\>

      \<BYTE\>FF \</BYTE\>

      \</TokenType\>

      \<Status\>

      \<USHORT\>11 00 \</USHORT\>

      \</Status\>

      \<CurCmd\>

      \<USHORT\>C1 00 \</USHORT\>

      \</CurCmd\>

      \<DoneRowCount\>

      \<LONGLONG\>00 00 00 00 00 00 00 00 \</LONGLONG\>

      \</DoneRowCount\>

      \</DONEINPROC\>

      \<DONEINPROC\>

      \<TokenType\>

      \<BYTE\>FF \</BYTE\>

      \</TokenType\>

      \<Status\>

      \<USHORT\>01 00 \</USHORT\>

      \</Status\>

      \<CurCmd\>

      \<USHORT\>C0 00 \</USHORT\>

      \</CurCmd\>

      \<DoneRowCount\>

      \<LONGLONG\>00 00 00 00 00 00 00 00 \</LONGLONG\>

      \</DoneRowCount\>

      \</DONEINPROC\>

      \<DONEINPROC\>

      \<TokenType\>

      \<BYTE\>FF \</BYTE\>

      \</TokenType\>

      \<Status\>

      \<USHORT\>11 00 \</USHORT\>

      \</Status\>

      \<CurCmd\>

      \<USHORT\>C1 00 \</USHORT\>

      \</CurCmd\>

      \<DoneRowCount\>

      \<LONGLONG\>01 00 00 00 00 00 00 00 \</LONGLONG\>

      \</DoneRowCount\>

      \</DONEINPROC\>

      \<DONEINPROC\>

      \<TokenType\>

      \<BYTE\>FF \</BYTE\>

      \</TokenType\>

      \<Status\>

      \<USHORT\>11 00 \</USHORT\>

      \</Status\>

      \<CurCmd\>

      \<USHORT\>C1 00 \</USHORT\>

      \</CurCmd\>

      \<DoneRowCount\>

      \<LONGLONG\>01 00 00 00 00 00 00 00 \</LONGLONG\>

      \</DoneRowCount\>

      \</DONEINPROC\>

      \<DONEINPROC\>

      \<TokenType\>

      \<BYTE\>FF \</BYTE\>

      \</TokenType\>

      \<Status\>

      \<USHORT\>11 00 \</USHORT\>

      \</Status\>

      \<CurCmd\>

      \<USHORT\>C1 00 \</USHORT\>

      \</CurCmd\>

      \<DoneRowCount\>

      \<LONGLONG\>00 00 00 00 00 00 00 00 \</LONGLONG\>

      \</DoneRowCount\>

      \</DONEINPROC\>

      \<DONEINPROC\>

      \<TokenType\>

      \<BYTE\>FF \</BYTE\>

      \</TokenType\>

      \<Status\>

      \<USHORT\>01 00 \</USHORT\>

      \</Status\>

      \<CurCmd\>

      \<USHORT\>C0 00 \</USHORT\>

      \</CurCmd\>

      \<DoneRowCount\>

      \<LONGLONG\>00 00 00 00 00 00 00 00 \</LONGLONG\>

      \</DoneRowCount\>

      \</DONEINPROC\>

      \<DONEINPROC\>

      \<TokenType\>

      \<BYTE\>FF \</BYTE\>

      \</TokenType\>

      \<Status\>

      \<USHORT\>11 00 \</USHORT\>

      \</Status\>

      \<CurCmd\>

      \<USHORT\>C1 00 \</USHORT\>

      \</CurCmd\>

      \<DoneRowCount\>

      \<LONGLONG\>01 00 00 00 00 00 00 00 \</LONGLONG\>

      \</DoneRowCount\>

      \</DONEINPROC\>

      \<DONEINPROC\>

      \<TokenType\>

      \<BYTE\>FF \</BYTE\>

      \</TokenType\>

      \<Status\>

      \<USHORT\>11 00 \</USHORT\>

      \</Status\>

      \<CurCmd\>

      \<USHORT\>C1 00 \</USHORT\>

      \</CurCmd\>

      \<DoneRowCount\>

      \<LONGLONG\>00 00 00 00 00 00 00 00 \</LONGLONG\>

      \</DoneRowCount\>

      \</DONEINPROC\>

      \<DONEINPROC\>

      \<TokenType\>

      \<BYTE\>FF \</BYTE\>

      \</TokenType\>

      \<Status\>

      \<USHORT\>01 00 \</USHORT\>

      \</Status\>

      \<CurCmd\>

      \<USHORT\>C0 00 \</USHORT\>

      \</CurCmd\>

      \<DoneRowCount\>

      \<LONGLONG\>00 00 00 00 00 00 00 00 \</LONGLONG\>

      \</DoneRowCount\>

      \</DONEINPROC\>

      \<DONEINPROC\>

      \<TokenType\>

      \<BYTE\>FF \</BYTE\>

      \</TokenType\>

      \<Status\>

      \<USHORT\>11 00 \</USHORT\>

      \</Status\>

      \<CurCmd\>

      \<USHORT\>C1 00 \</USHORT\>

      \</CurCmd\>

      \<DoneRowCount\>

      \<LONGLONG\>01 00 00 00 00 00 00 00 \</LONGLONG\>

      \</DoneRowCount\>

      \</DONEINPROC\>

      \<ENVCHANGE\>

      \<TokenType\>

      \<BYTE\>E3\</BYTE\>

      \</TokenType\>

      \<Length\>

      \<USHORT\>1B 00\</USHORT\>

      \</Length\>

      \<EnvValueData\>

      \<Type type=\"Database\"\>

      \<BYTE\>01\</BYTE\>

      \</Type\>

      \<NewValue\>

      \<B_VARCHAR\>

      \<BYTELEN\>

      \<BYTE\>06 \</BYTE\>

      \</BYTELEN\>

      \<BYTES ascii=\"t.e.s.t.d.b.\"\>74 00 65 00 73 00 74 00 64 00 62
      00 \</BYTES\>

      \</B_VARCHAR\>

      \</NewValue\>

      \<OldValue\>

      \<B_VARCHAR\>

      \<BYTELEN\>

      \<BYTE\>06 \</BYTE\>

      \</BYTELEN\>

      \<BYTES ascii=\"m.a.s.t.e.r.\"\>6D 00 61 00 73 00 74 00 65 00 72
      00\</BYTES\>

      \</B_VARCHAR\>

      \</OldValue\>

      \</EnvValueData\>

      \</ENVCHANGE\>

      \<INFO\>

      \<TokenType\>

      \<BYTE\>AB\</BYTE\>

      \</TokenType\>

      \<Length\>

      \<USHORT\>66 00\</USHORT\>

      \</Length\>

      \<Number\>

      \<LONG\>45 16 00 00

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
      .t.o. .\'.t.e.s.t.d.b.\'\...\"\> 43 00 68 00 61 00 6E 00

      67 00 65 00 64 00 20 00 64 00 61 00 74 00 61 00

      62 00 61 00 73 00 65 00 20 00 63 00 6F 00 6E 00

      74 00 65 00 78 00 74 00 20 00 74 00 6F 00 20 00

      27 00 74 00 65 00 73 00 74 00 64 00 62 00 27 00

      2E 00 \</BYTES\>

      \</US_VARCHAR\>

      \</MsgText\>

      \<ServerName\>

      \<B_VARCHAR\>

      \<BYTELEN\>

      \<BYTE\>07 \</BYTE\>

      \</BYTELEN\>

      \<BYTES ascii=\"t.e.s.t.s.v.r.\"\>74 00 65 00 73 00 74 00 73 00 76
      00 72

      00 \</BYTES\>

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

      \<LONG\>01 00 00 00\</LONG\>

      \</LineNumber\>

      \</INFO\>

      \<ENVCHANGE\>

      \<TokenType\>

      \<BYTE\>E3\</BYTE\>

      \</TokenType\>

      \<Length\>

      \<USHORT\>08 00 \</USHORT\>

      \</Length\>

      \<EnvValueData\>

      \<Type type=\"SQL Collation\"\>

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

      \<B_VARBYTE\>

      \<BYTELEN\>

      \<BYTE\>00\</BYTE\>

      \</BYTELEN\>

      \</B_VARBYTE\>

      \</OldValue\>

      \</ENVCHANGE\>

      \<ENVCHANGE\>

      \<TokenType\>

      \<BYTE\>E3\</BYTE\>

      \</TokenType\>

      \<Length\>

      \<USHORT\>17 00 \</USHORT\>

      \</Length\>

      \<EnvValueData\>

      \<Type type=\"language\"\>

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

      \<USHORT\>6A 00 \</USHORT\>

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
      .t.o. .\'.u.s.\_.e.n.g.l.i.s.h.\'\...\"\>

      43 00 68 00 61 00 6E 00

      67 00 65 00 64 00 20 00 6C 00 61 00 6E 00 67 00

      75 00 61 00 67 00 65 00 20 00 73 00 65 00 74 00

      74 00 69 00 6E 00 67 00 20 00 74 00 6F 00 20 00

      75 00 73 00 5F 00 65 00 6E 00 67 00 6C 00 69 00

      73 00 68 00 2E 00 \</BYTES\>

      \</US_VARCHAR\>

      \</MsgText\>

      \<ServerName\>

      \<B_VARCHAR\>

      \<BYTELEN\>

      \<BYTE\>07 \</BYTE\>

      \</BYTELEN\>

      \<BYTES ascii=\"t.e.s.t.s.v.r.\"\>74 00 65 00 73 00 74 00 73

      00 76 00 72 00 \</BYTES\>

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

      \<BYTE\>AD\</BYTE\>

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

      \<BYTES ascii=\"M.i.c.r.o.s.o.f.t. .S.Q.L. .S.e.r.v.e.r\.....\"\>

      4D 00 69 00 63 00 72 00 6F 00 73 00 6F 00 66 00

      74 00 20 00 53 00 51 00 4C 00 20 00 53 00 65 00

      72 00 76 00 65 00 72 00 00 00 00 00

      \</BYTES\>

      \</B_VARCHAR\>

      \</ProgName\>

      \<PROGVERSION\>

      \<DWORD\>0C 00 03 E8 \</DWORD\>

      \</PROGVERSION\>

      \</LOGINACK\>

      \<ENVCHANGE\>

      \<TokenType\>

      \<BYTE\>E3 \</BYTE\>

      \</TokenType\>

      \<Length\>

      \<USHORT\>13 00 \</USHORT\>

      \</Length\>

      \<EnvValueData\>

      \<Type type=\"Packet size\"\>

      \<BYTE\>04 \</BYTE\>

      \</Type\>

      \<DATA\>

      \<NewValue\>

      \<B_VARCHAR\>

      \<BYTELEN\>

      \<BYTE\>04 \</BYTE\>

      \</BYTELEN\>

      \<BYTES ascii=\"8.0.0.0.\"\>38 00 30 00 30 00 30 00 \</BYTES\>

      \</B_VARCHAR\>

      \</NewValue\>

      \<OldValue\>

      \<B_VARCHAR\>

      \<BYTELEN\>

      \<BYTE\>04 \</BYTE\>

      \<BYTES ascii=\"4.0.9.6.\"\>34 00 30 00 39 00 36 00 \</BYTES\>

      \</BYTES\>

      \</B_VARCHAR\>

      \</OldValue\>

      \</EnvValueData\>

      \</ENVCHANGE\>

      \<FeatureExtAck\>

      \<TokenType\>

      \<BYTE\>AE \</BYTE\>

      \</TokenType\>

      \<FeatureAckOpt\>

      \<FeatureId\>

      \<BYTE\>01 \</BYTE\>

      \</FeatureId\>

      \<FeatureAckDataLen\>

      \<DWORD\>77 00 00 00 \</DWORD\>

      \</FeatureAckDataLen\>

      \<FeatureAckData\>

      \<BYTE\>

      00 09 00 60 81 14 FF E7 FF FF 00 02 02 07 01 04

      01 00 05 04 FF FF FF FF 06 01 00 07 01 02 08 08

      00 00 00 00 00 00 00 00 09 04 FF FF FF FF 0B 47

      35 00 44 00 37 00 45 00 44 00 37 00 30 00 42 00

      2D 00 42 00 39 00 32 00 45 00 2D 00 34 00 31 00

      32 00 42 00 2D 00 42 00 33 00 32 00 46 00 2D 00

      37 00 36 00 30 00 43 00 44 00 37 00 34 00 44 00

      42 00 39 00 32 00 43

      \</BYTE\>

      \</FeatureAckData\>

      \</FeatureAckOpt\>

      \<FeatureAckOpt\>

      \<FeatureId\>

      \<BYTE\>04 \</BYTE\>

      \</FeatureId\>

      \<FeatureAckDataLen\>

      \<DWORD\>01 00 00 00 \</DWORD\>

      \</FeatureAckDataLen\>

      \<FeatureAckData\>

      \<BYTE\>01\</BYTE\>

      \</FeatureAckData\>

      \</FeatureAckOpt\>

      \<FeatureAckOpt\>

      \<FeatureId\>

      \<BYTE\>05 \</BYTE\>

      \</FeatureId\>

      \<FeatureAckDataLen\>

      \<DWORD\>01 00 00 00 \</DWORD\>

      \</FeatureAckDataLen\>

      \<FeatureAckData\>

      \<BYTE\>01\</BYTE\>

      \</FeatureAckData\>

      \</FeatureAckOpt\>

      \<FeatureAckOpt\>

      \<FeatureId\>

      \<AZURESQLSUPPORT\>

      \<BYTE\>08 \</BYTE\>

      \</AZURESQLSUPPORT\>

      \</FeatureId\>

      \<FeatureAckDataLen\>

      \<DWORD\>01 00 00 00 \</DWORD\>

      \</FeatureAckDataLen\>

      \<FeatureAckData\>

      \<BYTE\>01\</BYTE\>

      \</FeatureAckData\>

      \</FeatureAckOpt\>

      \<FeatureAckOpt\>

      \<TERMINATOR\>

      \<BYTE\>FF \</BYTE\>

      \</TERMINATOR\>

      \</FeatureAckOpt\>

      \</FeatureExtAck\>

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

      \</tds\>

