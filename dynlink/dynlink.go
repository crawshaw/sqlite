// Package dynlink dynamically links against an installed SQLite library.
package dynlink

// #cgo CFLAGS: -DSQLITE_THREADSAFE=2
// #cgo CFLAGS: -DSQLITE_DEFAULT_WAL_SYNCHRONOUS=1
// #cgo CFLAGS: -DSQLITE_ENABLE_FTS5
// #cgo CFLAGS: -DSQLITE_ENABLE_RTREE
// #cgo CFLAGS: -DSQLITE_LIKE_DOESNT_MATCH_BLOBS
// #cgo CFLAGS: -DSQLITE_OMIT_DEPRECATED
// #cgo CFLAGS: -DSQLITE_ENABLE_JSON1
// #cgo CFLAGS: -DSQLITE_ENABLE_SESSION
// #cgo CFLAGS: -DSQLITE_ENABLE_PREUPDATE_HOOK
// #cgo CFLAGS: -DSQLITE_USE_ALLOCA
// #cgo CFLAGS: -DSQLITE_ENABLE_COLUMN_METADATA
// #cgo linux LDFLAGS: -ldl -lm -lsqlite3
// #cgo linux CFLAGS: -std=c99
// #cgo darwin LDFLAGS: -lsqlite3
//
// #include <sqlite3.h>
//
// // sqlite3_unlock_notify not supported in libsqlite3.dylib on Mac OS X 10.13.
// // This is a dummy implementation that immediately returns SQLITE_LOCKED
// // for any contended statement step.
// //
// // TODO: use the real sqlite3_unlock_notify as soon as possible.
// int sqlite3_unlock_notify(sqlite3 *p, void (*xNotify)(void **apArg, int nArg), void *pNotifyArg) {
//	return SQLITE_LOCKED;
//}
import "C"

import _ "unsafe"

var _ = C.sqlite3_open_v2
