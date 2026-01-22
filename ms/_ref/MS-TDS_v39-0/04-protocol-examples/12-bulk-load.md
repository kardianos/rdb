## Bulk Load

BULKLOADBCP request sent from client to server:

3150. 07 01 00 26 00 00 01 00 81 01 00 00 00 00 00 05

      00 32 02 63 00 31 00 D1 00 FD 00 00 00 00 00 00

      00 00 00 00 00 00

      \<PacketHeader\>

      \<Type\>

      \<BYTE\>07 \</BYTE\>

      \</Type\>

      \<Status\>

      \<BYTE\>01 \</BYTE\>

      \</Status\>

      \<Length\>

      \<BYTE\>00 \</BYTE\>

      \<BYTE\>26 \</BYTE\>

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

      \<BulkLoadBCP\>

      \<COLMETADATA\>

      \<TokenType\>

      \<BYTE\>81 \</BYTE\>

      \</TokenType\>

      \<Count\>

      \<USHORT\>01 00 \</USHORT\>

      \</Count\>

      \<ColumnData\>

      \<UserType\>

      \<ULONG\>00 00 00 00 \</ULONG\>

      \</UserType\>

      \<Flags\>

      \<USHORT\>05 00 \</USHORT\>

      \</Flags\>

      \<TYPE_INFO\>

      \<FIXEDLENTYPE\>

      \<BYTE\>32 \</BYTE\>

      \</FIXEDLENTYPE\>

      \</TYPE_INFO\>

      \<ColName\>

      \<B_VARCHAR\>

      \<BYTELEN\>

      \<BYTE\>02 \</BYTE\>

      \</BYTELEN\>

      \<BYTES ascii=\"c.1.\"\>63 00 31 00 \</BYTES\>

      \</B_VARCHAR\>

      \</ColName\>

      \</ColumnData\>

      \</COLMETADATA\>

      \<ROW\>

      \<TokenType\>

      \<BYTE\>D1 \</BYTE\>

      \</TokenType\>

      \<TYPE_VARBYTE\>

      \<BYTES\>00 \</BYTES\>

      \</TYPE_VARBYTE\>

      \</ROW\>

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

      \</BulkLoadBCP\>

      \</PacketData\>

