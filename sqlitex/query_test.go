package sqlitex

import (
	"testing"

	"crawshaw.io/sqlite"
)

func TestResult(t *testing.T) {
	conn, err := sqlite.OpenConn(":memory:", 0)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	if err := Exec(conn, "CREATE TABLE t (c1);", nil); err != nil {
		t.Fatal(err)
	}

	insert := func(t *testing.T, value interface{}) {
		t.Helper()
		err := Exec(conn, "INSERT INTO t (c1) VALUES (?);", nil, value)
		if err != nil {
			t.Fatal(err)
		}
	}

	t.Helper()
	sel, err := conn.Prepare("SELECT c1 FROM t;")
	if err != nil {
		t.Fatal(err)
	}

	cleanup := func(tt *testing.T) {
		tt.Helper()
		if sel.DataCount() != 0 {
			tt.Fatal("Stmt was not reset")
		}

		// The following should immediately fail the parent test
		// because we cannot continue without this essential cleanup.
		if err := sel.Reset(); err != nil {
			t.Fatal(err)
		}
		if err := Exec(conn, "DELETE FROM t;", nil); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("Int64/ok", func(t *testing.T) {
		defer cleanup(t)

		input := int64(1234)
		insert(t, input)

		result, err := ResultInt64(sel)
		if err != nil {
			t.Fatal(err)
		}
		if result != input {
			t.Fatal("result != input")
		}
	})

	t.Run("Int64/ErrMultipleResults", func(t *testing.T) {
		defer cleanup(t)

		input := int64(1234)
		insert(t, input)
		insert(t, input)

		result, err := ResultInt64(sel)
		if err != ErrMultipleResults {
			t.Fatal("err != ErrMultipleResults", err)
		}
		if result != input {
			t.Fatal("result != input")
		}
	})

	t.Run("Int64/ErrNoResult", func(t *testing.T) {
		defer cleanup(t)

		_, err := ResultInt64(sel)
		if err != ErrNoResults {
			t.Fatal("err != ErrNoResults", err)
		}
	})

	t.Run("Float/ok", func(t *testing.T) {
		defer cleanup(t)

		input := float64(1234)
		insert(t, input)

		result, err := ResultFloat(sel)
		if err != nil {
			t.Fatal(err)
		}
		if result != input {
			t.Fatal("result != input")
		}
	})

	t.Run("Float/ErrMultipleResults", func(t *testing.T) {
		defer cleanup(t)

		input := float64(1234)
		insert(t, input)
		insert(t, input)

		result, err := ResultFloat(sel)
		if err != ErrMultipleResults {
			t.Fatal("err != ErrMultipleResults", err)
		}
		if result != input {
			t.Fatal()
		}
	})

	t.Run("Float/ErrNoResult", func(t *testing.T) {
		defer cleanup(t)

		_, err := ResultFloat(sel)
		if err != ErrNoResults {
			t.Fatal("err != ErrNoResults", err)
		}
	})

	t.Run("Text/ok", func(t *testing.T) {
		defer cleanup(t)

		input := "test"
		insert(t, input)

		result, err := ResultText(sel)
		if err != nil {
			t.Fatal(err)
		}
		if result != input {
			t.Fatal()
		}
	})

	t.Run("Text/ErrMultipleResults", func(t *testing.T) {
		defer cleanup(t)

		input := "test"
		insert(t, input)
		insert(t, input)

		result, err := ResultText(sel)
		if err != ErrMultipleResults {
			t.Fatal("err != ErrMultipleResults", err)
		}
		if result != input {
			t.Fatal("first result was not returned")
		}
	})

	t.Run("Text/ErrNoResult", func(t *testing.T) {
		defer cleanup(t)

		_, err := ResultText(sel)
		if err != ErrNoResults {
			t.Fatal("err != ErrNoResults", err)
		}
	})
}
