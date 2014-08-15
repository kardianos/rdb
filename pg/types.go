// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the rdb LICENSE file.

package pg

import (
	"bitbucket.org/kardianos/rdb"
)

type Oid int32

type rdbTypeValue struct {
	rdb.Type
	Generic rdb.Type
}

var typeLookup = map[Oid]rdbTypeValue{
	T_unknown: {rdb.TypeUnknown, rdb.TypeUnknown},
	T_bool:    {rdb.TypeBool, rdb.Bool},
	T_bit:     {rdb.TypeBool, rdb.Bool},
	T_bytea:   {rdb.TypeBinary, rdb.Binary},
	T_text:    {rdb.TypeText, rdb.Text},
	T_varchar: {rdb.TypeVarChar, rdb.Text},

	T_char:    {rdb.TypeInt8, rdb.Integer},
	T_int2:    {rdb.TypeInt16, rdb.Integer},
	T_int4:    {rdb.TypeInt32, rdb.Integer},
	T_int8:    {rdb.TypeInt64, rdb.Integer},
	T_float4:  {rdb.TypeFloat32, rdb.Float},
	T_float8:  {rdb.TypeFloat64, rdb.Float},
	T_numeric: {rdb.TypeDecimal, rdb.Decimal},

	T_varbit:      {rdb.TypeBool, rdb.Bool},
	T_time:        {rdb.TypeTime, rdb.Time},
	T_date:        {rdb.TypeDate, rdb.Time},
	T_timestamp:   {rdb.TypeTimestamp, rdb.Time},
	T_timestamptz: {rdb.TypeTimestampz, rdb.Time},

	T_uuid: {rdb.TypeUUID, rdb.Other},
	T_xml:  {rdb.TypeXml, rdb.Other},
	T_json: {rdb.TypeJson, rdb.Other},
}

type typeValue struct {
	Oid Oid
}

var rdbTypeLookup = map[rdb.Type]typeValue{
	rdb.TypeUnknown:     typeValue{Oid: T_unknown},
	rdb.Bool:            typeValue{Oid: T_bool},
	rdb.TypeBool:        typeValue{Oid: T_bool},
	rdb.Binary:          typeValue{Oid: T_bytea},
	rdb.TypeBinary:      typeValue{Oid: T_bytea},
	rdb.TypeText:        typeValue{Oid: T_text},
	rdb.Text:            typeValue{Oid: T_text},
	rdb.TypeVarChar:     typeValue{Oid: T_text},
	rdb.TypeAnsiVarChar: typeValue{Oid: T_text},
	rdb.TypeAnsiText:    typeValue{Oid: T_text},
	rdb.TypeChar:        typeValue{Oid: T_char},
	rdb.TypeInt8:        typeValue{Oid: T_char},
	rdb.TypeInt16:       typeValue{Oid: T_int2},
	rdb.TypeInt32:       typeValue{Oid: T_int4},
	rdb.TypeInt64:       typeValue{Oid: T_int8},
	rdb.Integer:         typeValue{Oid: T_int8},
	rdb.TypeFloat32:     typeValue{Oid: T_float4},
	rdb.TypeFloat64:     typeValue{Oid: T_float8},
	rdb.Float:           typeValue{Oid: T_float8},
	rdb.TypeDecimal:     typeValue{Oid: T_numeric},
	rdb.Decimal:         typeValue{Oid: T_numeric},
	rdb.TypeTime:        typeValue{Oid: T_time},
	rdb.TypeDate:        typeValue{Oid: T_date},
	rdb.TypeTimestamp:   typeValue{Oid: T_timestamp},
	rdb.TypeTimestampz:  typeValue{Oid: T_timestamptz},
	rdb.Time:            typeValue{Oid: T_timestamptz},
	rdb.TypeUUID:        typeValue{Oid: T_uuid},
	rdb.TypeXml:         typeValue{Oid: T_xml},
	rdb.TypeJson:        typeValue{Oid: T_json},
}

const (
	T_bool             Oid = 16
	T_bytea            Oid = 17
	T_char             Oid = 18
	T_name             Oid = 19
	T_int8             Oid = 20
	T_int2             Oid = 21
	T_int2vector       Oid = 22
	T_int4             Oid = 23
	T_regproc          Oid = 24
	T_text             Oid = 25
	T_oid              Oid = 26
	T_tid              Oid = 27
	T_xid              Oid = 28
	T_cid              Oid = 29
	T_oidvector        Oid = 30
	T_pg_type          Oid = 71
	T_pg_attribute     Oid = 75
	T_pg_proc          Oid = 81
	T_pg_class         Oid = 83
	T_json             Oid = 114
	T_xml              Oid = 142
	T__xml             Oid = 143
	T_pg_node_tree     Oid = 194
	T__json            Oid = 199
	T_smgr             Oid = 210
	T_point            Oid = 600
	T_lseg             Oid = 601
	T_path             Oid = 602
	T_box              Oid = 603
	T_polygon          Oid = 604
	T_line             Oid = 628
	T__line            Oid = 629
	T_cidr             Oid = 650
	T__cidr            Oid = 651
	T_float4           Oid = 700
	T_float8           Oid = 701
	T_abstime          Oid = 702
	T_reltime          Oid = 703
	T_tinterval        Oid = 704
	T_unknown          Oid = 705
	T_circle           Oid = 718
	T__circle          Oid = 719
	T_money            Oid = 790
	T__money           Oid = 791
	T_macaddr          Oid = 829
	T_inet             Oid = 869
	T__bool            Oid = 1000
	T__bytea           Oid = 1001
	T__char            Oid = 1002
	T__name            Oid = 1003
	T__int2            Oid = 1005
	T__int2vector      Oid = 1006
	T__int4            Oid = 1007
	T__regproc         Oid = 1008
	T__text            Oid = 1009
	T__tid             Oid = 1010
	T__xid             Oid = 1011
	T__cid             Oid = 1012
	T__oidvector       Oid = 1013
	T__bpchar          Oid = 1014
	T__varchar         Oid = 1015
	T__int8            Oid = 1016
	T__point           Oid = 1017
	T__lseg            Oid = 1018
	T__path            Oid = 1019
	T__box             Oid = 1020
	T__float4          Oid = 1021
	T__float8          Oid = 1022
	T__abstime         Oid = 1023
	T__reltime         Oid = 1024
	T__tinterval       Oid = 1025
	T__polygon         Oid = 1027
	T__oid             Oid = 1028
	T_aclitem          Oid = 1033
	T__aclitem         Oid = 1034
	T__macaddr         Oid = 1040
	T__inet            Oid = 1041
	T_bpchar           Oid = 1042
	T_varchar          Oid = 1043
	T_date             Oid = 1082
	T_time             Oid = 1083
	T_timestamp        Oid = 1114
	T__timestamp       Oid = 1115
	T__date            Oid = 1182
	T__time            Oid = 1183
	T_timestamptz      Oid = 1184
	T__timestamptz     Oid = 1185
	T_interval         Oid = 1186
	T__interval        Oid = 1187
	T__numeric         Oid = 1231
	T_pg_database      Oid = 1248
	T__cstring         Oid = 1263
	T_timetz           Oid = 1266
	T__timetz          Oid = 1270
	T_bit              Oid = 1560
	T__bit             Oid = 1561
	T_varbit           Oid = 1562
	T__varbit          Oid = 1563
	T_numeric          Oid = 1700
	T_refcursor        Oid = 1790
	T__refcursor       Oid = 2201
	T_regprocedure     Oid = 2202
	T_regoper          Oid = 2203
	T_regoperator      Oid = 2204
	T_regclass         Oid = 2205
	T_regtype          Oid = 2206
	T__regprocedure    Oid = 2207
	T__regoper         Oid = 2208
	T__regoperator     Oid = 2209
	T__regclass        Oid = 2210
	T__regtype         Oid = 2211
	T_record           Oid = 2249
	T_cstring          Oid = 2275
	T_any              Oid = 2276
	T_anyarray         Oid = 2277
	T_void             Oid = 2278
	T_trigger          Oid = 2279
	T_language_handler Oid = 2280
	T_internal         Oid = 2281
	T_opaque           Oid = 2282
	T_anyelement       Oid = 2283
	T__record          Oid = 2287
	T_anynonarray      Oid = 2776
	T_pg_authid        Oid = 2842
	T_pg_auth_members  Oid = 2843
	T__txid_snapshot   Oid = 2949
	T_uuid             Oid = 2950
	T__uuid            Oid = 2951
	T_txid_snapshot    Oid = 2970
	T_fdw_handler      Oid = 3115
	T_anyenum          Oid = 3500
	T_tsvector         Oid = 3614
	T_tsquery          Oid = 3615
	T_gtsvector        Oid = 3642
	T__tsvector        Oid = 3643
	T__gtsvector       Oid = 3644
	T__tsquery         Oid = 3645
	T_regconfig        Oid = 3734
	T__regconfig       Oid = 3735
	T_regdictionary    Oid = 3769
	T__regdictionary   Oid = 3770
	T_anyrange         Oid = 3831
	T_event_trigger    Oid = 3838
	T_int4range        Oid = 3904
	T__int4range       Oid = 3905
	T_numrange         Oid = 3906
	T__numrange        Oid = 3907
	T_tsrange          Oid = 3908
	T__tsrange         Oid = 3909
	T_tstzrange        Oid = 3910
	T__tstzrange       Oid = 3911
	T_daterange        Oid = 3912
	T__daterange       Oid = 3913
	T_int8range        Oid = 3926
	T__int8range       Oid = 3927
)
