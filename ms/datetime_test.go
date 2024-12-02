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
	dto := time.Date(2000, 1, 1, 22, 45, 01, 0, loc)
	dto2 := time.Date(2000, 1, 1, 11, 45, 01, 0, loc)

	cmd := &rdb.Command{
		SQL: `
			if object_id('tempdb..##timeTemp') is not null begin
				truncate table ##timeTemp

				insert into ##timeTemp (Name, TM)
				values ('DTO', @dto), ('DTO2', @dto2)
			end
			select
				dt_old = @dt_old,
				dt = @dt,
				dt2 = @dt2,
				d = @d,
				t = @t,
				t_str = format(@t_str, 'hh\:mm', 'en-us'),
				t2_str = format(@t2_str, 'hh\:mm', 'en-us'),
				dto = @dto,
				dto2 = @dto2,
				dto_str = convert(nvarchar(100), @dto_str),
				dto2_str = convert(nvarchar(100), @dto2_str),
				dt_str = convert(nvarchar(100), @dt_str),
				dt2_str = convert(nvarchar(100), @dt2_str),
				dt_v1_str = convert(nvarchar(100), @dt_v1_str),
				dt2_v1_str = convert(nvarchar(100), @dt2_v1_str)
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
		{name: "t", t: rdb.TypeTime, in: tm, want: tm},
		{name: "t_str", t: rdb.TypeTime, in: dto, want: "22:45"},
		{name: "t2_str", t: rdb.TypeTime, in: dto2, want: "11:45"},
		{name: "dto", t: rdb.TypeTimestampz, in: dto, want: dto},
		{name: "dto2", t: rdb.TypeTimestampz, in: dto2, want: dto2},
		{name: "dto_str", t: rdb.TypeTimestampz, in: dto, want: "2000-01-01 22:45:01.0000000 -08:00"},
		{name: "dto2_str", t: rdb.TypeTimestampz, in: dto2, want: "2000-01-01 11:45:01.0000000 -08:00"},
		{name: "dt_v1_str", t: TypeOldTD, in: dto, want: "Jan  1 2000 10:45PM"},
		{name: "dt2_v1_str", t: TypeOldTD, in: dto2, want: "Jan  1 2000 11:45AM"},
		{name: "dt_str", t: rdb.TypeTimestamp, in: dto, want: "2000-01-01 22:45:01.0000000"},
		{name: "dt2_str", t: rdb.TypeTimestamp, in: dto2, want: "2000-01-01 11:45:01.0000000"},
		{name: "dt", t: rdb.TypeTimestamp, in: dto, want: dto, proc: func(i interface{}) interface{} {
			return inZone(i.(time.Time), loc)
		}},
		{name: "dt2", t: rdb.TypeTimestamp, in: dto2, want: dto2, proc: func(i interface{}) interface{} {
			return inZone(i.(time.Time), loc)
		}},
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
			switch x := item.want.(type) {
			default:
				if !reflect.DeepEqual(item.got, x) {
					diff = true
				}
			case time.Duration:
				z := item.got.(time.Duration)
				diff = (z / 1_000) != (x / 1_000)
			case time.Time:
				if !x.Equal(item.got.(time.Time)) {
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
