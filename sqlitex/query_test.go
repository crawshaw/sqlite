package sqlitex

import (
	"fmt"
	"testing"

	"crawshaw.io/sqlite"
)

func TestResult(t *testing.T) {
	setup := func(t *testing.T) (conn *sqlite.Conn, ins, sel *sqlite.Stmt, done func()) {
		t.Helper()
		conn, err := sqlite.OpenConn(":memory:", 0)
		if err != nil {
			t.Fatal(err)
		}

		if err := Exec(conn, "CREATE TABLE t (c1);", nil); err != nil {
			conn.Close()
			t.Fatal(err)
		}

		ins, err = conn.Prepare("INSERT INTO t (c1) VALUES ($c1);")
		if err != nil {
			conn.Close()
			t.Fatal(err)
		}

		sel, _, err = conn.PrepareTransient("SELECT c1 FROM t")
		if err != nil {
			ins.Finalize()
			conn.Close()
			t.Fatal(err)
		}

		return conn, ins, sel, func() {
			ins.Finalize()
			sel.Finalize()
			conn.Close()
		}
	}

	t.Run("Int64/Result", func(t *testing.T) {
		_, ins, sel, done := setup(t)
		defer done()

		input := int64(1234)
		ins.SetInt64("$c1", input)
		if _, err := ins.Step(); err != nil {
			t.Fatal(err)
		}
		result, err := ResultInt64(sel)
		if err != nil {
			t.Fatal(err)
		}
		if result != input {
			t.Fatal()
		}
	})

	t.Run("Int64/AsText", func(t *testing.T) {
		_, ins, sel, done := setup(t)
		defer done()

		input := int64(1234)
		ins.SetInt64("$c1", input)
		if _, err := ins.Step(); err != nil {
			t.Fatal(err)
		}
		result, err := ResultText(sel)
		if err != nil {
			t.Fatal(err)
		}
		if result != fmt.Sprint(input) {
			t.Fatal(result, "!=", input)
		}
	})

	t.Run("Int64/ErrMultipleResults", func(t *testing.T) {
		_, ins, sel, done := setup(t)
		defer done()

		ins.SetInt64("$c1", int64(1234))
		if _, err := ins.Step(); err != nil {
			t.Fatal(err)
		}
		if err := ins.Reset(); err != nil {
			t.Fatal(err)
		}
		ins.SetInt64("$c1", int64(5678))
		if _, err := ins.Step(); err != nil {
			t.Fatal(err)
		}

		result, err := ResultInt64(sel)
		if result != 0 {
			t.Fatal()
		}

		// XXX: Using direct equality check as go.mod specifies Go 1.12, which does not
		// have errors.Is. When go.mod is updated past 1.12, this should use errors.Is.
		if err != ErrMultipleResults {
			t.Fatal("err != ErrMultipleResults", err)
		}
	})

	t.Run("Int64/ErrNoResult", func(t *testing.T) {
		_, _, sel, done := setup(t)
		defer done()

		result, err := ResultInt64(sel)
		if result != 0 {
			t.Fatal()
		}
		if err != ErrNoResults {
			t.Fatal("err != ErrNoResults", err)
		}
	})

	t.Run("Float/Result", func(t *testing.T) {
		_, ins, sel, done := setup(t)
		defer done()

		input := float64(1234)
		ins.SetFloat("$c1", input)
		if _, err := ins.Step(); err != nil {
			t.Fatal(err)
		}
		result, err := ResultFloat(sel)
		if err != nil {
			t.Fatal(err)
		}
		if result != input {
			t.Fatal()
		}
	})

	t.Run("Float/ErrMultipleResults", func(t *testing.T) {
		_, ins, sel, done := setup(t)
		defer done()

		ins.SetFloat("$c1", 123.4)
		if _, err := ins.Step(); err != nil {
			t.Fatal(err)
		}
		if err := ins.Reset(); err != nil {
			t.Fatal(err)
		}
		ins.SetFloat("$c1", 567.8)
		if _, err := ins.Step(); err != nil {
			t.Fatal(err)
		}
		result, err := ResultFloat(sel)
		if result != 0 {
			t.Fatal()
		}

		// XXX: Using direct equality check as go.mod specifies Go 1.12, which does not
		// have errors.Is. When go.mod is updated past 1.12, this should use errors.Is.
		if err != ErrMultipleResults {
			t.Fatal("err != ErrMultipleResults", err)
		}
	})

	t.Run("Float/ErrNoResult", func(t *testing.T) {
		_, _, sel, done := setup(t)
		defer done()

		result, err := ResultFloat(sel)
		if result != 0 {
			t.Fatal()
		}
		if err != ErrNoResults {
			t.Fatal("err != ErrNoResults", err)
		}
	})

	t.Run("Text/Result", func(t *testing.T) {
		_, ins, sel, done := setup(t)
		defer done()

		input := "test"
		ins.SetText("$c1", input)
		if _, err := ins.Step(); err != nil {
			t.Fatal(err)
		}
		result, err := ResultText(sel)
		if err != nil {
			t.Fatal(err)
		}
		if result != input {
			t.Fatal()
		}
	})

	t.Run("Text/ErrMultipleResults", func(t *testing.T) {
		_, ins, sel, done := setup(t)
		defer done()

		ins.SetText("$c1", "test1")
		if _, err := ins.Step(); err != nil {
			t.Fatal(err)
		}
		if err := ins.Reset(); err != nil {
			t.Fatal(err)
		}
		ins.SetText("$c1", "test2")
		if _, err := ins.Step(); err != nil {
			t.Fatal(err)
		}
		result, err := ResultText(sel)
		if result != "" {
			t.Fatal()
		}

		// XXX: Using direct equality check as go.mod specifies Go 1.12, which does not
		// have errors.Is. When go.mod is updated past 1.12, this should use errors.Is.
		if err != ErrMultipleResults {
			t.Fatal("err != ErrMultipleResults", err)
		}
	})

	t.Run("Text/ErrNoResult", func(t *testing.T) {
		_, _, sel, done := setup(t)
		defer done()

		result, err := ResultText(sel)
		if result != "" {
			t.Fatal()
		}
		if err != ErrNoResults {
			t.Fatal("err != ErrNoResults", err)
		}
	})

	t.Run("Text/AsInt64", func(t *testing.T) {
		_, ins, sel, done := setup(t)
		defer done()

		input := "test"
		ins.SetText("$c1", input)
		if _, err := ins.Step(); err != nil {
			t.Fatal(err)
		}
		result, err := ResultInt64(sel)
		if err != nil {
			t.Fatal(err)
		}
		if result != 0 {
			t.Fatal(result, "!=", 0)
		}
	})
}
