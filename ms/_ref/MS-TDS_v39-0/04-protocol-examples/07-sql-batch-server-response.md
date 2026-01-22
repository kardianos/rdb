## SQL Batch Server Response

Server response sent from the server to the client:

2807. 04 01 00 33 00 00 01 00 81 01 00 00 00 00 00 20

      00 A7 03 00 09 04 D0 00 34 03 62 00 61 00 72 00

      D1 03 00 66 6F 6F FD 10 00 C1 00 01 00 00 00 00

      00 00 00

      \<PacketHeader\>

      \<Type\>

      \<BYTE\>04 \</BYTE\>

      \</Type\>

      \<Status\>

      \<BYTE\>01 \</BYTE\>

      \</Status\>

      \<Length\>

      \<BYTE\>00 \</BYTE\>

      \<BYTE\>33 \</BYTE\>

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

      \<USHORT\>20 00 \</USHORT\>

      \</Flags\>

      \<TYPE_INFO\>

      \<VARLENTYPE\>

      \<USHORTLEN_TYPE\>

      \<BYTE\>A7 \</BYTE\>

      \</USHORTLEN_TYPE\>

      \</VARLENTYPE\>

      \<TYPE_VARLEN\>

      \<USHORTCHARBINLEN\>

      \<USHORT\>03 00 \</USHORT\>

      \</USHORTCHARBINLEN\>

      \</TYPE_VARLEN\>

      \<COLLATION\>

      \<BYTES\>09 04 D0 00 34 \</BYTES\>

      \</COLLATION\>

      \</TYPE_INFO\>

      \<ColName\>

      \<B_VARCHAR\>

      \<BYTELEN\>

      \<BYTE\>03 \</BYTE\>

      \</BYTELEN\>

      \<BYTES ascii=\"b.a.r.\"\>62 00 61 00 72 00 \</BYTES\>

      \</B_VARCHAR\>

      \</ColName\>

      \</ColumnData\>

      \</COLMETADATA\>

      \<ROW\>

      \<TokenType\>

      \<BYTE\>D1 \</BYTE\>

      \</TokenType\>

      \<TYPE_VARBYTE\>

      \<TYPE_VARLEN\>

      \<USHORTCHARBINLEN\>

      \<USHORT\>03 00 \</USHORT\>

      \</USHORTCHARBINLEN\>

      \</TYPE_VARLEN\>

      \<BYTES ascii=\"fio\"\>66 6F 6F \</BYTES\>

      \</TYPE_VARBYTE\>

      \</ROW\>

      \<DONE\>

      \<TokenType\>

      \<BYTE\>FD \</BYTE\>

      \</TokenType\>

      \<Status\>

      \<USHORT\>10 00 \</USHORT\>

      \</Status\>

      \<CurCmd\>

      \<USHORT\>C1 00 \</USHORT\>

      \</CurCmd\>

      \<DoneRowCount\>

      \<LONGLONG\>01 00 00 00 00 00 00 00 \</LONGLONG\>

      \</DoneRowCount\>

      \</DONE\>

      \</TableResponse\>

      \</PacketData\>

