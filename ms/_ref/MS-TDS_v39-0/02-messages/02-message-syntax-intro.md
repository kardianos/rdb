## Message Syntax

Character data, such as [**SQL
statements**](#gt_dc5ca224-43ec-4b44-9dba-726d6fd6057d), within a TDS
message is in [**Unicode**](#gt_c305d0ab-8b94-461a-bd76-13b40cb8c4d8),
unless the character data represents the data value of an ASCII data
type, such as a non-Unicode data column. A character count within TDS is
a count of characters, rather than of bytes, except when that character
count is explicitly specified as a byte count.

