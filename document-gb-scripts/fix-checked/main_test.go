package main

import (
	"database/sql"
	"fmt"
	tnp "github.com/gb-inc/tylerdb-nightly-processes"
	"testing"
	"time"
)

const wantI = 3

func TestFixCheckedDate(t *testing.T) {
	db, err := tnp.NewDB(".", "1433", "TaxDB_Dev")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	defer db.Close()

	getRandomRows := func(tx *sql.Tx) ([]int64, error) {
		selids := `SELECT TOP 3 PropertyID FROM Property WHERE ParcelEliminated IS NULL
ORDER BY NEWID()`

		rows, err := db.Query(selids)
		if err != nil {
			return nil, fmt.Errorf("error getting propertyIDs for test: %v", err)
		}
		ids := []int64{}
		var i int64
		for rows.Next() {
			if err := rows.Scan(&i); err != nil {
				return nil, fmt.Errorf("error scanning selguid result: %v", err)
			}
			ids = append(ids, i)
		}

		_ = rows.Close()
		return ids, nil
	}

	updateChecked := func(tx *sql.Tx, id, check int64) error {
		updchk := `UPDATE Property SET Checked=@check, CheckDate='1/1/1989' WHERE PropertyID=@PropertyID`
		if _, err := tx.Exec(updchk, sql.Named("check", check), sql.Named("PropertyID", id)); err != nil {
			return err
		}
		return nil
	}
	insertNote := func(tx *sql.Tx, id int64, createdOn time.Time) error {
		insNote := `INSERT INTO PropertyNotes (PropertyID, NoteType, CreatedOn, CreatedBy, NoteText)
VALUES (@PropertyID, 'Checked', @CreatedOn, 'tt1', 'test note text')`
		if _, err := tx.Exec(insNote, sql.Named("PropertyID", id), sql.Named("CreatedOn", createdOn)); err != nil {
			return err
		}
		return nil
	}

	type result struct {
		check int64
		by string
		dte time.Time
	}
	selResult := func(tx *sql.Tx, id int64) (*result, error) {
		selRes := `SELECT Checked, CheckedBy, CheckDate FROM Property WHERE PropertyID=@PropertyID`
		rslt := result{}
		if err = tx.QueryRow(selRes, sql.Named("PropertyID", id)).Scan(&rslt.check, &rslt.by, &rslt.dte); err != nil {
			return nil, err
		}
		return &rslt, nil
	}

	tx, err := db.Begin()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	defer tx.Rollback()

//setup
	//get 3 random properties
	props, err := getRandomRows(tx)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	//update their check dates to 1/1/1989
	//update one to checked to 0 another to 2 and another to 4
	ts := time.Now()
	checkfs := []int64{0,2,4}
	for i, id := range props {
		if err = updateChecked(tx, id, checkfs[i]); err != nil {
			t.Error(err)
			t.FailNow()
		}
		//add note record with 'Checked' type and created on now with createdby tt1
		if err = insertNote(tx, id, ts); err != nil {
			t.Error(err)
			t.FailNow()
		}
	}

	i, err := tnp.QueryReturningRowCount(tx, fixCheckedSql)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if i != wantI {
		t.Errorf("unexpected row count - got: %d, want: %d", i, wantI)
		t.FailNow()
	}

	var r *result
	checkfsr := []int64{-1, 3, 4}
	for i, id := range props {
		r, err = selResult(tx, id)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}
		if r.check != checkfsr[i] {
			t.Errorf("test row [%d]: bad checked flag set for original flag %d - want: %d got %d", i, checkfs[i], checkfsr[i], r.check)
		}
		if r.by != "tt1" {
			t.Errorf("test row [%d]: bad CheckedBy: got %s want: tt1", i, r.by)
		}
		wantdf := ts.Format("01/02/2006 15:04")
		gotdf := ts.Format("01/02/2006 15:04")
		if wantdf != gotdf {
			t.Errorf("test row [%d]: bad CheckedOn: got %s, want: %s", i, gotdf, wantdf)
		}
	}
}