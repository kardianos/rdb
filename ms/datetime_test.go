// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package ms

import (
	"reflect"
	"testing"
	"time"

	"github.com/kardianos/rdb"
)

func TestDateTimeRoundTrip(t *testing.T) {
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
		Sql: `
			if object_id('tempdb..##timeTemp') is not null begin
				truncate table ##timeTemp

				insert into ##timeTemp (Name, TM)
				values ('DTO', @dto), ('DTO2', @dto2)
			end
			select
				dtV = cast(@dtS as datetime),
				dtS = cast(@dtS as nvarchar(max)),
				dt2V = cast(@dt2S as datetime2),
				dt2S = cast(@dt2S as nvarchar(max)),
				dtoS = cast(@dto as nvarchar(max)),
				dt = @dt,
				d = @d,
				t = @t,
				t2 = format(@t2, 'hh\:mm', 'en-us'),
				-- dt2 = @dt2,
				dto = @dto,
				dto2 = @dto2
		`,
		Arity: rdb.OneMust,
	}

	list := []struct {
		name string
		t    rdb.Type
		in   interface{}
		want interface{}
		got  interface{}
		proc func(interface{}) interface{}
	}{
		{name: "dt", t: TypeOldTD, in: dt, want: dt, proc: func(v interface{}) interface{} {
			return v.(time.Time).Round(truncTo)
		}},
		{name: "d", t: rdb.TypeDate, in: d, want: d},
		{name: "t", t: rdb.TypeTime, in: tm, want: tm},
		{name: "t2", t: rdb.TypeTime, in: dto2, want: "11:45"},
		{name: "dto", t: rdb.TypeTimestampz, in: dto, want: dto},
		{name: "dto2", t: rdb.TypeTimestampz, in: dto2, want: dto2},
		{name: "dtS", t: TypeOldTD, in: dto, want: "Jan  2 2000  6:45AM"},
		{name: "dt2S", t: rdb.TypeTimestamp, in: dto, want: "2000-01-02 06:45:01.0000000"},
	}

	params := make([]rdb.Param, 0, len(list))

	for _, item := range list {
		params = append(params, rdb.Param{
			Name:  item.name,
			Type:  item.t,
			Value: item.in,
		})
	}

	res := db.Query(cmd, params...)
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
	if parallel {
		t.Parallel()
	}
	defer assertFreeConns(t)
	defer recoverTest(t)

	dStaticCheck := time.Date(2015, 11, 18, 0, 0, 0, 0, time.UTC)

	var dStaticOut time.Time

	cmd := &rdb.Command{
		Sql: `
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
	res := db.Query(cmd)
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
	defer assertFreeConns(t)
	defer recoverTest(t)

	datecheck := time.Date(2017, 1, 9, 20, 30, 0, 0, time.FixedZone("Pacific", -8*60*60))

	const wantDate = "01/09/2017"

	cmd := &rdb.Command{
		Sql:   `select DS = convert(nvarchar(100), @d, 101);`,
		Arity: rdb.OneMust,
	}
	res := db.Query(cmd, rdb.Param{Name: "d", Type: rdb.TypeDate, Value: datecheck})
	defer res.Close()

	var dOut string
	res.Prep("DS", &dOut)

	res.Scan()

	if dOut != wantDate {
		t.Fatalf("wanted %q, got %q", wantDate, dOut)
	}
}
