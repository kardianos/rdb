## TVP Insert Statement

TVP insert statement sent from client to server:

3292. 03 01 00 52 00 00 01 00 16 00 00 00

      12 00 00 00 02 00 00 00 00 00 00 00

      00 00 00 00 00 01 03 00 66 00 6F 00

      6F 00 00 00 00 00 F3 00 03 64 00 62

      00 6F 00 07 74 00 76 00 70 00 74 00

      79 00 70 00 65 00 01 00 00 00 00 00

      00 00 26 01 00 00 01 01 02 00

      \<tds version=\"katmai\"\>

      \<PacketHeader\>

      \<Type\>

      \<BYTE\>03 \</BYTE\>

      \</Type\>

      \<Status\>

      \<BYTE\>01 \</BYTE\>

      \</Status\>

      \<Length\>

      \<BYTE\>00 \</BYTE\>

      \<BYTE\>52 \</BYTE\>

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

      \<RPCRequest\>

      \<All_HEADERS\>

      \<TotalLength\>

      \<DWORD\>16 00 00 00 \</DWORD\>

      \</TotalLength\>

      \<Header\>

      \<HeaderLength\>

      \<DWORD\>12 00 00 00 \</DWORD\>

      \</HeaderLength\>

      \<HeaderType\>

      \<USHORT\>02 00 \</USHORT\>

      \</HeaderType\>

      \<HeaderData\>

      \<MARS\>

      \<TransactionDescriptor\>

      \<ULONGLONG\>00 00 00 00 00 00 00 00 \</ULONGLONG\>

      \</TransactionDescriptor\>

      \<OutstandingRequestCount\>

      \<DWORD\>00 00 00 01 \</DWORD\>

      \</OutstandingRequestCount\>

      \</MARS\>

      \</HeaderData\>

      \</Header\>

      \</All_HEADERS\>

      \<RPCReqBatch\>

      \<NameLenProcID\>

      \<ProcName\>

      \<US_VARCHAR\>

      \<USHORTLEN\>03 00 \</USHORTLEN\>

      \<BYTES ascii=\"f.o.o.\"\>66 00 6F 00 6F 00 \</BYTES\>

      \</US_VARCHAR\>

      \</ProcName\>

      \</NameLenProcID\>

      \<OptionFlags\>

      \<fWithRecomp\>

      \<BIT\>0\</BIT\>

      \</fWithRecomp\>

      \<fNoMetaData\>

      \<BIT\>0\</BIT\>

      \</fNoMetaData\>

      \<fReuseMetaData\>

      \<BIT\>0\</BIT\>

      \</fReuseMetaData\>

      \</OptionFlags\>

      \<ParameterData\>

      \<ParamMetaData\>

      \<B_VARCHAR\>

      \<BYTELEN\>00 \</BYTELEN\>

      \<BYTES ascii=\"\"\>

      \</BYTES\>

      \</B_VARCHAR\>

      \<StatusFlags\>

      \<fByRefValue\>

      \<BIT\>0\</BIT\>

      \</fByRefValue\>

      \<fDefaultValue\>

      \<BIT\>0\</BIT\>

      \</fDefaultValue\>

      \<fEncrypted\>

      \<BIT\>0\</BIT\>

      \</fEncrypted\>

      \</StatusFlags\>

      \<TVP_TYPE_INFO\>

      \<TVPTYPE\>

      \<BYTE\>F3 \</BYTE\>

      \</TVPTYPE\>

      \<TVP_TYPENAME\>

      \<DbName\>

      \<B_VARCHAR\>

      \<BYTELEN\>

      \<BYTE\>00 \</BYTE\>

      \</BYTELEN\>

      \<BYTES ascii=\"\"\>

      \</BYTES\>

      \</B_VARCHAR\>

      \</DbName\>

      \<OwningSchema\>

      \<B_VARCHAR\>

      \<BYTELEN\>

      \<BYTE\>03 \</BYTE\>

      \</BYTELEN\>

      \<BYTES ascii=\"dbo\"\>64 00 62 00 6F 00 \</BYTES\>

      \</B_VARCHAR\>

      \</OwningSchema\>

      \<TypeName\>

      \<B_VARCHAR\>

      \<BYTELEN\>

      \<BYTE\>07 \</BYTE\>

      \</BYTELEN\>

      \<BYTES ascii=\"tvptype\"\>74 00 76 00 70 00 74 00 79 00 70 00 65
      00 \</BYTES\>

      \</B_VARCHAR\>

      \</TypeName\>

      \</TVP_TYPENAME\>

      \<TVP_COLMETADATA\>

      \<Count\>

      \<USHORT\>01 00 \</USHORT\>

      \</Count\>

      \<TvpColumnMetaData\>

      \<UserType\>

      \<ULONG\>00 00 00 00 \</ULONG\>

      \</UserType\>

      \<Flags\>

      \<USHORT\>00 00 \</USHORT\>

      \</Flags\>

      \<TYPE_INFO\>

      \<VARLENTYPE\>

      \<BYTELEN_TYPE\>

      \<BYTE\>26 \</BYTE\>

      \</BYTELEN_TYPE\>

      \</VARLENTYPE\>

      \<TYPE_VARLEN\>

      \<BYTELEN\>

      \<BYTE\>01 \</BYTE\>

      \</BYTELEN\>

      \</TYPE_VARLEN\>

      \</TYPE_INFO\>

      \<ColName\>

      \<B_VARCHAR\>

      \<BYTELEN\>

      \<BYTE\>00 \</BYTE\>

      \</BYTELEN\>

      \<BYTES ascii=\"\"\>

      \</BYTES\>

      \</B_VARCHAR\>

      \</ColName\>

      \</TvpColumnMetaData\>

      \</TVP_COLMETADATA\>

      \<TVP_END_TOKEN\>

      \<TokenType\>

      \<BYTE\>00 \</BYTE\>

      \</TokenType\>

      \</TVP_END_TOKEN\>

      \<TVP_ROW\>

      \<TokenType\>

      \<BYTE\>01 \</BYTE\>

      \</TokenType\>

      \<AllColumnData\>

      \<TYPE_VARBYTE\>

      \<TYPE_VARLEN\>

      \<BYTELEN\>

      \<BYTE\>01\</BYTE\>

      \</BYTELEN\>

      \<BYTES\>02\</BYTES\>

      \</TYPE_VARLEN\>

      \</TYPE_VARBYTE\>

      \</AllColumnData\>

      \</TVP_ROW\>

      \<TVP_END_TOKEN\>

      \<TokenType\>

      \<BYTE\>00 \</BYTE\>

      \</TokenType\>

      \</TVP_END_TOKEN\>

      \</TVP_TYPE_INFO\>

      \</ParamMetaData\>

      \<ParamLenData\>

      \</ParamLenData\>

      \</ParameterData\>

      \</RPCReqBatch\>

      \</RPCRequest\>

      \</PacketData\>

      \</tds\>

