# Protocol Examples

The following are examples of the binary representation of various
client requests and the responses from the server.

## CLNT_UCAST_EX

Request:

12. 03

Response:

13. 05 47 01 53 65 72 76 65 72 4e 61 6d 65 3b 49 4c \| .G.ServerName;IL

    53 55 4e 47 31 3b 49 6e 73 74 61 6e 63 65 4e 61 \| SUNG1;InstanceNa

    6d 65 3b 59 55 4b 4f 4e 53 54 44 3b 49 73 43 6c \| me;YUKONSTD;IsCl

    75 73 74 65 72 65 64 3b 4e 6f 3b 56 65 72 73 69 \| ustered;No;Versi

    6f 6e 3b 39 2e 30 30 2e 31 33 39 39 2e 30 36 3b \| on;9.00.1399.06;

    74 63 70 3b 35 37 31 33 37 3b 3b 53 65 72 76 65 \| tcp;57137;;Serve

    72 4e 61 6d 65 3b 49 4c 53 55 4e 47 31 3b 49 6e \| rName;ILSUNG1;In

    73 74 61 6e 63 65 4e 61 6d 65 3b 59 55 4b 4f 4e \| stanceName;YUKON

    44 45 56 3b 49 73 43 6c 75 73 74 65 72 65 64 3b \| DEV;IsClustered;

    4e 6f 3b 56 65 72 73 69 6f 6e 3b 39 2e 30 30 2e \| No;Version;9.00.

    31 33 39 39 2e 30 36 3b 6e 70 3b 5c 5c 49 4c 53 \|
    1399.06;np;\\\\ILS

    55 4e 47 31 5c 70 69 70 65 5c 4d 53 53 51 4c 24 \|
    UNG1\\pipe\\MSSQL\$

    59 55 4b 4f 4e 44 45 56 5c 73 71 6c 5c 71 75 65 \|
    YUKONDEV\\sql\\que

    72 79 3b 3b 53 65 72 76 65 72 4e 61 6d 65 3b 49 \| ry;;ServerName;I

    4c 53 55 4e 47 31 3b 49 6e 73 74 61 6e 63 65 4e \| LSUNG1;InstanceN

    61 6d 65 3b 4d 53 53 51 4c 53 45 52 56 45 52 3b \| ame;MSSQLSERVER;

    49 73 43 6c 75 73 74 65 72 65 64 3b 4e 6f 3b 56 \| IsClustered;No;V

    65 72 73 69 6f 6e 3b 39 2e 30 30 2e 31 33 39 39 \| ersion;9.00.1399

    2e 30 36 3b 74 63 70 3b 31 34 33 33 3b 6e 70 3b \| .06;tcp;1433;np;

    5c 5c 49 4c 53 55 4e 47 31 5c 70 69 70 65 5c 73 \|
    \\\\ILSUNG1\\pipe\\s

    71 6c 5c 71 75 65 72 79 3b 3b \| ql\\query;;

The response conveys the instance information for three instances named
\"YUKONSTD\", \"YUKONDEV\", and \"MSSQLSERVER\".

## CLNT_UCAST_INST

Request:

35. 04 59 55 4b 4f 4e 53 54 44 00 \| .YUKONSTD

The request is for information for an instance named \"YUKONSTD\".

Response:

36. 05 58 00 53 65 72 76 65 72 4e 61 6d 65 3b 49 4c \| .X.ServerName;IL

    53 55 4e 47 31 3b 49 6e 73 74 61 6e 63 65 4e 61 \| SUNG1;InstanceNa

    6d 65 3b 59 55 4b 4f 4e 53 54 44 3b 49 73 43 6c \| me;YUKONSTD;IsCl

    75 73 74 65 72 65 64 3b 4e 6f 3b 56 65 72 73 69 \| ustered;No;Versi

    6f 6e 3b 39 2e 30 30 2e 31 33 39 39 2e 30 36 3b \| on;9.00.1399.06;

    74 63 70 3b 35 37 31 33 37 3b 3b \| tcp;57137;;

## CLNT_UCAST_DAC

Request:

42. 0f 01 59 55 4b 4f 4e 53 54 44 00 \| ..YUKONSTD

The request is for the
[**DAC**](#gt_d50a91b6-9599-4d29-bad9-83fd1f6e6bf6) port of an instance
named \"YUKONSTD\".

Response:

43. 05 06 00 01 32 df \| \....2

The port number is 0xDF32 or 57138.

