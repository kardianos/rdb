package example

import (
	"bitbucket.org/kardianos/rdb"
	"bitbucket.org/kardianos/tds"
	"reflect"
	"testing"
	"time"
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
	config := rdb.ParseConfigMust("tds://TESTU@localhost/SqlExpress?db=master")

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
		Arity: rdb.OneOnly,
		Input: []rdb.Param{
			rdb.Param{N: "dt", T: tds.TypeOldOnlyDateTime, V: dt},
			rdb.Param{N: "d", T: rdb.TypeOnlyDate, V: d},
			rdb.Param{N: "t", T: rdb.TypeOnlyTime, V: tm},
			rdb.Param{N: "dt2", T: rdb.TypeOnlyDateTime, V: dt2},
			rdb.Param{N: "dto", T: rdb.TypeTime, V: dto},
			rdb.Param{N: "dtS", T: tds.TypeOldOnlyDateTime, V: dtS},
			rdb.Param{N: "dt2S", T: rdb.TypeOnlyDateTime, V: dtS},
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

	dto = res.Get("dto").V.(time.Time)

	dt = dt.Round(truncTo)

	t.Logf("D: %v", d)
	t.Logf("DT2: %v", dt2)
	t.Logf("DTO: %v", dto)

	t.Logf("DTV: %v", res.Get("dtV").V.(time.Time))
	t.Logf("DTS: %s", res.Get("dtS").V.([]byte))
	t.Logf("DT2V: %v", res.Get("dt2V").V.(time.Time))
	t.Logf("DT2S: %s", res.Get("dt2S").V.([]byte))
	t.Logf("dtoS: %s", res.Get("dtoS").V.([]byte))

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