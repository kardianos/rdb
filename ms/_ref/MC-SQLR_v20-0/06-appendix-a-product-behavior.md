# Appendix A: Product Behavior

The information in this specification is applicable to the following
Microsoft products or supplemental software. References to product
versions include updates to those products.

-   Microsoft SQL Server 2000

-   Microsoft SQL Server 2005

-   Microsoft SQL Server 2008

-   Microsoft SQL Server 2008 R2

-   Microsoft SQL Server 2012

-   Microsoft SQL Server 2014

-   Microsoft SQL Server 2016

-   Microsoft SQL Server 2017

-   Microsoft SQL Server 2019

-   Microsoft SQL Server 2022

-   Microsoft SQL Server 2025

-   Windows XP operating system

-   Windows Server 2003 operating system

-   Windows Vista operating system

-   Windows Server 2008 operating system

-   Windows 7 operating system

-   Windows Server 2008 R2 operating system

-   Windows 8 operating system

-   Windows Server 2012 operating system

-   Windows 8.1 operating system

-   Windows Server 2012 R2 operating system

-   Windows 10 operating system

-   Windows Server 2016 operating system

-   Windows Server operating system

-   Windows Server 2019 operating system

-   Windows Server 2022 operating system

-   Windows 11 operating system

-   Windows Server 2025 operating system

Exceptions, if any, are noted in this section. If an update version,
service pack or Knowledge Base (KB) number appears with a product name,
the behavior changed in that update. The new behavior also applies to
subsequent updates unless otherwise specified. If a product edition
appears with the product version, behavior is different in that product
edition.

Unless otherwise specified, any statement of optional behavior in this
specification that is prescribed using the terms \"SHOULD\" or \"SHOULD
NOT\" implies product behavior in accordance with the SHOULD or SHOULD
NOT prescription. Unless otherwise specified, the term \"MAY\" implies
that the product does not follow the prescription.

[\<1\> Section 2.2.5](\l): The information for an instance of Microsoft
SQL Server can include the RPC_INFO, SPX_INFO, ADSP_INFO, and BV_INFO
tokens only if the version of the SQL Server instance is SQL Server
2000. SQL Server 2000 Browser, SQL Server 2005 Browser, SQL Server 2008
Browser, and SQL Server 2008 R2 Browser support sending information
about instances of SQL Server 2000 and will send these tokens.

[\<2\> Section 3.2.2](\l): Windows implements the timers for these two
messages as follows:

For the [CLNT_UCAST_INST](#Section_c97b04b5d80f4d3e919583bbfe246639)
request:

Windows implementations that use Microsoft Data Access Components (MDAC)
or Windows Data Access Components (Windows DAC) time out if no response
is received within 1 second. If a valid response is received within 1
second, the response is passed to the higher layer. If the response is
not valid, the process is repeated.

Windows implementations that use Microsoft SQL Server Native Client time
out if no response is received within 1 second. If a valid response is
received within 1 second, the response is immediately passed to the
higher layer. If the response is not valid, an error is passed to the
higher layer.

For the [CLNT_UCAST_DAC](#Section_20ebabbf46644f36bee04e3676a7aecd)
request:

Windows implementations that use MDAC or Windows DAC do not support this
request.

Windows implementations that use SQL Server Native Client time out if no
response is received within 1 second. If a valid response is received
within 1 second, the response is immediately passed to the higher layer.
If the response is not valid, an error is passed to the higher layer.

[\<3\> Section 3.2.2](\l): Windows implements the timers for these two
messages as follows:

For the [CLNT_UCAST_EX](#Section_ee0e41b0204f4a95b8bd5783a7c72cb2)
request:

Windows implementations that use Microsoft Data Access Components (MDAC)
or Windows Data Access Components (Windows DAC) time out if no response
is received within 0.5 second. If a valid response is received, it is
appended to the results. If the response is not valid, it is discarded.
The process is repeated until a time-out occurs.

Windows implementations that use SQL Server Native Client time out if no
response is received within the lesser of 5 seconds or the specified
logon time-out (the default logon time-out is 15 seconds.) If a valid
response is received, it is appended to the results. If the response is
not valid, it is discarded. The process is repeated for a maximum time
period of the lesser of 5 seconds or the specified logon time-out.

For the [CLNT_BCAST_EX](#Section_a3035afac2684699b8fd4f351e5c8e9e)
request:

Windows implementations that use MDAC or Windows DAC time out if no
response is received within 0.5 second. If a valid response is received,
it is appended to the results. If the response is not valid, it is
discarded. The process is repeated until a time-out occurs. There is no
maximum time-out limit.

Windows implementations that use SQL Server Native Client time out if no
response is received within 5 seconds and then each 1 second up to 15
seconds or to the specified logon time-out, if valid responses are not
received within each respective interval. If valid responses are
received, they are appended to the results; however, invalid responses
are discarded. The default logon time-out is 15 seconds.

[\<4\> Section 3.2.5.4](\l): Microsoft clients, such as Microsoft Data
Access Components (MDAC), Windows Data Access Components (Windows DAC),
or SQL Server Native Client, consider a SVR_RESP message to a
CLNT_UCAST_EX type request to be improperly formatted if the RESP_DATA
field is more than 4,096 bytes.

