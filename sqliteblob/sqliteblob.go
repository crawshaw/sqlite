package sqliteblob

import (
	"fmt"
	"io"

	"crawshaw.io/sqlite"
	"crawshaw.io/sqlite/sqlitex"
)

func Insert(conn *sqlite.Conn, db, table, column string, r io.Reader, len int64) (rowid int64, err error) {
	defer sqlitex.Save(conn)(&err)

	defer func() {
		if err == nil {
			return
		}
		if sqErr, isErr := err.(sqlite.Error); isErr {
			sqErr.Loc = "sqliteblob.Insert: " + sqErr.Loc
			err = sqErr
		}
	}()

	if db == "" {
		db = "main"
	}

	stmt := conn.Prep(fmt.Sprintf("INSERT INTO %q.%q (%q) VALUES ($blob);", db, table, column))
	stmt.SetZeroBlob("$blob", len)
	if _, err := stmt.Step(); err != nil {
		return 0, err
	}
	rowid = conn.LastInsertRowID()

	blob, err := conn.OpenBlob(db, table, column, rowid, true)
	if err != nil {
		return 0, err
	}
	n, err := io.Copy(blob, r)
	if closeErr := blob.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return 0, err
	}
	if n != len {
		return 0, fmt.Errorf("sqliteblob.Insert: worte %d bytes, expected %d bytes", n, len)
	}

	return rowid, nil
}

/*func ReadAll(dst io.Writer, conn *sqlite.Conn, db, table, column string, rowid int64) (n int64, err error) {
	blob, err := conn.OpenBlob(db, table, column, rowid, false)
	if err != nil {
		return 0, err
	}
	n, err = io.Copy(dst, blob)
	closeErr := blob.Close()
	if err == nil {
		err = closeErr
	}
	return n, err
}*/

/*func Update(conn *sqlite.Conn, db, table, column string, row int64, r io.Reader) error {
}
*/
