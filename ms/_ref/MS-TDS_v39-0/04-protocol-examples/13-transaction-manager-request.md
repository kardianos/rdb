## Transaction Manager Request

Transaction Manager Request sent from client to server:

3232. 0E 01 00 20 00 00 01 00 16 00 00 00

      12 00 00 00 02 00 00 00 00 00 00 00

      00 01 00 00 00 00 16 00

      \<PacketHeader\>

      \<Type\>

      \<BYTE\>0E \</BYTE\>

      \</Type\>

      \<Status\>

      \<BYTE\>01 \</BYTE\>

      \</Status\>

      \<Length\>

      \<BYTE\>00 \</BYTE\>

      \<BYTE\>20 \</BYTE\>

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

      \<TransMgrReq\>

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

      \<RequestType\>

      \<USHORT\>16 00 \</USHORT\>

      \</RequestType\>

      \<RequestPayload\>

      \<TM_PROMOTE_XACT\>

      \</TM_PROMOTE_XACT\>

      \</RequestPayload\>

      \</TransMgrReq\>

      \</PacketData\>

