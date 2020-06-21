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

// sqlite3.c is a thin wrapper around the real sqlite3.c
//
// This allows the Go library to build against various SQLite3 C versions.
//
// The CGO macro is set in static.go and prevents Go from building the real
// sqlite3.c when linking against a prebuilt object file.
//
// The macro CGO_SQLEET is defined by sqleet.go if the sqleet go build tag is
// set.

#ifdef CGO
#ifdef CGO_SQLEET
#include "c/sqleet/sqleet.c"
#else
#include "c/sqlite/sqlite3.c"
#endif
#endif
