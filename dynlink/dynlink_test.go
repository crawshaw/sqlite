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
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"crawshaw.io/sqlite"
	_ "crawshaw.io/sqlite/dynlink"
	"crawshaw.io/sqlite/sqliteutil"
)

func TestConn(t *testing.T) {
	checkLibInstalled(t)

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

func checkLibInstalled(t *testing.T) {
	dir, err := ioutil.TempDir("", "sqlite-dynlink-")
	if err != nil {
		t.Fatal(err)
	}
	const data = "int main(void) { return 1; }"
	emptyC := filepath.Join(dir, "empty.c")
	if err := ioutil.WriteFile(emptyC, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	ccb, err := exec.Command("go", "env", "CC").CombinedOutput()
	if err != nil {
		t.Fatalf("cannot find CC: %v", err)
	}
	cc := strings.TrimSpace(string(ccb))
	outFlag := "-o" + filepath.Join(dir, "emptybin")
	if out, err := exec.Command(cc, outFlag, emptyC).CombinedOutput(); err != nil {
		t.Fatalf("base compilation failed: %v, %s", err, out)
	}
	if _, err := exec.Command(cc, "-lsqlite3", outFlag, emptyC).CombinedOutput(); err != nil {
		t.Skip("no sqlite library installed, skipping dynamic linking test")
	}
}
