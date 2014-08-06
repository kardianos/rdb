// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package ms

import (
	"reflect"
	"testing"
	"time"

	"bitbucket.org/kardianos/rdb"
)

func TestDateTime(t *testing.T) {
	defer recoverTest(t)

	// Truncate as the round trip for DateTimeN is slightly lossy.
	truncTo := 200 * time.Millisecond

	dt := time.Now().Round(truncTo)
	d := time.Now().Truncate(time.Hour * 24).UTC()
	tm := time.Now().Sub(time.Now().Truncate(time.Hour * 24))
	dt2 := time.Now()

	locName := "America/Los_Angeles"
	loc, err := time.LoadLocation(locName)
	if err != nil {
		t.Fatalf("Could not load location: %s, %v", locName, err)
	}
	dto := time.Date(2000, 1, 1, 22, 45, 01, 0, loc)
	dto2 := time.Date(2000, 1, 1, 11, 45, 01, 0, loc)
	dtS := time.Now()

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
				dt2 = @dt2,
				dto = @dto,
				dto2 = @dto2
		`,
		Arity: rdb.OneMust,
	}

	params := []rdb.Param{
		{Name: "dt", Type: TypeOldTD, Value: dt},
		{Name: "d", Type: rdb.TypeDate, Value: d},
		{Name: "t", Type: rdb.TypeTime, Value: tm},
		{Name: "dt2", Type: rdb.TypeTimestamp, Value: dt2},
		{Name: "dto", Type: rdb.TypeTimestampz, Value: dto},
		{Name: "dto2", Type: rdb.TypeTimestampz, Value: dto2},
		{Name: "dtS", Type: TypeOldTD, Value: dtS},
		{Name: "dt2S", Type: rdb.TypeTimestamp, Value: dtS},
	}
	res := db.Query(cmd, params...)
	defer res.Close()

	res.Prep("dt", &dt)
	res.Prep("d", &d)
	res.Prep("t", &tm)
	res.Prep("dt2", &dt2)

	res.Scan()

	dto = res.Get("dto").(time.Time)
	dto2 = res.Get("dto2").(time.Time)

	dt = dt.Round(truncTo)

	t.Logf("D: %v", d)
	t.Logf("DT2: %v", dt2)
	t.Logf("DTO: %v", dto)
	t.Logf("DTO2: %v", dto2)

	t.Logf("DTV: %v", res.Get("dtV").(time.Time))
	t.Logf("DTS: %s", res.Get("dtS").([]byte))
	t.Logf("DT2V: %v", res.Get("dt2V").(time.Time))
	t.Logf("DT2S: %s", res.Get("dt2S").([]byte))
	t.Logf("dtoS: %s", res.Get("dtoS").([]byte))

	compare := []interface{}{dt, d, tm, dt2, dto, dto2}

	for i := range compare {
		if i >= len(params) {
			return
		}
		in := params[i]
		diff := false
		if tv, ok := in.Value.(time.Time); ok {
			if !tv.Equal(compare[i].(time.Time)) {
				diff = true
			}
		} else if !reflect.DeepEqual(compare[i], in.Value) {
			diff = true
		}
		if diff {
			t.Errorf("Param %s did not round trip: Want (%v) got (%v)", in.Name, in.Value, compare[i])
		}
	}
}
