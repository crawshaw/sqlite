// Copyright (c) 2020 John Brooks <john.brooks@crimson.no>
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
	"testing"

	"crawshaw.io/sqlite"
	"crawshaw.io/sqlite/sqlitex"
)

func TestCollation(t *testing.T) {
	c, err := sqlite.OpenConn(":memory:", 0)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := c.Close(); err != nil {
			t.Error(err)
		}
	}()

	xCompare := func(a, b string) int {
		if len(a) > len(b) {
			return 1
		} else if len(a) < len(b) {
			return -1
		} else {
			return 0
		}
	}

	if err := c.CreateCollation("sort_strlen", xCompare); err != nil {
		t.Fatal(err)
	}

	err = sqlitex.ExecScript(c, `
		CREATE TABLE strs (str);
		INSERT INTO strs (str) VALUES ('ccc'),('a'),('bb'),('a');
	`)
	if err != nil {
		t.Fatal(err)
	}

	stmt, _, err := c.PrepareTransient("SELECT str FROM strs ORDER BY str COLLATE sort_strlen")
	if err != nil {
		t.Fatal(err)
	}
	wants := []string{"a", "a", "bb", "ccc"}
	for i, want := range wants {
		if _, err := stmt.Step(); err != nil {
			t.Fatal(err)
		}
		if got := stmt.ColumnText(0); got != want {
			t.Errorf("sort_strlen %d got %s, wanted %s", i, got, want)
		}
	}
	stmt.Finalize()
}
