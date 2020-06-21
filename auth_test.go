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
	"testing"

	"crawshaw.io/sqlite"
)

func TestSetAuthorizer(t *testing.T) {
	c, err := sqlite.OpenConn(":memory:", 0)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := c.Close(); err != nil {
			t.Error(err)
		}
	}()

	authResult := sqlite.AuthResult(0)
	var lastAction sqlite.OpType
	auth := sqlite.AuthorizeFunc(func(info sqlite.ActionInfo) sqlite.AuthResult {
		lastAction = info.Action
		return authResult
	})
	c.SetAuthorizer(auth)

	t.Run("Allowed", func(t *testing.T) {
		authResult = 0
		stmt, _, err := c.PrepareTransient("SELECT 1;")
		if err != nil {
			t.Fatal(err)
		}
		stmt.Finalize()
		if lastAction != sqlite.SQLITE_SELECT {
			t.Errorf("action = %q; want SQLITE_SELECT", lastAction)
		}
	})

	t.Run("Denied", func(t *testing.T) {
		authResult = sqlite.SQLITE_DENY
		stmt, _, err := c.PrepareTransient("SELECT 1;")
		if err == nil {
			stmt.Finalize()
			t.Fatal("PrepareTransient did not return an error")
		}
		if got, want := sqlite.ErrCode(err), sqlite.SQLITE_AUTH; got != want {
			t.Errorf("sqlite.ErrCode(err) = %v; want %v", got, want)
		}
		if lastAction != sqlite.SQLITE_SELECT {
			t.Errorf("action = %q; want SQLITE_SELECT", lastAction)
		}
	})
}
