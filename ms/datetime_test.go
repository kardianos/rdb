// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package ms

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/kardianos/rdb"
)

func TestDateTimeRoundTrip(t *testing.T) {
	checkSkip(t)

	if parallel {
		t.Parallel()
	}
	defer assertFreeConns(t)
	defer recoverTest(t)

	// Truncate as the round trip for DateTimeN is slightly lossy.
	truncTo := 200 * time.Millisecond

	dt := time.Now().Round(truncTo)
	d := time.Now().Truncate(time.Hour * 24).UTC()
	tm := time.Now().Sub(time.Now().Truncate(time.Hour * 24))

	locName := "America/Los_Angeles"
	loc, err := time.LoadLocation(locName)
	if err != nil {
		t.Fatalf("Could not load location: %s, %v", locName, err)
	}
	dto1 := time.Date(2000, 1, 1, 23, 45, 01, 0, loc)
	dto2 := time.Date(2000, 1, 1, 1, 45, 01, 0, loc)
	dto3 := time.Date(2000, 1, 1, 23, 45, 01, 0, time.UTC)
	dto4 := time.Date(2000, 1, 1, 1, 45, 01, 0, time.UTC)

	cmd := &rdb.Command{
		SQL: `
select
	dt_old = @dt_old,
	d = @d,
	t = @t,

	d1_str = convert(nvarchar(100), @d1_str),
	d2_str = convert(nvarchar(100), @d2_str),
	d1u_str = convert(nvarchar(100), @d1u_str),
	d2u_str = convert(nvarchar(100), @d2u_str),

	dt1_v1_str = convert(nvarchar(100), @dt1_v1_str),
	dt2_v1_str = convert(nvarchar(100), @dt2_v1_str),
	dt1 = @dt1,
	dt2 = @dt2,
	t1_str = format(@t1_str, 'hh\:mm', 'en-us'),
	t2_str = format(@t2_str, 'hh\:mm', 'en-us'),

	dto1 = @dto1,
	dto2 = @dto2,
	dtx1 = @dtx1,
	dtx2 = @dtx2,
	dto1_str = convert(nvarchar(100), @dto1_str),
	dto2_str = convert(nvarchar(100), @dto2_str),
	dt1_str = convert(nvarchar(100), @dt1_str),
	dt2_str = convert(nvarchar(100), @dt2_str),

	dto3 = @dto3,
	dto4 = @dto4,
	dto3_str = convert(nvarchar(100), @dto3_str),
	dto4_str = convert(nvarchar(100), @dto4_str),
	dt3_str = convert(nvarchar(100), @dt3_str),
	dt4_str = convert(nvarchar(100), @dt4_str),

	dto1u = @dto1u,
	dto2u = @dto2u,
	dt1u = @dt1u,
	dt2u = @dt2u,
	dto1u_str = convert(nvarchar(100), @dto1u_str),
	dto2u_str = convert(nvarchar(100), @dto2u_str),
	dt1u_str = convert(nvarchar(100), @dt1u_str),
	dt2u_str = convert(nvarchar(100), @dt2u_str)
		`,
		Arity: rdb.OneMust,
	}

	inZone := func(x time.Time, loc *time.Location) time.Time {
		return time.Date(x.Year(), x.Month(), x.Day(), x.Hour(), x.Minute(), x.Second(), x.Nanosecond(), loc)
	}

	list := []struct {
		name string
		t    rdb.Type
		in   interface{}
		want interface{}
		got  interface{}
		proc func(interface{}) interface{}
	}{
		{name: "dt_old", t: TypeOldTD, in: dt, want: dt, proc: func(v interface{}) interface{} {
			x := v.(time.Time).Round(truncTo)
			return inZone(x, time.Local)
		}},
		{name: "d", t: rdb.TypeDate, in: d, want: d},
		{name: "d1_str", t: rdb.TypeDate, in: dto1, want: "2000-01-01"},
		{name: "d2_str", t: rdb.TypeDate, in: dto2, want: "2000-01-01"},
		{name: "d1u_str", t: rdb.TypeDate, in: dto1.UTC(), want: "2000-01-02"},
		{name: "d2u_str", t: rdb.TypeDate, in: dto2.UTC(), want: "2000-01-01"},
		{name: "t", t: rdb.TypeTime, in: tm, want: tm},
		{name: "t1_str", t: rdb.TypeTime, in: dto1, want: "23:45"},
		{name: "t2_str", t: rdb.TypeTime, in: dto2, want: "01:45"},
		{name: "dt1_v1_str", t: TypeOldTD, in: dto1, want: "Jan  1 2000 11:45PM"},
		{name: "dt2_v1_str", t: TypeOldTD, in: dto2, want: "Jan  1 2000  1:45AM"},
		{name: "dt1", t: rdb.TypeTimestamp, in: dto1, want: dto1, proc: func(i interface{}) interface{} {
			return inZone(i.(time.Time), loc)
		}},
		{name: "dt2", t: rdb.TypeTimestamp, in: dto2, want: dto2, proc: func(i interface{}) interface{} {
			return inZone(i.(time.Time), loc)
		}},

		{name: "dto1", t: rdb.TypeTimestampz, in: dto1, want: dto1},
		{name: "dto2", t: rdb.TypeTimestampz, in: dto2, want: dto2},
		{name: "dtx1", t: rdb.TypeTimestampz, in: dto1, want: "2000-01-01 23:45:01 -0800 UTC -8:00"},
		{name: "dtx2", t: rdb.TypeTimestampz, in: dto2, want: "2000-01-01 01:45:01 -0800 UTC -8:00"},
		{name: "dto1_str", t: rdb.TypeTimestampz, in: dto1, want: "2000-01-01 23:45:01.0000000 -08:00"},
		{name: "dto2_str", t: rdb.TypeTimestampz, in: dto2, want: "2000-01-01 01:45:01.0000000 -08:00"},
		{name: "dt1_str", t: rdb.TypeTimestamp, in: dto1, want: "2000-01-01 23:45:01.0000000"},
		{name: "dt2_str", t: rdb.TypeTimestamp, in: dto2, want: "2000-01-01 01:45:01.0000000"},

		{name: "dto3", t: rdb.TypeTimestampz, in: dto3, want: dto3},
		{name: "dto4", t: rdb.TypeTimestampz, in: dto4, want: dto4},
		{name: "dto3_str", t: rdb.TypeTimestampz, in: dto3, want: "2000-01-01 23:45:01.0000000 +00:00"},
		{name: "dto4_str", t: rdb.TypeTimestampz, in: dto4, want: "2000-01-01 01:45:01.0000000 +00:00"},
		{name: "dt3_str", t: rdb.TypeTimestamp, in: dto3, want: "2000-01-01 23:45:01.0000000"},
		{name: "dt4_str", t: rdb.TypeTimestamp, in: dto4, want: "2000-01-01 01:45:01.0000000"},

		{name: "dto1u", t: rdb.TypeTimestampz, in: dto1.UTC(), want: "2000-01-02 07:45:01 +0000 UTC"},
		{name: "dto2u", t: rdb.TypeTimestampz, in: dto2.UTC(), want: "2000-01-01 09:45:01 +0000 UTC"},
		{name: "dt1u", t: rdb.TypeTimestamp, in: dto1.UTC(), want: "2000-01-02 07:45:01 +0000 UTC"},
		{name: "dt2u", t: rdb.TypeTimestamp, in: dto2.UTC(), want: "2000-01-01 09:45:01 +0000 UTC"},

		{name: "dto1u_str", t: rdb.TypeTimestampz, in: dto1.UTC(), want: "2000-01-02 07:45:01.0000000 +00:00"},
		{name: "dto2u_str", t: rdb.TypeTimestampz, in: dto2.UTC(), want: "2000-01-01 09:45:01.0000000 +00:00"},
		{name: "dt1u_str", t: rdb.TypeTimestamp, in: dto1.UTC(), want: "2000-01-02 07:45:01.0000000"},
		{name: "dt2u_str", t: rdb.TypeTimestamp, in: dto2.UTC(), want: "2000-01-01 09:45:01.0000000"},
	}

	params := make([]rdb.Param, 0, len(list))

	for _, item := range list {
		params = append(params, rdb.Param{
			Name:  item.name,
			Type:  item.t,
			Value: item.in,
		})
	}

	ctx := context.Background()
	res := db.Query(ctx, cmd, params...)
	defer res.Close()

	if res.Next() == false {
		t.Fatal("expected row")
	}

	res.Scan()

	for i, item := range list {
		v := res.Get(item.name)
		if b, ok := v.([]byte); ok {
			v = string(b)
		}
		if item.proc != nil {
			v = item.proc(v)
		}
		list[i].got = v
	}

	for _, item := range list {
		t.Run(item.name, func(t *testing.T) {
			diff := false
			switch w := item.want.(type) {
			default:
				if !reflect.DeepEqual(item.got, w) {
					diff = true
				}
			case string:
				switch g := item.got.(type) {
				default:
					t.Fatal("unsupported got type")
				case string:
					diff = (g != w)
				case time.Duration:
					diff = (g.String() != w)
				case time.Time:
					diff = (g.String() != w)
				}
			case time.Duration:
				z := item.got.(time.Duration)
				diff = (z / 1_000) != (w / 1_000)
			case time.Time:
				if !w.Equal(item.got.(time.Time)) {
					diff = true
				}
			}
			if diff {
				t.Errorf("Param %s did not round trip: Want (%v) got (%v)", item.name, item.want, item.got)
			}
		})
	}
}

func TestDateTimePull(t *testing.T) {
	checkSkip(t)
	if parallel {
		t.Parallel()
	}
	defer assertFreeConns(t)
	defer recoverTest(t)

	dStaticCheck := time.Date(2015, 11, 18, 0, 0, 0, 0, time.UTC)

	var dStaticOut time.Time

	cmd := &rdb.Command{
		SQL: `
if object_id('tempdb..#timeTemp') is not null begin
	drop table #timeTemp
end

create table #timeTemp (
	D date
)
insert into #timeTemp (D)
values ('2015-11-18');

select
	dStatic = D
from
	#timeTemp
;
		`,
		Arity: rdb.OneMust,
	}
	res := db.Query(context.Background(), cmd)
	defer res.Close()

	res.Prep("dStatic", &dStaticOut)

	res.Scan()

	t.Logf("D Static Check: %v", dStaticCheck)
	t.Logf("D Static Out: %v", dStaticOut)

	if dStaticOut.Equal(dStaticCheck) == false {
		t.Errorf("dStatc not equal")
	}
}

func TestDateTZ(t *testing.T) {
	checkSkip(t)
	defer assertFreeConns(t)
	defer recoverTest(t)

	datecheck := time.Date(2017, 1, 9, 20, 30, 0, 0, time.FixedZone("Pacific", -8*60*60))

	const wantDate = "01/09/2017"

	cmd := &rdb.Command{
		SQL:   `select DS = convert(nvarchar(100), @d, 101);`,
		Arity: rdb.OneMust,
	}
	res := db.Query(context.Background(), cmd, rdb.Param{Name: "d", Type: rdb.TypeDate, Value: datecheck})
	defer res.Close()

	var dOut string
	res.Prep("DS", &dOut)

	res.Scan()

	if dOut != wantDate {
		t.Fatalf("wanted %q, got %q", wantDate, dOut)
	}
}
