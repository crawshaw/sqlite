package main

import (
	"fmt"
	"log"

	"crawshaw.io/sqlite"
)

func main() {
	dbString := "file:./database.db?key=swordfish&journal_mode=wal"

	conn, err := sqlite.OpenConn(dbString, 0)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	stmt, err := conn.Prepare(`CREATE TABLE IF NOT EXISTS groups ( id TEXT, name TEXT );`)
	if err != nil {
		log.Println(err)
		return
	}
	stmt.Step()

	stmt, err = conn.Prepare(`INSERT INTO groups (id, name) VALUES ('1', 'name 1');`)
	if err != nil {
		log.Println(err)
		return
	}
	stmt.Step()

	stmt, err = conn.Prepare(`SELECT * FROM groups;`)
	if err != nil {
		log.Println(err)
		return
	}

	for {
		rowReturned, err := stmt.Step()
		if err != nil {
			log.Println(err)
			return
		}

		if !rowReturned {
			break
		}

		id := stmt.GetText("id")
		name := stmt.GetText("name")

		fmt.Println(id, name)
	}

}
