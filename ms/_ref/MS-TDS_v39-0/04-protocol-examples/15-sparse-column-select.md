## SparseColumn Select Statement

SparseColumn select statement sent from client to server:

3487. 04 01 01 B9 00 00 01 00 81 02 00 00 00 00 00 09 00

      26 04 02 69 00 64 00 00 00 00 00 0B 04 F1 00 11 73

      00 70 00 61 00 72 00 73 00 65 00 50 00 72 00 6F 00

      70 00 65 00 72 00 74 00 79 00 53 00 65 00 74 00 D1

      04 01 00 00 00 FE FF FF FF FF FF FF FF 7A 00 00 00

      3C 00 73 00 70 00 61 00 72 00 73 00 65 00 50 00 72

      00 6F 00 70 00 31 00 3E 00 31 00 30 00 30 00 30 00

      3C 00 2F 00 73 00 70 00 61 00 72 00 73 00 65 00 50

      00 72 00 6F 00 70 00 31 00 3E 00 3C 00 73 00 70 00

      61 00 72 00 73 00 65 00 50 00 72 00 6F 00 70 00 32

      00 3E 00 66 00 6F 00 6F 00 3C 00 2F 00 73 00 70 00

      61 00 72 00 73 00 65 00 50 00 72 00 6F 00 70 00 32

      00 3E 00 00 00 00 00 D1 04 02 00 00 00 FE FF FF FF

      FF FF FF FF 3E 00 00 00 3C 00 73 00 70 00 61 00 72

      00 73 00 65 00 50 00 72 00 6F 00 70 00 31 00 3E 00

      31 00 30 00 30 00 30 00 3C 00 2F 00 73 00 70 00 61

      00 72 00 73 00 65 00 50 00 72 00 6F 00 70 00 31

      00 3E 00 00 00 00 00 D1 04 03 00 00 00 FE FF FF

      FF FF FF FF FF 3E 00 00 00 3C 00 73 00 70 00 61

      00 72 00 73 00 65 00 50 00 72 00 6F 00 70 00 32

      00 3E 00 61 00 62 00 63 00 64 00 3C 00 2F 00 73

      00 70 00 61 00 72 00 73 00 65 00 50 00 72 00 6F

      00 70 00 32 00 3E 00 00 00 00 00 D2 02 04 04 00

      00 00 D2 02 04 05 00 00 00 D2 02 04 06 00 00 00

      D2 02 04 07 00 00 00 D2 02 04 08 00 00 00 D2 02

      04 09 00 00 00 D2 02 04 0A 00 00 00 FD 10 00 C1

      00 0A 00 00 00 00 00 00 00

      \<tds version=\"katmai\"\>

      \<PacketHeader\>

      \<Type\>

      \<BYTE\>04 \</BYTE\>

      \</Type\>

      \<Status\>

      \<BYTE\>01 \</BYTE\>

      \</Status\>

      \<Length\>

      \<BYTE\>01 \</BYTE\>

      \<BYTE\>B9 \</BYTE\>

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

      \<USHORT\>02 00 \</USHORT\>

      \</Count\>

      \<ColumnData\>

      \<UserType\>

      \<ULONG\>00 00 00 00 \</ULONG\>

      \</UserType\>

      \<Flags\>

      \<USHORT\>09 00 \</USHORT\>

      \</Flags\>

      \<TYPE_INFO\>

      \<VARLENTYPE\>

      \<BYTELEN_TYPE\>

      \<BYTE\>26 \</BYTE\>

      \</BYTELEN_TYPE\>

      \</VARLENTYPE\>

      \<TYPE_VARLEN\>

      \<BYTELEN\>

      \<BYTE\>04 \</BYTE\>

      \</BYTELEN\>

      \</TYPE_VARLEN\>

      \</TYPE_INFO\>

      \<ColName\>

      \<B_VARCHAR\>

      \<BYTELEN\>

      \<BYTE\>02 \</BYTE\>

      \</BYTELEN\>

      \<BYTES ascii=\"i.d.\"\>69 00 64 00 \</BYTES\>

      \</B_VARCHAR\>

      \</ColName\>

      \</ColumnData\>

      \<ColumnData\>

      \<UserType\>

      \<ULONG\>00 00 00 00 \</ULONG\>

      \</UserType\>

      \<Flags fSparseColumn=\"true\"\>

      \<USHORT\>0B 04 \</USHORT\>

      \</Flags\>

      \<TYPE_INFO\>

      \<VARLENTYPE\>

      \<USHORTLEN_TYPE\>

      \<BYTE\>F1 \</BYTE\>

      \</USHORTLEN_TYPE\>

      \</VARLENTYPE\>

      \<XML_INFO\>

      \<SCHEMA_PRESENT\>

      \<BYTE\>00 \</BYTE\>

      \</SCHEMA_PRESENT\>

      \</XML_INFO\>

      \</TYPE_INFO\>

      \<ColName\>

      \<B_VARCHAR\>

      \<BYTELEN\>

      \<BYTE\>11 \</BYTE\>

      \</BYTELEN\>

      \<BYTES ascii=\"s.p.a.r.s.e.P.r.o.p.e.r.t.y.S.e.t.\"\>73 00 70 00
      61 00 72 00 73 00 65 00 50 00 72 00 6F 00 70 00 65 00 72 00 74 00
      79 00 53 00 65 00 74 00 \</BYTES\>

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

      \<BYTELEN\>

      \<BYTE\>04 \</BYTE\>

      \</BYTELEN\>

      \</TYPE_VARLEN\>

      \<BYTES\>01 00 00 00 \</BYTES\>

      \</TYPE_VARBYTE\>

      \<TYPE_VARBYTE\>

      \<BYTES\>FE FF FF FF FF FF FF FF 7A 00 00 00 3C 00 73 00 70 00 61
      00 72 00 73 00 65 00 50 00 72 00 6F 00 70 00 31 00 3E 00 31 00 30
      00 30 00 30 00 3C 00 2F 00 73 00 70 00 61 00 72 00 73 00 65 00 50
      00 72 00 6F 00 70 00 31 00 3E 00 3C 00 73 00 70 00 61 00 72 00 73
      00 65 00 50 00 72 00 6F 00 70 00 32 00 3E 00 66 00 6F 00 6F 00 3C
      00 2F 00 73 00 70 00 61 00 72 00 73 00 65 00 50 00 72 00 6F 00 70
      00 32 00 3E 00 00 00 00 00 \</BYTES\>

      \</TYPE_VARBYTE\>

      \</ROW\>

      \<ROW\>

      \<TokenType\>

      \<BYTE\>D1 \</BYTE\>

      \</TokenType\>

      \<TYPE_VARBYTE\>

      \<TYPE_VARLEN\>

      \<BYTELEN\>

      \<BYTE\>04 \</BYTE\>

      \</BYTELEN\>

      \</TYPE_VARLEN\>

      \<BYTES\>02 00 00 00 \</BYTES\>

      \</TYPE_VARBYTE\>

      \<TYPE_VARBYTE\>

      \<BYTES\>FE FF FF FF FF FF FF FF 3E 00 00 00 3C 00 73 00 70 00 61
      00 72 00 73 00 65 00 50 00 72 00 6F 00 70 00 31 00 3E 00 31 00 30
      00 30 00 30 00 3C 00 2F 00 73 00 70 00 61 00 72 00 73 00 65 00 50
      00 72 00 6F 00 70 00 31 00 3E 00 00 00 00 00 \</BYTES\>

      \</TYPE_VARBYTE\>

      \</ROW\>

      \<ROW\>

      \<TokenType\>

      \<BYTE\>D1 \</BYTE\>

      \</TokenType\>

      \<TYPE_VARBYTE\>

      \<TYPE_VARLEN\>

      \<BYTELEN\>

      \<BYTE\>04 \</BYTE\>

      \</BYTELEN\>

      \</TYPE_VARLEN\>

      \<BYTES\>03 00 00 00 \</BYTES\>

      \</TYPE_VARBYTE\>

      \<TYPE_VARBYTE\>

      \<BYTES\>FE FF FF FF FF FF FF FF 3E 00 00 00 3C 00 73 00 70 00 61
      00 72 00 73 00 65 00 50 00 72 00 6F 00 70 00 32 00 3E 00 61 00 62
      00 63 00 64 00 3C 00 2F 00 73 00 70 00 61 00 72 00 73 00 65 00 50
      00 72 00 6F 00 70 00 32 00 3E 00 00 00 00 00 \</BYTES\>

      \</TYPE_VARBYTE\>

      \</ROW\>

      \<NBCROW\>

      \<TokenType\>

      \<BYTE\>D2 \</BYTE\>

      \</TokenType\>

      \<NullBitMap\>

      \<BYTES\>02 \</BYTES\>

      \</NullBitMap\>

      \<TYPE_VARBYTE\>

      \<TYPE_VARLEN\>

      \<BYTELEN\>

      \<BYTE\>04 \</BYTE\>

      \</BYTELEN\>

      \</TYPE_VARLEN\>

      \<BYTES\>04 00 00 00 \</BYTES\>

      \</TYPE_VARBYTE\>

      \</NBCROW\>

      \<NBCROW\>

      \<TokenType\>

      \<BYTE\>D2 \</BYTE\>

      \</TokenType\>

      \<NullBitMap\>

      \<BYTES\>02 \</BYTES\>

      \</NullBitMap\>

      \<TYPE_VARBYTE\>

      \<TYPE_VARLEN\>

      \<BYTELEN\>

      \<BYTE\>04 \</BYTE\>

      \</BYTELEN\>

      \</TYPE_VARLEN\>

      \<BYTES\>05 00 00 00 \</BYTES\>

      \</TYPE_VARBYTE\>

      \</NBCROW\>

      \<NBCROW\>

      \<TokenType\>

      \<BYTE\>D2 \</BYTE\>

      \</TokenType\>

      \<NullBitMap\>

      \<BYTES\>02 \</BYTES\>

      \</NullBitMap\>

      \<TYPE_VARBYTE\>

      \<TYPE_VARLEN\>

      \<BYTELEN\>

      \<BYTE\>04 \</BYTE\>

      \</BYTELEN\>

      \</TYPE_VARLEN\>

      \<BYTES\>06 00 00 00 \</BYTES\>

      \</TYPE_VARBYTE\>

      \</NBCROW\>

      \<NBCROW\>

      \<TokenType\>

      \<BYTE\>D2 \</BYTE\>

      \</TokenType\>

      \<NullBitMap\>

      \<BYTES\>02 \</BYTES\>

      \</NullBitMap\>

      \<TYPE_VARBYTE\>

      \<TYPE_VARLEN\>

      \<BYTELEN\>

      \<BYTE\>04 \</BYTE\>

      \</BYTELEN\>

      \</TYPE_VARLEN\>

      \<BYTES\>07 00 00 00 \</BYTES\>

      \</TYPE_VARBYTE\>

      \</NBCROW\>

      \<NBCROW\>

      \<TokenType\>

      \<BYTE\>D2 \</BYTE\>

      \</TokenType\>

      \<NullBitMap\>

      \<BYTES\>02 \</BYTES\>

      \</NullBitMap\>

      \<TYPE_VARBYTE\>

      \<TYPE_VARLEN\>

      \<BYTELEN\>

      \<BYTE\>04 \</BYTE\>

      \</BYTELEN\>

      \</TYPE_VARLEN\>

      \<BYTES\>08 00 00 00 \</BYTES\>

      \</TYPE_VARBYTE\>

      \</NBCROW\>

      \<NBCROW\>

      \<TokenType\>

      \<BYTE\>D2 \</BYTE\>

      \</TokenType\>

      \<NullBitMap\>

      \<BYTES\>02 \</BYTES\>

      \</NullBitMap\>

      \<TYPE_VARBYTE\>

      \<TYPE_VARLEN\>

      \<BYTELEN\>

      \<BYTE\>04 \</BYTE\>

      \</BYTELEN\>

      \</TYPE_VARLEN\>

      \<BYTES\>09 00 00 00 \</BYTES\>

      \</TYPE_VARBYTE\>

      \</NBCROW\>

      \<NBCROW\>

      \<TokenType\>

      \<BYTE\>D2 \</BYTE\>

      \</TokenType\>

      \<NullBitMap\>

      \<BYTES\>02 \</BYTES\>

      \</NullBitMap\>

      \<TYPE_VARBYTE\>

      \<TYPE_VARLEN\>

      \<BYTELEN\>

      \<BYTE\>04 \</BYTE\>

      \</BYTELEN\>

      \</TYPE_VARLEN\>

      \<BYTES\>0A 00 00 00 \</BYTES\>

      \</TYPE_VARBYTE\>

      \</NBCROW\>

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

      \<LONGLONG\>0A 00 00 00 00 00 00 00 \</LONGLONG\>

      \</DoneRowCount\>

      \</DONE\>

      \</TableResponse\>

      \</PacketData\>

      \</tds\>

