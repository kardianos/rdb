package example

import (
	"bitbucket.org/kardianos/rdb"
	_ "bitbucket.org/kardianos/tds"
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

	// Truncate as the round trip is slightly lossy.
	truncTo := 200 * time.Millisecond
	tm := time.Now().Truncate(truncTo)

	cmd := &rdb.Command{
		Sql: `
			select
				-- dto = cast('2014-01-31' as datetime)
				-- dtoS = cast(@dto as nvarchar(max)),
				dto = @dto
		`,
		Arity: rdb.OneOnly,
		Input: []rdb.Param{
			rdb.Param{N: "dto", T: rdb.TypeOnlyDateTime, V: tm},
		},
	}

	db := rdb.OpenMust(config)
	defer db.Close()

	var dto time.Time

	res := db.Query(cmd)
	defer res.Close()

	res.PrepAll(&dto)

	res.Scan()

	dto = dto.Truncate(truncTo)

	compare := []interface{}{dto}

	for i := range compare {
		if i >= len(cmd.Input) {
			return
		}
		in := cmd.Input[i]
		if !reflect.DeepEqual(compare[i], in.V) {
			t.Errorf("Param %s did not round trip: Want (%v) got (%v)", in.N, in.V, compare[i])
		}
	}
}
