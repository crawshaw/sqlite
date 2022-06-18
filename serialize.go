package sqlite

// #include <sqlite3.h>
// #include <stdlib.h>
import "C"
import (
	"runtime"
	"unsafe"
)

// Serialized contains schema and serialized data for a database.
type Serialized struct {
	schema         string
	data           []byte
	sqliteOwnsData bool
	shouldFreeData bool
}

// NewSerialized creates a new serialized DB from the given schema and data.
//
// If copyToSqlite is true, the data will be copied. This should be set to true
// if this will be used with SQLITE_DESERIALIZE_FREEONCLOSE.
func NewSerialized(schema string, data []byte, copyToSqlite bool) *Serialized {
	s := &Serialized{
		schema: schema,
		data:   data,
	}
	if copyToSqlite {
		sqliteData := (*[1 << 28]uint8)(unsafe.Pointer(C.sqlite3_malloc(C.int(len(data)))))[:len(data):len(data)]
		copy(sqliteData, data)
		s.data = sqliteData
		s.shouldFreeData = true
		s.sqliteOwnsData = true
		runtime.SetFinalizer(s, func(s *Serialized) { s.free() })
	}
	return s
}

// Schema returns the schema for this serialized DB.
func (s *Serialized) Schema() string {
	if s.schema == "" {
		return "main"
	}
	return s.schema
}

// Bytes returns the serialized bytes. Do not mutate this value. This is only
// valid for the life of its receiver and should be copied for any other
// longer-term use.
func (s *Serialized) Bytes() []byte { return s.data }

func (s *Serialized) free() {
	if len(s.data) > 0 && s.shouldFreeData {
		s.shouldFreeData = false
		s.sqliteOwnsData = false
		C.sqlite3_free(unsafe.Pointer(&s.data[0]))
	}
	s.data = nil
}

// SerializeFlags are flags used for Serialize.
type SerializeFlags int

const (
	SQLITE_SERIALIZE_NOCOPY SerializeFlags = C.SQLITE_SERIALIZE_NOCOPY
)

// Serialize serializes the given schema. Returns nil on error. If
// SQLITE_SERIALIZE_NOCOPY flag is set, the data may only be valid as long as
// the database.
//
// https://www.sqlite.org/c3ref/serialize.html
func (conn *Conn) Serialize(schema string, flags ...SerializeFlags) *Serialized {
	var serializeFlags SerializeFlags
	for _, f := range flags {
		serializeFlags |= f
	}

	cschema := cmain
	if schema != "" && schema != "main" {
		cschema = C.CString(schema)
		defer C.free(unsafe.Pointer(cschema))
	}

	var csize C.sqlite3_int64
	res := C.sqlite3_serialize(conn.conn, cschema, &csize, C.uint(serializeFlags))
	if res == nil {
		return nil
	}

	s := &Serialized{
		schema:         schema,
		data:           (*[1 << 28]uint8)(unsafe.Pointer(res))[:csize:csize],
		sqliteOwnsData: true,
	}
	// Free the memory only if they didn't specify nocopy
	if serializeFlags&SQLITE_SERIALIZE_NOCOPY == 0 {
		s.shouldFreeData = true
		runtime.SetFinalizer(s, func(s *Serialized) { s.free() })
	}
	return s
}

// DeserializeFlags are flags used for Deserialize.
type DeserializeFlags int

const (
	SQLITE_DESERIALIZE_FREEONCLOSE DeserializeFlags = C.SQLITE_DESERIALIZE_FREEONCLOSE
	SQLITE_DESERIALIZE_RESIZEABLE  DeserializeFlags = C.SQLITE_DESERIALIZE_RESIZEABLE
	SQLITE_DESERIALIZE_READONLY    DeserializeFlags = C.SQLITE_DESERIALIZE_READONLY
)

// Reopens the database as in-memory representation of given serialized bytes.
// The given *Serialized instance should remain referenced (i.e. not GC'd) for
// the life of the DB since the bytes within are referenced directly.
//
// Callers should only use SQLITE_DESERIALIZE_FREEONCLOSE and
// SQLITE_DESERIALIZE_RESIZEABLE if the param came from Serialize or
// copyToSqlite was given to NewSerialized.
//
// The Serialized parameter should no longer be used after this call.
//
// https://www.sqlite.org/c3ref/deserialize.html
func (conn *Conn) Deserialize(s *Serialized, flags ...DeserializeFlags) error {
	var deserializeFlags DeserializeFlags
	for _, f := range flags {
		deserializeFlags |= f
	}

	cschema := cmain
	if s.schema != "" && s.schema != "main" {
		cschema = C.CString(s.schema)
		defer C.free(unsafe.Pointer(cschema))
	}

	// If they set to free on close, remove the free flag from the param
	if deserializeFlags&SQLITE_DESERIALIZE_FREEONCLOSE == 1 {
		s.shouldFreeData = false
	}

	res := C.sqlite3_deserialize(
		conn.conn,
		cschema,
		(*C.uchar)(unsafe.Pointer(&s.data[0])),
		C.sqlite3_int64(len(s.data)),
		C.sqlite3_int64(len(s.data)),
		C.uint(deserializeFlags),
	)
	if res != C.SQLITE_OK {
		return conn.extreserr("Conn.Deserialize", "", res)
	}
	return nil
}
