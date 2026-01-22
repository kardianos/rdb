## Token Stream Communication

The following two examples highlight token stream communication. The
packaging of these token streams into packets is not shown in this
section. Actual TDS network data samples are available in section
[4](#Section_e48ebc472e954970a116da45862af90b).

### Sending a SQL Batch

In this example, a [**SQL
statement**](#gt_dc5ca224-43ec-4b44-9dba-726d6fd6057d) is sent to the
server and the results are sent to the client. The SQL statement is as
follows:

4619. SQLStatement = select name, empid from employees

      update employees set salary = salary \* 1.1

      select name from employees where department = \'HR\'

      Client: SQLStatement

      Server: COLMETADATA data stream

      ROW data stream

      .

      .

      ROW data stream

      DONE data stream (with DONE_COUNT & DONE_MORE

      bits set)

      DONE data stream (for UPDATE, with DONE_COUNT &

      DONE_MORE bits set)

      COLMETADATA data stream

      ROW data stream

      .

      .

      ROW data stream

      DONE data stream (with DONE_COUNT bit set)

### Out-of-Band Attention Signal

In this example, a [**SQL
statement**](#gt_dc5ca224-43ec-4b44-9dba-726d6fd6057d) is sent to the
server, yet before all the data has been returned an interrupt or
\"Attention Signal\" is sent to the server. The client reads and
discards any data received between the time the interrupt was sent and
the interrupt acknowledgment was received. The interrupt acknowledgment
from the server is a bit set in the status field of the DONE token.

4640. Client: select name, empid from employees

      Server: COLMETADATA data stream

      ROW data stream

      .

      .

      ROW data stream

      Client: ATTENTION SENT

\[The client reads and discards any data already buffered by the server
until the acknowledgment is found. There might be or might not be a DONE
token with the DONE_MORE bit clear prior to the DONE token with the
DONE_ATTN bit set.\]

4649. Server: DONE data stream (with DONE_ATTN bit set)

