## SQL Batch Client Request

Client request sent from the client to the server:

2744. 01 01 00 5C 00 00 01 00 16 00 00 00 12 00 00 00

      02 00 00 00 00 00 00 00 00 01 00 00 00 00 0A 00

      73 00 65 00 6C 00 65 00 63 00 74 00 20 00 27 00

      66 00 6F 00 6F 00 27 00 20 00 61 00 73 00 20 00

      27 00 62 00 61 00 72 00 27 00 0A 00 20 00 20 00

      20 00 20 00 20 00 20 00 20 00 20 00

      \<PacketHeader\>

      \<Type\>

      \<BYTE\>01 \</BYTE\>

      \</Type\>

      \<Status\>

      \<BYTE\>01 \</BYTE\>

      \</Status\>

      \<Length\>

      \<BYTE\>00 \</BYTE\>

      \<BYTE\>5C \</BYTE\>

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

      \<SQLBatch\>

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

      \<SQLText\>

      \<UNICODESTREAM\>

      \<BYTES\>0A 00 73 00 65 00 6C 00 65 00 63 00 74 00 20 00 27 00 66

      00 6F 00 6F 00 27 00 20 00 61 00 73 00 20 00 27 00 62 00 61 00 72
      00 27 00

      0A 00 20 00 20 00 20 00 20 00 20 00 20 00 20 00 20 00 \</BYTES\>

      \</UNICODESTREAM\>

      \</SQLText\>

      \</SQLBatch\>

      \</PacketData\>

