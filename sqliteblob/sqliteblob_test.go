package sqliteblob_test

import (
	"strings"
	"testing"

	"crawshaw.io/sqlite"
	"crawshaw.io/sqlite/sqliteblob"
	"crawshaw.io/sqlite/sqlitex"
)

func TestInsert(t *testing.T) {
	conn, err := sqlite.OpenConn(":memory:", 0)
	if err != nil {
		t.Fatal(err)
	}
	err = sqlitex.ExecScript(conn, `
	CREATE TABLE sqliteblob (c);
	ATTACH DATABASE ":memory:" AS db2;
	CREATE TABLE db2.sqliteblob2 (c);`)
	if err != nil {
		t.Fatal(err)
	}

	const s = "Hello, World! I'm a tiny blob."

	id1, err := sqliteblob.Insert(conn, "", "sqliteblob", "c", strings.NewReader(s), int64(len(s)))
	if err != nil {
		t.Fatal(err)
	}
	if id1 == 0 {
		t.Errorf("first insert returned zero rowid")
	}
	id2, err := sqliteblob.Insert(conn, "", "sqliteblob", "c", strings.NewReader(s), int64(len(s)))
	if err != nil {
		t.Fatal(err)
	}
	if id2 == 0 {
		t.Errorf("second insert returned zero rowid")
	}
	if id1 == id2 {
		t.Errorf("multiple inserts returning matching rowid: %d", id1)
	}

	count, err := sqlitex.ResultInt(conn.Prep(`SELECT count(*) FROM sqliteblob`))
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Errorf("want 2 rows, got %d", count)
	}
	_, err = sqliteblob.Insert(conn, "", "sqliteblob", "c", strings.NewReader(s), int64(len(s)-1))
	if err == nil {
		t.Error("expected error when reader is too long")
	}
	_, err = sqliteblob.Insert(conn, "", "sqliteblob", "c", strings.NewReader(s), int64(len(s)+1))
	if err == nil {
		t.Error("expected error when reader is too short")
	}

	count, err = sqlitex.ResultInt(conn.Prep(`SELECT count(*) FROM sqliteblob`))
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Errorf("want 2 rows, got %d", count)
	}

	db2id1, err := sqliteblob.Insert(conn, "db2", "sqliteblob2", "c", strings.NewReader(s), int64(len(s)))
	if err != nil {
		t.Fatal(err)
	}
	if db2id1 == 0 {
		t.Errorf("third insert (second db) returned zero rowid")
	}

	count, err = sqlitex.ResultInt(conn.Prep(`SELECT count(*) FROM sqliteblob`))
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Errorf("want 2 rows, got %d", count)
	}
}
