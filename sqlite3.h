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

// sqlite3.h is a thin wrapper around the real sqlite3.h
//
// This allows the Go library to build against various SQLite3 C versions.
//
// The macro CGO_SQLEET is defined if the sqleet go build tag is set.
#ifndef CGO_SQLITE
#ifdef CGO_SQLEET
#define CGO_SQLITE "sqleet"
#include "c/sqleet/sqleet.h"
#else
#define CGO_SQLITE "sqlite3"
#include "c/sqlite/sqlite3.h"
#endif
#endif
