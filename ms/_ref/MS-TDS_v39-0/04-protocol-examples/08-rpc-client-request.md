## RPC Client Request

RPC request sent from the client to the server:

2904. 03 01 00 2F 00 00 01 00 16 00 00 00 12 00 00 00

      02 00 00 00 00 00 00 00 00 01 00 00 00 00 04 00

      66 00 6F 00 6F 00 33 00 00 00 00 02 26 02 00

      \<PacketHeader\>

      \<Type\>

      \<BYTE\>03 \</BYTE\>

      \</Type\>

      \<Status\>

      \<BYTE\>01 \</BYTE\>

      \</Status\>

      \<Length\>

      \<BYTE\>00 \</BYTE\>

      \<BYTE\>2F \</BYTE\>

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

      \<ULONGLONG\>00 00 00 00 00 00 00 01 \</ULONGLONG\>

      \</TransactionDescriptor\>

      \<OutstandingRequestCount\>

      \<DWORD\>00 00 00 00 \</DWORD\>

      \</OutstandingRequestCount\>

      \</MARS\>

      \</HeaderData\>

      \</Header\>

      \</All_HEADERS\>

      \<RPCReqBatch\>

      \<NameLenProcID\>

      \<ProcName\>

      \<US_VARCHAR\>

      \<USHORTLEN\>

      \<USHORT\>04 00 \</USHORT\>

      \</USHORTLEN\>

      \<BYTES ascii=\"f.o.o.3.\"\>66 00 6F 00 6F 00 33 00 \</BYTES\>

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

      \<BYTELEN\>

      \<BYTE\>00 \</BYTE\>

      \</BYTELEN\>

      \<BYTES ascii=\"\"\>

      \</BYTES\>

      \</B_VARCHAR\>

      \<StatusFlags\>

      \<fByRefValue\>

      \<BIT\>0\</BIT\>

      \</fByRefValue\>

      \<fDefaultValue\>

      \<BIT\>1\</BIT\>

      \</fDefaultValue\>

      \</StatusFlags\>

      \<TYPE_INFO\>

      \<VARLENTYPE\>

      \<BYTELEN_TYPE\>

      \<BYTE\>26 \</BYTE\>

      \</BYTELEN_TYPE\>

      \</VARLENTYPE\>

      \<TYPE_VARLEN\>

      \<BYTELEN\>

      \<BYTE\>02 \</BYTE\>

      \</BYTELEN\>

      \</TYPE_VARLEN\>

      \</TYPE_INFO\>

      \</ParamMetaData\>

      \<ParamLenData\>

      \<TYPE_VARBYTE\>

      \<TYPE_VARLEN\>

      \<BYTELEN\>

      \<BYTE\>00 \</BYTE\>

      \</BYTELEN\>

      \</TYPE_VARLEN\>

      \<BYTES\>

      \</BYTES\>

      \</TYPE_VARBYTE\>

      \</ParamLenData\>

      \</ParameterData\>

      \</RPCReqBatch\>

      \</RPCRequest\>

      \</PacketData\>

