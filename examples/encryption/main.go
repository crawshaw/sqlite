package main

import (
	"crawshaw.io/sqlite"
	"crawshaw.io/sqlite/sqlitex"
)

func main() {
	dbString := "file:./database.db?key=swordfish&foreign_keys=on"

	poolSize := 10

	dbpool, err := sqlitex.Open(dbString, 0, poolSize)
	if err != nil {
		panic(err)
	}

	// enable all connection to have Foreign key feature
	for i := 0; i < poolSize; i++ {
		conn := dbpool.Get(nil)
		err := sqlitex.Exec(conn, `PRAGMA foreign_keys = ON;`, nil)
		if err != nil {
			panic(err)
		}
		dbpool.Put(conn)
	}

	// // enable all connection to have Foreign key feature
	// for i := 0; i < poolSize; i++ {
	// 	conn := dbpool.Get(nil)
	// 	err := sqlitex.Exec(conn, `PRAGMA key='swordfish';`, nil)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	dbpool.Put(conn)
	// }

	conn := dbpool.Get(nil)
	err = sqlitex.Exec(conn, `PRAGMA key=swordfish;`, nil)
	if err != nil {
		panic(err)
	}

	err = sqlite.Crypto(conn, "swordfish2")
	if err != nil {
		panic(err)
	}

	stmt, err := conn.Prepare(`CREATE TABLE IF NOT EXISTS groups ( id TEXT PRIMARY KEY, name TEXT );`)
	if err != nil {
		panic(err)
	}
	stmt.Step()
	dbpool.Put(conn)

	conn = dbpool.Get(nil)
	stmt, err = conn.Prepare(`INSERT INTO groups (id, name) VALUES ('1', 'name 1');`)
	if err != nil {
		panic(err)
	}
	stmt.Step()
	dbpool.Put(conn)

}
