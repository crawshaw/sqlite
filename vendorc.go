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

// +build vendorc

// This file is a workaround for `go mod vendor` which doesn't vendor
// directories that don't contain a Go file. This prevents the necessary C
// files c/sqlite/sqlite3.c or c/sqleet/sqlite3.c files from being vendored.
//
// This Go file imports the c pseudo package, which imports the specific C
// pseudo packages, so that the `go mod vendor` sees these direcotries as
// required.
//
// The vendorc build tag is used to prevent needlessly building these files.
//
// See this issue for reference: https://github.com/golang/go/issues/26366

package sqlite

import (
	_ "crawshaw.io/sqlite/c"
)
