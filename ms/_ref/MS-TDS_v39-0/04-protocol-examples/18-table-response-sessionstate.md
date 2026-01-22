## Table Response with SESSIONSTATE Token Data

A response message that contains SESSIONSTATE token data:

4531. 04 01 00 32 00 00 01 00 FD 01 00 BE 00 00 00 00

      00 00 00 00 00 E4 0B 00 00 00 01 00 00 00 01 09

      04 FF FF FF FF FD 00 00 FD 00 00 00 00 00 00 00

      00 00

      \<tds version=\"latest\"\>

      \<PacketHeader\>

      \<Type\>

      \<BYTE\>04 \</BYTE\>

      \</Type\>

      \<Status\>

      \<BYTE\>01 \</BYTE\>

      \</Status\>

      \<Length\>

      \<BYTE\>00 \</BYTE\>

      \<BYTE\>32 \</BYTE\>

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

      \<DONE\>

      \<TokenType\>

      \<BYTE\>FD \</BYTE\>

      \</TokenType\>

      \<Status\>

      \<USHORT\>01 00 \</USHORT\>

      \</Status\>

      \<CurCmd\>

      \<USHORT\>BE 00 \</USHORT\>

      \</CurCmd\>

      \<DoneRowCount\>

      \<LONGLONG\>00 00 00 00 00 00 00 00 \</LONGLONG\>

      \</DoneRowCount\>

      \</DONE\>

      \<SESSIONSTATE\>

      \<TokenType\>

      \<BYTE\>E4 \</BYTE\>

      \</TokenType\>

      \<Length\>

      \<DWORD\>0B 00 00 00 \</DWORD\>

      \</Length\>

      \<SeqNo\>

      \<DWORD\>01 00 00 00 \</DWORD\>

      \</SeqNo\>

      \<Status\>

      \<BYTE\>01 \</BYTE\>

      \</Status\>

      \<SessionStateDataSet\>

      \<SessionStateData\>

      \<StateId\>

      \<BYTE\>09 \</BYTE\>

      \</StateId\>

      \<StateLen\>

      \<BYTE\>04 \</BYTE\>

      \</StateLen\>

      \<StateValue\>

      \<BYTES\>FF FF FF FF \</BYTES\>

      \</StateValue\>

      \</SessionStateData\>

      \</SessionStateDataSet\>

      \</SESSIONSTATE\>

      \<DONE\>

      \<TokenType\>

      \<BYTE\>FD \</BYTE\>

      \</TokenType\>

      \<Status\>

      \<USHORT\>00 00 \</USHORT\>

      \</Status\>

      \<CurCmd\>

      \<USHORT\>FD 00 \</USHORT\>

      \</CurCmd\>

      \<DoneRowCount\>

      \<LONGLONG\>00 00 00 00 00 00 00 00 \</LONGLONG\>

      \</DoneRowCount\>

      \</DONE\>

      \</TableResponse\>

      \</PacketData\>

      \</tds\>

