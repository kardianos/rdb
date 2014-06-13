// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package rdb

type Roller func(t TransactionMust, savepoint string)

// A method to take many panicing members and return a normal error.
// The Roller function will rollback to an existing savepoint if it has not
// already been commited. An empty savepoint parameter to Roller will roll
// the transaction back entirely.
/*
	func ExampleRun() error {
		return rdb.Run(func(r rdb.Roller) error {
			db := rdb.OpenMust(config)
			t := db.Begin()
			r(t, "")

			t.Query(cmd1)

			t.SavePoint("Foo")
			r(t, "Foo")

			t.Query(cmd2)

			t.Commit()
			return nil
		})
	}
*/
func Run(f func(r Roller) error) (err error) {
	trans := make(map[TransactionMust]string)
	defer func() {
		if recovered := recover(); recovered != nil {
			if must, is := recovered.(MustError); is {
				err = must.Err
				return
			}
			panic(recovered)
		}
	}()
	err = f(func(t TransactionMust, savepoint string) {
		trans[t] = savepoint
	})
	var terr error
	for t, savepoint := range trans {
		if !t.Active() {
			continue
		}
		var loopErr error
		nt := t.Normal()
		if len(savepoint) == 0 {
			loopErr = nt.Rollback()
			if terr == nil {
				terr = loopErr
			}
			continue
		}
		loopErr = nt.RollbackTo(savepoint)
		if terr == nil {
			terr = loopErr
		}
		loopErr = nt.Commit()
		if terr == nil {
			terr = loopErr
		}
	}
	if err == nil {
		err = terr
	}
	return
}
