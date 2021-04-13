package sqlitex

import (
	"errors"

	"crawshaw.io/sqlite"
)

var ErrNoResults = errors.New("sqlite: statement has no results")
var ErrMultipleResults = errors.New("sqlite: statement has multiple result rows")

func resultSetup(stmt *sqlite.Stmt) error {
	hasRow, err := stmt.Step()
	if err != nil {
		stmt.Reset()
		return err
	}
	if !hasRow {
		stmt.Reset()
		return ErrNoResults
	}
	return nil
}

func resultTeardown(stmt *sqlite.Stmt) error {
	hasRow, err := stmt.Step()
	if err != nil {
		stmt.Reset()
		return err
	}
	if hasRow {
		stmt.Reset()
		return ErrMultipleResults
	}
	return stmt.Reset()
}

// ResultInt steps the Stmt once and returns the first column as an int.
//
// If there are no rows in the result set, errors.Is(ErrNoResults) will
// be true.
//
// If there are multiple rows, errors.Is(ErrMultipleResults) will be true.
//
func ResultInt(stmt *sqlite.Stmt) (int, error) {
	res, err := ResultInt64(stmt)
	return int(res), err
}

// ResultInt64 steps the Stmt once and returns the first column as an int64.
//
// If there are no rows in the result set, errors.Is(ErrNoResults) will
// be true.
//
// If there are multiple rows, errors.Is(ErrMultipleResults) will be true.
//
func ResultInt64(stmt *sqlite.Stmt) (int64, error) {
	if err := resultSetup(stmt); err != nil {
		return 0, err
	}
	res := stmt.ColumnInt64(0)
	if err := resultTeardown(stmt); err != nil {
		return 0, err
	}
	return res, nil
}

// ResultText steps the Stmt once and returns the first column as a string.
//
// If there are no rows in the result set, errors.Is(ErrNoResults) will
// be true.
//
// If there are multiple rows, errors.Is(ErrMultipleResults) will be true.
//
func ResultText(stmt *sqlite.Stmt) (string, error) {
	if err := resultSetup(stmt); err != nil {
		return "", err
	}
	res := stmt.ColumnText(0)
	if err := resultTeardown(stmt); err != nil {
		return "", err
	}
	return res, nil
}

// ResultFloat steps the Stmt once and returns the first column as a loat.
//
// If there are no rows in the result set, errors.Is(ErrNoResults) will
// be true.
//
// If there are multiple rows, errors.Is(ErrMultipleResults) will be true.
//
func ResultFloat(stmt *sqlite.Stmt) (float64, error) {
	if err := resultSetup(stmt); err != nil {
		return 0, err
	}
	res := stmt.ColumnFloat(0)
	if err := resultTeardown(stmt); err != nil {
		return 0, err
	}
	return res, nil
}
