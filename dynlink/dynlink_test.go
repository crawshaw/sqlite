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

package dynlink_test

import (
	"testing"

	"crawshaw.io/sqlite"
	_ "crawshaw.io/sqlite/dynlink"
	"crawshaw.io/sqlite/sqliteutil"
)

func TestConn(t *testing.T) {
	conn, err := sqlite.OpenConn(":memory:", 0)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			t.Error(err)
		}
	}()

	err = sqliteutil.ExecScript(conn, `
		DROP TABLE IF EXISTS t;
		CREATE TABLE t (c);
		INSERT INTO t VALUES (1);
		INSERT INTO t VALUES (2);
		INSERT INTO t VALUES (3);
	`)
	if err != nil {
		t.Error(err)
	}

	count, err := sqliteutil.ResultInt(conn.Prep("SELECT count(*) FROM t;"))
	if err != nil {
		t.Error(err)
	}
	if count != 3 {
		t.Errorf("want 3 rows, got %d", count)
	}
}
