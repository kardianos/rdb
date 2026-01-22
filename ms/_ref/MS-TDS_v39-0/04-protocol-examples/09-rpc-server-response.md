## RPC Server Response

RPC response sent from the server to the client:

3022. 04 01 00 27 00 00 01 00 FF 11 00 C1 00 01 00 00

      00 00 00 00 00 79 00 00 00 00 FE 00 00 E0 00 00

      00 00 00 00 00 00 00

      \<PacketHeader\>

      \<Type\>

      \<BYTE\>04 \</BYTE\>

      \</Type\>

      \<Status\>

      \<BYTE\>01 \</BYTE\>

      \</Status\>

      \<Length\>

      \<BYTE\>00 \</BYTE\>

      \<BYTE\>27 \</BYTE\>

      \</Length\>

      \<SPID\>

      \<BYTE\>00 \</BYTE\>

      \<BYTE\>00 \</BYTE\>

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

      \<RETURNSTATUS\>

      \<TokenType\>

      \<BYTE\>79 \</BYTE\>

      \</TokenType\>

      \<VALUE\>

      \<LONG\>00 00 00 00 \</LONG\>

      \</VALUE\>

      \</RETURNSTATUS\>

      \<DONEPROC\>

      \<TokenType\>

      \<BYTE\>FE \</BYTE\>

      \</TokenType\>

      \<Status\>

      \<USHORT\>00 00 \</USHORT\>

      \</Status\>

      \<CurCmd\>

      \<USHORT\>E0 00 \</USHORT\>

      \</CurCmd\>

      \<DoneRowCount\>

      \<LONGLONG\>00 00 00 00 00 00 00 00 \</LONGLONG\>

      \</DoneRowCount\>

      \</DONEPROC\>

      \</TableResponse\>

      \</PacketData\>

