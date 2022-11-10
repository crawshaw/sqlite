// Copyright (c) 2018 David Crawshaw <david@zentus.com>
//
// Permission to use, copy, modify, and distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
// WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
// MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
// ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
// WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
// ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
// OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.

package sqlite_test

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"crawshaw.io/sqlite"
	"crawshaw.io/sqlite/sqlitex"
)

func TestConn(t *testing.T) {
	c, err := sqlite.OpenConn(":memory:", 0)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := c.Close(); err != nil {
			t.Error(err)
		}
	}()

	mustCreateBarTable(t, c)

	fooVals := []string{
		"bar",
		"baz",
		"bop",
	}

	for i, val := range fooVals {
		stmt, err := c.Prepare("INSERT INTO bartable (foo1, foo2) VALUES ($f1, $f2);")
		if err != nil {
			t.Fatal(err)
		}
		stmt.SetText("$f1", val)
		stmt.SetInt64("$f2", int64(i))
		hasRow, err := stmt.Step()
		if err != nil {
			t.Errorf("INSERT %q: %v", val, err)
		}
		if hasRow {
			t.Errorf("INSERT %q: has row", val)
		}
	}

	stmt, err := c.Prepare("SELECT foo1, foo2 FROM bartable;")
	if err != nil {
		t.Fatal(err)
	}
	gotVals := []string{}
	gotInts := []int64{}
	for {
		hasRow, err := stmt.Step()
		if err != nil {
			t.Errorf("SELECT: %v", err)
		}
		if !hasRow {
			break
		}
		val := stmt.ColumnText(0)
		if getVal := stmt.GetText("foo1"); getVal != val {
			t.Errorf(`GetText("foo1")=%q, want %q`, getVal, val)
		}
		intVal := stmt.ColumnInt64(1)
		if getIntVal := stmt.GetInt64("foo2"); getIntVal != intVal {
			t.Errorf(`GetText("foo2")=%q, want %q`, getIntVal, intVal)
		}
		typ := stmt.ColumnType(0)
		if typ != sqlite.SQLITE_TEXT {
			t.Errorf(`ColumnType(0)=%q, want %q`, typ, sqlite.SQLITE_TEXT)
		}
		if getType := stmt.GetType("foo1"); getType != typ {
			t.Errorf(`GetType("foo1")=%q, want %q`, getType, typ)
		}
		intTyp := stmt.ColumnType(1)
		if intTyp != sqlite.SQLITE_INTEGER {
			t.Errorf(`ColumnType(1)=%q, want %q`, intTyp, sqlite.SQLITE_INTEGER)
		}
		if getIntType := stmt.GetType("foo2"); getIntType != intTyp {
			t.Errorf(`GetType("foo2")=%q, want %q`, getIntType, intTyp)
		}
		gotVals = append(gotVals, val)
		gotInts = append(gotInts, intVal)
	}

	if !reflect.DeepEqual(fooVals, gotVals) {
		t.Errorf("SELECT foo1: got %v, want %v", gotVals, fooVals)
	}

	wantInts := []int64{0, 1, 2}
	if !reflect.DeepEqual(wantInts, gotInts) {
		t.Errorf("SELECT foo2: got %v, want %v", gotInts, wantInts)
	}
}

func mustCreateBarTable(tb testing.TB, c *sqlite.Conn) {
	tb.Helper()

	stmt, _, err := c.PrepareTransient("CREATE TABLE bartable (foo1 string, foo2 integer);")
	if err != nil {
		tb.Fatal(err)
	}
	if hasRow, err := stmt.Step(); err != nil {
		tb.Fatal(err)
	} else if hasRow {
		tb.Fatal("CREATE TABLE reports having a row")
	}
	if err := stmt.Finalize(); err != nil {
		tb.Fatal(err)
	}
}

func TestEarlyInterrupt(t *testing.T) {
	c, err := sqlite.OpenConn(":memory:", 0)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := c.Close(); err != nil {
			t.Error(err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	c.SetInterrupt(ctx.Done())

	mustCreateBarTable(t, c)

	cancel()

	_, err = c.Prepare("INSERT INTO bartable (foo1, foo2) VALUES ($f1, $f2);")
	if err == nil {
		t.Fatal("Prepare err=nil, want prepare to fail")
	}
	if code := sqlite.ErrCode(err); code != sqlite.SQLITE_INTERRUPT {
		t.Fatalf("Prepare err=%s, want SQLITE_INTERRUPT", code)
	}
}

func TestStmtInterrupt(t *testing.T) {
	conn, err := sqlite.OpenConn(":memory:", 0)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	stmt := sqlite.InterruptedStmt(conn, "CREATE TABLE intt (c);")

	_, err = stmt.Step()
	if err == nil {
		t.Fatal("interrupted stmt Step succeeded")
	}
	if got := sqlite.ErrCode(err); got != sqlite.SQLITE_INTERRUPT {
		t.Errorf("Step err=%v, want SQLITE_INTERRUPT", got)
	}
}

func TestInterruptStepReset(t *testing.T) {
	c, err := sqlite.OpenConn(":memory:", 0)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := c.Close(); err != nil {
			t.Error(err)
		}
	}()

	err = sqlitex.ExecScript(c, `CREATE TABLE resetint (c);
INSERT INTO resetint (c) VALUES (1);
INSERT INTO resetint (c) VALUES (2);
INSERT INTO resetint (c) VALUES (3);`)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	c.SetInterrupt(ctx.Done())

	stmt := c.Prep("SELECT * FROM resetint;")
	if _, err := stmt.Step(); err != nil {
		t.Fatal(err)
	}
	cancel()
	// next Step needs to reset stmt
	if _, err := stmt.Step(); sqlite.ErrCode(err) != sqlite.SQLITE_INTERRUPT {
		t.Fatalf("want SQLITE_INTERRUPT, got %v", err)
	}
	c.SetInterrupt(nil)
	stmt = c.Prep("SELECT c FROM resetint ORDER BY c;")
	if _, err := stmt.Step(); err != nil {
		t.Fatalf("statement after interrupt reset failed: %v", err)
	}
	stmt.Reset()
}

func TestInterruptReset(t *testing.T) {
	c, err := sqlite.OpenConn(":memory:", 0)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := c.Close(); err != nil {
			t.Error(err)
		}
	}()

	err = sqlitex.ExecScript(c, `CREATE TABLE resetint (c);
INSERT INTO resetint (c) VALUES (1);
INSERT INTO resetint (c) VALUES (2);
INSERT INTO resetint (c) VALUES (3);`)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	c.SetInterrupt(ctx.Done())

	stmt := c.Prep("SELECT * FROM resetint;")
	if _, err := stmt.Step(); err != nil {
		t.Fatal(err)
	}
	cancel()
	c.SetInterrupt(nil)
	stmt = c.Prep("SELECT c FROM resetint ORDER BY c;")
	if _, err := stmt.Step(); err != nil {
		t.Fatalf("statement after interrupt reset failed: %v", err)
	}
	stmt.Reset()
}

func TestTrailingBytes(t *testing.T) {
	conn, err := sqlite.OpenConn(":memory:", 0)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			t.Error(err)
		}
	}()

	stmt, trailingBytes, err := conn.PrepareTransient("BEGIN; -- 56")
	if err != nil {
		t.Error(err)
	}
	stmt.Finalize()
	const want = 6
	if trailingBytes != want {
		t.Errorf("trailingBytes=%d, want %d", trailingBytes, want)
	}
}

func TestTrailingBytesError(t *testing.T) {
	conn, err := sqlite.OpenConn(":memory:", 0)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			t.Error(err)
		}
	}()

	if _, err := conn.Prepare("BEGIN; -- 56"); err == nil {
		t.Error("expecting error on trailing bytes")
	}
}

func TestBadParam(t *testing.T) {
	c, err := sqlite.OpenConn(":memory:", 0)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := c.Close(); err != nil {
			t.Error(err)
		}
	}()

	stmt, err := c.Prepare("CREATE TABLE IF NOT EXISTS badparam (a, b, c);")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := stmt.Step(); err != nil {
		t.Fatal(err)
	}

	stmt, err = c.Prepare("INSERT INTO badparam (a, b, c) VALUES ($a, $b, $c);")
	if err != nil {
		t.Fatal(err)
	}
	stmt.SetText("$a", "col_a")
	stmt.SetText("$b", "col_b")
	stmt.SetText("$badparam", "notaval")
	stmt.SetText("$c", "col_c")
	_, err = stmt.Step()
	if err == nil {
		t.Fatal("expecting error from bad param name, got no error")
	}
	if got := err.Error(); !strings.Contains(got, "$badparam") {
		t.Errorf(`error does not mention "$badparam": %v`, got)
	}
	stmt.Finalize()
}

func TestParallel(t *testing.T) {
	c, err := sqlite.OpenConn(":memory:", 0)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := c.Close(); err != nil {
			t.Error(err)
		}
	}()

	stmt := c.Prep("CREATE TABLE testparallel (c);")
	if _, err := stmt.Step(); err != nil {
		t.Fatal(err)
	}

	stmt = c.Prep("INSERT INTO testparallel (c) VALUES ($c);")
	stmt.SetText("$c", "text")
	if _, err := stmt.Step(); err != nil {
		t.Fatal(err)
	}
	stmt.Reset()
	if _, err := stmt.Step(); err != nil {
		t.Fatal(err)
	}

	stmt = c.Prep("SELECT * from testparallel;")
	if _, err := stmt.Step(); err != nil {
		t.Fatal(err)
	}

	stmt2 := c.Prep("SELECT count(*) from testparallel;")
	if hasRow, err := stmt2.Step(); err != nil {
		t.Fatal(err)
	} else if !hasRow {
		t.Error("expecting count row")
	}

	if hasRow, err := stmt.Step(); err != nil {
		t.Fatal(err)
	} else if !hasRow {
		t.Error("expecting second row")
	}
}

func TestBindBytes(t *testing.T) {
	tests := []struct {
		name     string
		val      []byte
		wantNull bool
	}{
		{
			name: "Data with NUL bytes",
			val:  []byte("\x00\x00hello world\x00\x00\x00"),
		},
		{
			name: "Empty non-nil slice",
			val:  []byte{},
		},
		{
			name:     "Nil slice stores NULL",
			wantNull: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c, err := sqlite.OpenConn(":memory:", 0)
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := c.Close(); err != nil {
					t.Error(err)
				}
			}()

			stmt := c.Prep("CREATE TABLE IF NOT EXISTS bindbytes (c);")
			if _, err := stmt.Step(); err != nil {
				t.Fatal(err)
			}
			stmt = c.Prep("INSERT INTO bindbytes (c) VALUES ($bytes);")
			stmt.SetBytes("$bytes", test.val)
			if _, err := stmt.Step(); err != nil {
				t.Fatal(err)
			}

			if !test.wantNull {
				stmt = c.Prep("SELECT count(*) FROM bindbytes WHERE c = $bytes;")
				stmt.SetBytes("$bytes", test.val)
			} else {
				stmt = c.Prep("SELECT count(*) FROM bindbytes WHERE c IS NULL")
			}
			if hasRow, err := stmt.Step(); err != nil {
				t.Fatal(err)
			} else if !hasRow {
				t.Error("SetBytes: result has no row")
			}
			if got := stmt.ColumnInt(0); got != 1 {
				t.Errorf("SetBytes: count is %d, want 1", got)
			}

			stmt.Reset()

			if test.wantNull {
				// Skip blob test, won't work with a NULL value
				return
			}

			blob, err := c.OpenBlob("", "bindbytes", "c", 1, false)
			if err != nil {
				t.Fatalf("SetBytes: OpenBlob: %v", err)
			}
			defer func() {
				if err := blob.Close(); err != nil {
					t.Error(err)
				}
			}()

			wantEOF := (len(test.val) == 0)
			storedVal := make([]byte, len(test.val)+8)
			n, err := blob.Read(storedVal)
			if err != nil && (!wantEOF || err != io.EOF) {
				t.Fatalf("SetBytes: Read: %v", err)
			}
			if got := (err == io.EOF); got != wantEOF {
				t.Fatalf("SetBytes: Read: got EOF? %t, want %t", got, wantEOF)
			}
			if !bytes.Equal(test.val, storedVal[:n]) {
				t.Fatalf("SetBytes: want: %x, got: %x", test.val, storedVal)
			}
		})
	}
}

func TestBindText(t *testing.T) {
	c, err := sqlite.OpenConn(":memory:", 0)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := c.Close(); err != nil {
			t.Error(err)
		}
	}()

	const val = "column_value"

	stmt := c.Prep("CREATE TABLE IF NOT EXISTS bindtext (c);")
	if _, err := stmt.Step(); err != nil {
		t.Fatal(err)
	}
	stmt = c.Prep("INSERT INTO bindtext (c) VALUES ($text);")
	stmt.SetText("$text", val)
	if _, err := stmt.Step(); err != nil {
		t.Fatal(err)
	}

	stmt = c.Prep("SELECT count(*) FROM bindtext WHERE c = $text;")
	stmt.SetText("$text", val)
	if hasRow, err := stmt.Step(); err != nil {
		t.Fatal(err)
	} else if !hasRow {
		t.Error("SetText: result has no row")
	}
	if got := stmt.ColumnInt(0); got != 1 {
		t.Errorf("SetText: count is %d, want 1", got)
	}

	stmt.Reset()

	stmt.SetText("$text", val)
	if hasRow, err := stmt.Step(); err != nil {
		t.Fatal(err)
	} else if !hasRow {
		t.Error("SetText: result has no row")
	}
	if got := stmt.ColumnInt(0); got != 1 {
		t.Errorf("SetText: count is %d, want 1", got)
	}
}

func TestExtendedCodes(t *testing.T) {
	c, err := sqlite.OpenConn(":memory:", 0)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := c.Close(); err != nil {
			t.Error(err)
		}
	}()

	stmt := c.Prep("CREATE TABLE IF NOT EXISTS extcodes (c UNIQUE);")
	if _, err := stmt.Step(); err != nil {
		t.Fatal(err)
	}
	stmt = c.Prep("INSERT INTO extcodes (c) VALUES ($c);")
	stmt.SetText("$c", "value1")
	if _, err := stmt.Step(); err != nil {
		t.Fatal(err)
	}
	stmt.Reset()
	stmt.SetText("$c", "value1")
	_, err = stmt.Step()
	if err == nil {
		t.Fatal("expected UNIQUE error, got nothing")
	}
	if got, want := sqlite.ErrCode(err), sqlite.SQLITE_CONSTRAINT_UNIQUE; got != want {
		t.Errorf("got err=%s, want %s", got, want)
	}
}

type errWithMessage struct {
	err error
	msg string
}

func (e errWithMessage) Cause() error {
	return e.err
}

func (e errWithMessage) Error() string {
	return e.msg + ": " + e.err.Error()
}

func TestWrappedErrors(t *testing.T) {
	rawErr := sqlite.Error{
		Code:  sqlite.SQLITE_INTERRUPT,
		Loc:   "chao",
		Query: "SELECT * FROM thieves",
	}
	if got, want := sqlite.ErrCode(rawErr), sqlite.SQLITE_INTERRUPT; got != want {
		t.Errorf("got err=%s, want %s", got, want)
	}

	wrappedErr := errWithMessage{err: rawErr, msg: "Doing something"}
	if got, want := sqlite.ErrCode(wrappedErr), sqlite.SQLITE_INTERRUPT; got != want {
		t.Errorf("got err=%s, want %s", got, want)
	}
}

func TestJournalMode(t *testing.T) {
	dir, err := ioutil.TempDir("", "crawshaw.io")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	tests := []struct {
		db    string
		mode  string
		flags sqlite.OpenFlags
	}{
		{
			"test-delete.db",
			"delete",
			sqlite.SQLITE_OPEN_READWRITE | sqlite.SQLITE_OPEN_CREATE,
		},
		{
			"test-wal.db",
			"wal",
			sqlite.SQLITE_OPEN_READWRITE | sqlite.SQLITE_OPEN_CREATE | sqlite.SQLITE_OPEN_WAL,
		},
		{
			"test-default-wal.db",
			"wal",
			0,
		},
		// memory databases can't have wal, only journal_mode=memory
		{
			":memory:",
			"memory",
			0,
		},
		// temp databases can't have wal, only journal_mode=delete
		{
			"",
			"delete",
			0,
		},
	}

	for _, test := range tests {
		if test.db != ":memory:" && test.db != "" {
			test.db = filepath.Join(dir, test.db)
		}
		c, err := sqlite.OpenConn(test.db, test.flags)
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			if err := c.Close(); err != nil {
				t.Error(err)
			}
		}()
		stmt := c.Prep("PRAGMA journal_mode;")
		if hasRow, err := stmt.Step(); err != nil {
			t.Fatal(err)
		} else if !hasRow {
			t.Error("PRAGMA journal_mode: has no row")
		}
		if got := stmt.GetText("journal_mode"); got != test.mode {
			t.Errorf("journal_mode not set properly, got: %s, want: %s", got, test.mode)
		}
	}
}

func TestBusyTimeout(t *testing.T) {
	dir, err := ioutil.TempDir("", "crawshaw.io")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	db := filepath.Join(dir, "busytest.db")

	flags := sqlite.SQLITE_OPEN_READWRITE | sqlite.SQLITE_OPEN_CREATE | sqlite.SQLITE_OPEN_WAL

	conn0, err := sqlite.OpenConn(db, flags)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := conn0.Close(); err != nil {
			t.Error(err)
		}
	}()

	conn1, err := sqlite.OpenConn(db, flags)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := conn1.Close(); err != nil {
			t.Error(err)
		}
	}()

	err = sqlitex.ExecScript(conn0, `
		CREATE TABLE t (c);
		INSERT INTO t (c) VALUES (1);
	`)
	if err != nil {
		t.Fatal(err)
	}

	c0Lock := func() {
		if _, err := conn0.Prep("BEGIN;").Step(); err != nil {
			t.Fatal(err)
		}
		if _, err := conn0.Prep("INSERT INTO t (c) VALUES (2);").Step(); err != nil {
			t.Fatal(err)
		}
	}
	c0Unlock := func() {
		if _, err := conn0.Prep("COMMIT;").Step(); err != nil {
			t.Fatal(err)
		}
	}

	c0Lock()
	done := make(chan struct{})
	go func() {
		_, err = conn1.Prep("INSERT INTO t (c) VALUES (3);").Step()
		if err != nil {
			t.Errorf("insert failed: %v", err)
		}
		close(done)
	}()

	time.Sleep(10 * time.Millisecond)
	select {
	case <-done:
		t.Errorf("done before unlock")
	default:
	}

	c0Unlock()
	<-done

	c0Lock()
	done = make(chan struct{})
	go func() {
		conn1.SetBusyTimeout(5 * time.Millisecond)
		_, err = conn1.Prep("INSERT INTO t (c) VALUES (4);").Step()
		if sqlite.ErrCode(err) != sqlite.SQLITE_BUSY {
			t.Errorf("want SQLITE_BUSY got %v", err)
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Errorf("short busy timeout got stuck")
	}

	c0Unlock()
	<-done
}

func TestColumnIndex(t *testing.T) {
	c, err := sqlite.OpenConn(":memory:", 0)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := c.Close(); err != nil {
			t.Error(err)
		}
	}()

	stmt, err := c.Prepare("CREATE TABLE IF NOT EXISTS columnindex (a, b, c);")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := stmt.Step(); err != nil {
		t.Fatal(err)
	}

	stmt, err = c.Prepare("SELECT b, 1 AS d, a, c FROM columnindex")
	if err != nil {
		t.Fatal(err)
	}

	cols := []struct {
		name string
		idx  int
	}{
		{"a", 2},
		{"b", 0},
		{"c", 3},
		{"d", 1},
		{"badcol", -1},
	}

	for _, col := range cols {
		if got := stmt.ColumnIndex(col.name); got != col.idx {
			t.Errorf("expected column %s to have index %d, got %d", col.name, col.idx, got)
		}
	}

	stmt.Finalize()
}

func TestBindParamName(t *testing.T) {
	c, err := sqlite.OpenConn(":memory:", 0)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := c.Close(); err != nil {
			t.Error(err)
		}
	}()

	stmt, err := c.Prepare("SELECT :foo, :bar;")
	if err != nil {
		t.Fatal(err)
	}
	defer stmt.Finalize()

	var got []string
	for i, n := 1, stmt.BindParamCount(); i <= n; i++ {
		got = append(got, stmt.BindParamName(i))
	}
	// We don't care what indices SQLite picked, so sort returned names.
	sort.Strings(got)
	want := []string{":bar", ":foo"}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("names = %q; want %q", got, want)
	}
}

func TestLimit(t *testing.T) {
	c, err := sqlite.OpenConn(":memory:", 0)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := c.Close(); err != nil {
			t.Error(err)
		}
	}()

	c.Limit(sqlite.SQLITE_LIMIT_SQL_LENGTH, 1)
	stmt, err := c.Prepare("SELECT 1;")
	if err == nil {
		stmt.Finalize()
		t.Fatal("Prepare did not return an error")
	}
	if got, want := sqlite.ErrCode(err), sqlite.SQLITE_TOOBIG; got != want {
		t.Errorf("sqlite.ErrCode(err) = %v; want %v", got, want)
	}
}

func TestReturningClause(t *testing.T) {
	// The RETURNING syntax has been supported by SQLite since version 3.35.0 (2021-03-12).
	// https://www.sqlite.org/lang_returning.html
	c, err := sqlite.OpenConn(":memory:", 0)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := c.Close(); err != nil {
			t.Error(err)
		}
	}()

	stmt := c.Prep(`
		CREATE TABLE fruits (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL
		)
	`)
	if _, err = stmt.Step(); err != nil {
		t.Fatal(err)
	}
	_ = stmt.Finalize()

	stmt = c.Prep(`
		INSERT INTO fruits (name)
		VALUES (:fruit)
		RETURNING id
	`)
	stmt.SetText(":fruit", "bananananana")
	if ok, err := stmt.Step(); err != nil {
		t.Fatal(err)
	} else if !ok {
		t.Fatal("no fruit id returned")
	}
	if id := stmt.GetInt64("id"); id != 1 {
		t.Fatalf("want returned fruit id to be 1, got %d", id)
	}
}

func BenchmarkPrepareTransientAndFinalize(b *testing.B) {
	benchs := []struct {
		name  string
		query string
	}{
		{
			name:  "Select constant",
			query: "SELECT 1",
		},
		{
			name:  "Insert constants",
			query: "INSERT INTO bartable (foo1, foo2) VALUES ('bar', 0);",
		},
		{
			name:  "Insert with params",
			query: "INSERT INTO bartable (foo1, foo2) VALUES ($f1, $f2);",
		},
	}

	for _, bench := range benchs {
		b.Run(bench.name, func(b *testing.B) {
			c, err := sqlite.OpenConn(":memory:", 0)
			if err != nil {
				b.Fatal(err)
			}
			defer func() {
				if err := c.Close(); err != nil {
					b.Error(err)
				}
			}()

			mustCreateBarTable(b, c)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				var err error
				JunkStmt, JunkInt, err = c.PrepareTransient(bench.query)
				if err != nil {
					b.Fatal(err)
				}
				if err := JunkStmt.Finalize(); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkPreparePrimed(b *testing.B) {
	benchs := []struct {
		name  string
		query string
	}{
		{
			name:  "Select constant",
			query: "SELECT 1",
		},
		{
			name:  "Insert constants",
			query: "INSERT INTO bartable (foo1, foo2) VALUES ('bar', 0);",
		},
		{
			name:  "Insert with params",
			query: "INSERT INTO bartable (foo1, foo2) VALUES ($f1, $f2);",
		},
	}

	for _, bench := range benchs {
		b.Run(bench.name, func(b *testing.B) {
			c, err := sqlite.OpenConn(":memory:", 0)
			if err != nil {
				b.Fatal(err)
			}
			defer func() {
				if err := c.Close(); err != nil {
					b.Error(err)
				}
			}()

			mustCreateBarTable(b, c)

			if _, err := c.Prepare(bench.query); err != nil {
				b.Fatalf("Unable to prime Prepare cache with %q: %v", bench.query, err)
			}
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				var err error
				JunkStmt, err = c.Prepare(bench.query)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkPrepareBindAndSelect(b *testing.B) {
	const (
		existingText = "bar"
		existingInt  = 42
	)
	benchs := []struct {
		name     string
		query    string
		bindInt  bool
		i        int64
		bindText bool
		s        string
		doReset  bool
		doClear  bool
	}{
		{
			name:  "Select constant",
			query: "SELECT 18",
		},
		{
			name:    "Select constant, explicit reset",
			query:   "SELECT 18",
			doReset: true,
		},
		{
			name:    "Select int param directly",
			query:   "SELECT $f1",
			bindInt: true,
			i:       existingInt,
		},
		{
			name:    "Select int param directly, explicit reset",
			query:   "SELECT $f1",
			bindInt: true,
			i:       existingInt,
			doReset: true,
		},
		{
			name:     "Select from table",
			query:    "SELECT foo2 FROM bartable WHERE foo1=$f1",
			bindText: true,
			s:        existingText,
		},
		{
			name:     "Select from table, explicit reset and clear",
			query:    "SELECT foo2 FROM bartable WHERE foo1=$f1",
			bindText: true,
			s:        existingText,
			doReset:  true,
			doClear:  true,
		},
	}

	for _, bench := range benchs {
		b.Run(bench.name, func(b *testing.B) {
			c, err := sqlite.OpenConn(":memory:", 0)
			if err != nil {
				b.Fatal(err)
			}
			defer func() {
				if err := c.Close(); err != nil {
					b.Error(err)
				}
			}()

			mustCreateBarTable(b, c)
			mustInsertIntoBarTable(b, c, existingText, existingInt)

			if _, err := c.Prepare(bench.query); err != nil {
				b.Fatalf("Unable to prime Prepare cache with %q: %v", bench.query, err)
			}
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				stmt, err := c.Prepare(bench.query)
				if err != nil {
					b.Fatal(err)
				}

				if bench.bindInt {
					stmt.SetInt64("$f1", bench.i)
				} else if bench.bindText {
					stmt.SetText("$f1", bench.s)
				}

				if hasRow, err := stmt.Step(); err != nil {
					b.Fatal(err)
				} else if !hasRow {
					b.Fatal("No rows")
				}
				JunkInt64 = stmt.ColumnInt64(0)

				if bench.doReset {
					if err := stmt.Reset(); err != nil {
						b.Fatal(err)
					}
				}
				if bench.doClear {
					if err := stmt.ClearBindings(); err != nil {
						b.Fatal(err)
					}
				}
			}
		})
	}
}

func mustInsertIntoBarTable(tb testing.TB, c *sqlite.Conn, s string, i int64) {
	tb.Helper()

	stmt, _, err := c.PrepareTransient("INSERT INTO bartable (foo1, foo2) VALUES ($f1, $f2);")
	if err != nil {
		tb.Fatal(err)
	}
	defer func() {
		if err := stmt.Finalize(); err != nil {
			tb.Fatal(err)
		}
	}()

	stmt.SetText("$f1", s)
	stmt.SetInt64("$f2", i)
	if hasRow, err := stmt.Step(); err != nil {
		tb.Fatalf("INSERT %q, %d: %v", s, i, err)
	} else if hasRow {
		tb.Fatalf("INSERT %q, %d: has row", s, i)
	}
}

// Variables set by benchmarks, to ensure things aren't optimized away
var (
	JunkStmt  *sqlite.Stmt
	JunkInt   int
	JunkInt64 int64
)
