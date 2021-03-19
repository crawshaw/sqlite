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

package sqlite

// #include <stdint.h>
// #include <sqlite3.h>
// #include "wrappers.h"
//
// static int go_sqlite3_create_collation_v2(
//   sqlite3 *db,
//   const char *zName,
//   int eTextRep,
//   uintptr_t pApp,
//   int (*xCompare)(void*,int,const void*,int,const void*),
//   void (*xDestroy)(void*)
// ) {
//   return sqlite3_create_collation_v2(
//     db,
//     zName,
//     eTextRep,
//     (void*)pApp,
//     xCompare,
//     xDestroy);
// }
import "C"
import (
	"sync"
)

type xcollation struct {
	id       int
	name     string
	conn     *Conn
	xCompare func(string, string) int
}

var xcollations = struct {
	mu   sync.Mutex
	m    map[int]*xcollation
	next int
}{
	m: make(map[int]*xcollation),
}

// CreateCollation registers a Go function as a SQLite collation function.
//
// These function are used with the COLLATE operator to implement custom sorting in queries.
//
// The xCompare function must return an integer that is negative, zero, or positive if the first
// string is less than, equal to, or greater than the second, respectively. The function must
// always return the same result for the same inputs and must be commutative.
//
// These are the same properties as strings.Compare().
//
// https://sqlite.org/datatype3.html#collation
// https://sqlite.org/c3ref/create_collation.html
func (conn *Conn) CreateCollation(name string, xCompare func(string, string) int) error {
	cname := C.CString(name)
	eTextRep := C.int(C.SQLITE_UTF8)

	x := &xcollation{
		name:     name,
		conn:     conn,
		xCompare: xCompare,
	}

	xcollations.mu.Lock()
	xcollations.next++
	x.id = xcollations.next
	xcollations.m[x.id] = x
	xcollations.mu.Unlock()

	res := C.go_sqlite3_create_collation_v2(
		conn.conn,
		cname,
		eTextRep,
		C.uintptr_t(x.id),
		(*[0]byte)(C.c_collation_tramp),
		(*[0]byte)(C.c_destroy_collation_tramp),
	)
	return conn.reserr("Conn.CreateCollation", name, res)
}

//export go_collation_tramp
func go_collation_tramp(ptr uintptr, aLen C.int, a *C.char, bLen C.int, b *C.char) C.int {
	xcollations.mu.Lock()
	x := xcollations.m[int(ptr)]
	xcollations.mu.Unlock()
	return C.int(x.xCompare(C.GoStringN((*C.char)(a), aLen), C.GoStringN((*C.char)(b), bLen)))
}

//export go_destroy_collation_tramp
func go_destroy_collation_tramp(ptr uintptr) {
	id := int(ptr)
	xcollations.mu.Lock()
	delete(xcollations.m, id)
	xcollations.mu.Unlock()
}
