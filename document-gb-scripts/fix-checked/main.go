package main

import (
	_ "embed"
	"fmt"
	tnp "github.com/gb-inc/tylerdb-nightly-processes"
	"log"
	"time"
)

//go:embed fixChecked.sql
var fixCheckedSql string

func main() {
	lw, err := tnp.NewLogWriter(fmt.Sprintf("./fixChecked_%s.log", time.Now().Format("20060102150405")))
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(lw)

	log.Println("start")
	db, err := tnp.NewDB("db", "14330", "TylerDB")
	if err != nil {
		log.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		_ = db.Close()
		log.Fatal(err)
	}

	var ok bool
	do := func() error {
		defer db.Close()
		defer tnp.HandleTxFunc(tx, &ok)

		i, err := tnp.QueryReturningRowCount(tx, fixCheckedSql)
		if err != nil {
			return err
		}

		log.Printf("successfully fixed checked flags on %d rows\n", i)
		ok = true
		return nil
	}
	if err := do(); err != nil {
		log.Fatal(err)
	}
	log.Println("done")
}
