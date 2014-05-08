// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package example

import (
	"reflect"
	"testing"
	"time"

	"bitbucket.org/kardianos/rdb"
	"bitbucket.org/kardianos/rdb/ms"
)

func TestDateTime(t *testing.T) {
	defer func() {
		if re := recover(); re != nil {
			if localError, is := re.(rdb.MustError); is {
				t.Errorf("SQL Error: %v", localError)
				return
			}
			panic(re)
		}
	}()
	config := rdb.ParseConfigMust(testConnectionString)

	// Truncate as the round trip for DateTimeN is slightly lossy.
	truncTo := 200 * time.Millisecond

	dt := time.Now().Round(truncTo)
	d := time.Now().Truncate(time.Hour * 24).UTC()
	tm := time.Now().Sub(time.Now().Truncate(time.Hour * 24))
	dt2 := time.Now()
	dto := time.Now()
	dtS := time.Now()

	cmd := &rdb.Command{
		Sql: `
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
				dto = @dto
		`,
		Arity: rdb.OneMust,
		Input: []rdb.Param{
			{N: "dt", T: ms.TypeOldTD, V: dt},
			{N: "d", T: rdb.TypeDate, V: d},
			{N: "t", T: rdb.TypeTime, V: tm},
			{N: "dt2", T: rdb.TypeTD, V: dt2},
			{N: "dto", T: rdb.TypeTDZ, V: dto},
			{N: "dtS", T: ms.TypeOldTD, V: dtS},
			{N: "dt2S", T: rdb.TypeTD, V: dtS},
		},
	}

	db := rdb.OpenMust(config)
	defer db.Close()

	res := db.Query(cmd)
	defer res.Close()

	res.Prep("dt", &dt)
	res.Prep("d", &d)
	res.Prep("t", &tm)
	res.Prep("dt2", &dt2)

	res.Scan()

	dto = res.Get("dto").(time.Time)

	dt = dt.Round(truncTo)

	t.Logf("D: %v", d)
	t.Logf("DT2: %v", dt2)
	t.Logf("DTO: %v", dto)

	t.Logf("DTV: %v", res.Get("dtV").(time.Time))
	t.Logf("DTS: %s", res.Get("dtS").([]byte))
	t.Logf("DT2V: %v", res.Get("dt2V").(time.Time))
	t.Logf("DT2S: %s", res.Get("dt2S").([]byte))
	t.Logf("dtoS: %s", res.Get("dtoS").([]byte))

	compare := []interface{}{dt, d, tm, dt2, dto}

	for i := range compare {
		if i >= len(cmd.Input) {
			return
		}
		in := cmd.Input[i]
		diff := false
		if tv, ok := in.V.(time.Time); ok {
			if !tv.Equal(compare[i].(time.Time)) {
				diff = true
			}
		} else if !reflect.DeepEqual(compare[i], in.V) {
			diff = true
		}
		if diff {
			t.Errorf("Param %s did not round trip: Want (%v) got (%v)", in.N, in.V, compare[i])
		}
	}
}
