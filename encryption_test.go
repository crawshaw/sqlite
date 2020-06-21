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
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"crawshaw.io/sqlite"
	"crawshaw.io/sqlite/sqlitex"
	"github.com/stretchr/testify/require"
)

func TestEncryption(t *testing.T) {
	dir, err := ioutil.TempDir("", "crawshaw.io")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	dbFile := filepath.Join(dir, "encrypted.sqlite3")

	// Open a new DB and set up encryption. Create a table and insert some
	// data, then close the DB.
	require := require.New(t)
	// Using WAL mode requires supplying the key as a URI param, which
	// prevents us from testing Conn.Unlock.
	openFlags := sqlite.OpenFlagsDefault ^ sqlite.SQLITE_OPEN_WAL
	conn, err := sqlite.OpenConn(dbFile, openFlags)
	require.NoErrorf(err, "sqlite.OpenConn(%q, 0)", dbFile)
	defer func() {
		if conn != nil {
			conn.Close()
		}
	}()

	key := []byte("password")

	if sqlite.CLib == sqlite.CLibSQLite {
		require.Error(conn.SetEncryptionKey(key))
		require.Error(conn.Unlock(key))
		return
	}

	require.NoError(sqlitex.ExecScript(conn, `
CREATE TABLE test(a, b, c);
INSERT INTO test VALUES (1, 2, 3);
INSERT INTO test VALUES (4, 5, 6);
`))
	require.NoError(conn.SetEncryptionKey(key))
	err = conn.Close()
	conn = nil
	require.NoError(err)

	conn, err = sqlite.OpenConn("file:"+dbFile, openFlags)
	require.NoErrorf(err, "sqlite.OpenConn(%q, 0)", dbFile)

	require.Error(sqlitex.ExecScript(conn, "SELECT * FROM test;"))
	require.Error(conn.Unlock([]byte("invalid")))
	require.Error(sqlitex.ExecScript(conn, "SELECT * FROM test;"))
	require.Error(conn.SetEncryptionKey([]byte("invalid")))

	require.NoError(conn.Unlock(key))
	require.NoError(sqlitex.ExecScript(conn, "SELECT * FROM test;"))
	require.NoError(conn.SetEncryptionKey(nil))
	err = conn.Close()
	conn = nil
	require.NoError(err)

	conn, err = sqlite.OpenConn("file:"+dbFile, openFlags)
	require.NoErrorf(err, "sqlite.OpenConn(%q, 0)", dbFile)
	require.NoError(sqlitex.ExecScript(conn, "SELECT * FROM test;"))
}
