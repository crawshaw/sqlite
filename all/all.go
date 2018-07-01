// Package all provides the C object code for SQLite.
//
// It is built with a wide range of SQLite features.
package all

// #include <sqlite3.h>
// #cgo CFLAGS: -DSQLITE_THREADSAFE=2
// #cgo CFLAGS: -DSQLITE_DEFAULT_WAL_SYNCHRONOUS=1
// #cgo CFLAGS: -DSQLITE_ENABLE_UNLOCK_NOTIFY
// #cgo CFLAGS: -DSQLITE_ENABLE_FTS5
// #cgo CFLAGS: -DSQLITE_ENABLE_RTREE
// #cgo CFLAGS: -DSQLITE_LIKE_DOESNT_MATCH_BLOBS
// #cgo CFLAGS: -DSQLITE_OMIT_DEPRECATED
// #cgo CFLAGS: -DSQLITE_ENABLE_JSON1
// #cgo CFLAGS: -DSQLITE_ENABLE_SESSION
// #cgo CFLAGS: -DSQLITE_ENABLE_PREUPDATE_HOOK
// #cgo CFLAGS: -DSQLITE_USE_ALLOCA
// #cgo CFLAGS: -DSQLITE_ENABLE_COLUMN_METADATA
// #cgo windows LDFLAGS: -Wl,-Bstatic -lwinpthread -Wl,-Bdynamic
// #cgo linux LDFLAGS: -ldl -lm
// #cgo linux CFLAGS: -std=c99
// #cgo openbsd LDFLAGS: -lm
// #cgo openbsd CFLAGS: -std=c99
import "C"

import _ "unsafe"

var _ = C.sqlite3_open_v2

var _ = C.sqlite3_unlock_notify
