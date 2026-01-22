## Login Request

LOGIN7 stream sent from the client to the server:

1498. 10 01 00 90 00 00 01 00 88 00 00 00 02 00 09 72

      00 10 00 00 00 00 00 07 00 01 00 00 00 00 00 00

      E0 03 00 00 00 00 00 00 09 04 00 00 5E 00 08 00

      6E 00 02 00 72 00 00 00 72 00 07 00 80 00 00 00

      80 00 00 00 80 00 04 00 88 00 00 00 88 00 00 00

      00 50 8B E2 B7 8F 88 00 00 00 88 00 00 00 88 00

      00 00 00 00 00 00 73 00 6B 00 6F 00 73 00 74 00

      6F 00 76 00 31 00 73 00 61 00 4F 00 53 00 51 00

      4C 00 2D 00 33 00 32 00 4F 00 44 00 42 00 43 00

      \<PacketHeader\>

      \<Type\>

      \<BYTE\>10 \</BYTE\>

      \</Type\>

      \<Status\>

      \<BYTE\>01 \</BYTE\>

      \</Status\>

      \<Length\>

      \<BYTE\>00 \</BYTE\>

      \<BYTE\>90 \</BYTE\>

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

      \<LOGIN7\>

      \<Length\>

      \<DWORD\>88 00 00 00 \</DWORD\>

      \</Length\>

      \<TDSVersion\>

      \<DWORD\>02 00 09 72 \</DWORD\>

      \</TDSVersion\>

      \<PacketSize\>

      \<DWORD\>00 10 00 00 \</DWORD\>

      \</PacketSize\>

      \<ClientProgVer\>

      \<DWORD\>00 00 00 07 \</DWORD\>

      \</ClientProgVer\>

      \<ClientPID\>

      \<DWORD\>00 01 00 00 \</DWORD\>

      \</ClientPID\>

      \<ConnectionID\>

      \<DWORD\>00 00 00 00 \</DWORD\>

      \</ConnectionID\>

      \<OptionFlags1\>

      \<BYTE\>E0 \</BYTE\>

      \</OptionFlags1\>

      \<OptionFlags2\>

      \<BYTE\>03 \</BYTE\>

      \</OptionFlags2\>

      \<TypeFlags\>

      \<BYTE\>00 \</BYTE\>

      \</TypeFlags\>

      \<OptionFlags3\>

      \<BYTE\>00 \</BYTE\>

      \</OptionFlags3\>

      \<ClientTimeZone\>

      \<LONG\>00 00 00 00 \</LONG\>

      \</ClientTimeZone\>

      \<ClientLCID\>

      \<DWORD\>09 04 00 00 \</DWORD\>

      \</ClientLCID\>

      \<OffsetLength\>

      \<ibHostName\>

      \<USHORT\>5E 00 \</USHORT\>

      \</ibHostName\>

      \<cchHostName\>

      \<USHORT\>08 00 \</USHORT\>

      \</cchHostName\>

      \<ibUserName\>

      \<USHORT\>6E 00 \</USHORT\>

      \</ibUserName\>

      \<cchUserName\>

      \<USHORT\>02 00 \</USHORT\>

      \</cchUserName\>

      \<ibPassword\>

      \<USHORT\>72 00 \</USHORT\>

      \</ibPassword\>

      \<cchPassword\>

      \<USHORT\>00 00 \</USHORT\>

      \</cchPassword\>

      \<ibAppName\>

      \<USHORT\>72 00 \</USHORT\>

      \</ibAppName\>

      \<cchAppName\>

      \<USHORT\>07 00 \</USHORT\>

      \</cchAppName\>

      \<ibServerName\>

      \<USHORT\>80 00 \</USHORT\>

      \</ibServerName\>

      \<cchServerName\>

      \<USHORT\>00 00 \</USHORT\>

      \</cchServerName\>

      \<ibUnused\>

      \<USHORT\>80 00 \</USHORT\>

      \</ibUnused\>

      \<cbUnused\>

      \<USHORT\>00 00 \</USHORT\>

      \</cbUnused\>

      \<ibCltIntName\>

      \<USHORT\>80 00 \</USHORT\>

      \</ibCltIntName\>

      \<cchCltIntName\>

      \<USHORT\>04 00 \</USHORT\>

      \</cchCltIntName\>

      \<ibLanguage\>

      \<USHORT\>88 00 \</USHORT\>

      \</ibLanguage\>

      \<cchLanguage\>

      \<USHORT\>00 00 \</USHORT\>

      \</cchLanguage\>

      \<ibDatabase\>

      \<USHORT\>88 00 \</USHORT\>

      \</ibDatabase\>

      \<cchDatabase\>

      \<USHORT\>00 00 \</USHORT\>

      \</cchDatabase\>

      \<ClientID\>

      \<BYTES\>00 50 8B E2 B7 8F \</BYTES\>

      \</ClientID\>

      \<ibSSPI\>

      \<USHORT\>88 00 \</USHORT\>

      \</ibSSPI\>

      \<cbSSPI\>

      \<USHORT\>00 00 \</USHORT\>

      \</cbSSPI\>

      \<ibAtchDBFile\>

      \<USHORT\>88 00 \</USHORT\>

      \</ibAtchDBFile\>

      \<cchAtchDBFile\>

      \<USHORT\>00 00 \</USHORT\>

      \</cchAtchDBFile\>

      \<ibChangePassword\>

      \<USHORT\>88 00 \</USHORT\>

      \</ibChangePassword\>

      \<cchChangePassword\>

      \<USHORT\>00 00 \</USHORT\>

      \</cchChangePassword\>

      \<cbSSPILong\>

      \<LONG\>00 00 00 00 \</LONG\>

      \</cbSSPILong\>

      \</OffsetLength\>

      \<Data\>

      \<BYTES\>73 00 6B 00 6F 00 73 00 74 00 6F 00 76 00 31 00 73 00 61
      00

      4F 00 53 00 51 00 4C 00 2D 00 33 00 32 00 4F 00 44 00 42 00 43 00
      \</BYTES\>

      \</Data\>

      \</LOGIN7\>

      \</PacketData\>

