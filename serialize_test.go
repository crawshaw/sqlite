package sqlite_test

import (
	"reflect"
	"strconv"
	"strings"
	"testing"

	"crawshaw.io/sqlite"
	"crawshaw.io/sqlite/sqlitex"
)

func TestSerialize(t *testing.T) {
	conn, err := sqlite.OpenConn(":memory:", 0)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// Create table and insert a few records
	err = sqlitex.Exec(conn, "CREATE TABLE mytable (v1 PRIMARY KEY, v2, v3);", nil)
	if err != nil {
		t.Fatal(err)
	}
	err = sqlitex.Exec(conn,
		"INSERT INTO mytable (v1, v2, v3)  VALUES ('foo', 'bar', 'baz'), ('foo2', 'bar2', 'baz2');", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Serialize
	ser := conn.Serialize("")
	if ser == nil {
		t.Fatal("unexpected nil")
	}
	origLen := len(ser.Bytes())
	t.Logf("Initial serialized size: %v", origLen)

	// Create new connection, confirm table not there
	conn, err = sqlite.OpenConn(":memory:", 0)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	err = sqlitex.Exec(conn, "SELECT * FROM mytable ORDER BY v1;", nil)
	if err == nil || !strings.Contains(err.Error(), "no such table") {
		t.Fatalf("expected no-table error, got: %v", err)
	}

	// Deserialize into connection and allow resizing
	err = conn.Deserialize(ser, sqlite.SQLITE_DESERIALIZE_FREEONCLOSE|sqlite.SQLITE_DESERIALIZE_RESIZEABLE)
	if err != nil {
		t.Fatal(err)
	}

	// Confirm data there
	data := [][3]string{}
	err = sqlitex.Exec(conn, "SELECT * FROM mytable ORDER BY v1;", func(stmt *sqlite.Stmt) error {
		data = append(data, [3]string{stmt.ColumnText(0), stmt.ColumnText(1), stmt.ColumnText(2)})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	expected := [][3]string{{"foo", "bar", "baz"}, {"foo2", "bar2", "baz2"}}
	if !reflect.DeepEqual(expected, data) {
		t.Fatalf("expected %v, got %v", expected, data)
	}

	// Confirm 1000 inserts can be made
	for i := 0; i < 1000; i++ {
		toAppend := strconv.Itoa(i + 3)
		err = sqlitex.Exec(conn, "INSERT INTO mytable (v1, v2, v3)  VALUES ('foo"+
			toAppend+"', 'bar"+toAppend+"', 'baz3"+toAppend+"')", nil)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Serialize again, this time with no-copy
	ser = conn.Serialize("")
	if ser == nil {
		t.Fatal("unexpected nil")
	}
	newLen := len(ser.Bytes())
	if newLen <= origLen {
		t.Fatalf("expected %v > %v", newLen, origLen)
	}
	t.Logf("New serialized size: %v", newLen)

	// Copy the serialized bytes but to not let sqlite own them
	ser = sqlite.NewSerialized(ser.Schema(), ser.Bytes(), false)

	// Create new conn, deserialize read only
	conn, err = sqlite.OpenConn(":memory:", 0)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	err = conn.Deserialize(ser, sqlite.SQLITE_DESERIALIZE_READONLY)
	if err != nil {
		t.Fatal(err)
	}

	// Count
	var total int64
	err = sqlitex.Exec(conn, "SELECT COUNT(1) FROM mytable;", func(stmt *sqlite.Stmt) error {
		total = stmt.ColumnInt64(0)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	} else if total != 1002 {
		t.Fatalf("expected 1002, got %v", total)
	}

	// Try to insert again
	err = sqlitex.Exec(conn, "INSERT INTO mytable (v1, v2, v3)  VALUES ('a', 'b', 'c');", nil)
	if err == nil || !strings.Contains(err.Error(), "readonly") {
		t.Fatalf("expected readonly error, got: %v", err)
	}
}
