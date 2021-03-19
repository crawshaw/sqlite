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

package sqlitex

import (
	"fmt"
	"io"

	"crawshaw.io/sqlite"
)

// BlobList is the minimum implmentation required to permit File to
// store data as a collection of sqlite blobs (which may or may not be
// stored in an sqlite database file)
type BlobList interface {
	//Add adds a blob of a certain size to a collection and returns an open
	// sqlite.Blob ready for reading/writing
	Add(size int64) (*sqlite.Blob, error)

  // Get returns the sqlite.Blob at index 
  Get(index int) *sqlite.Blob

  // Delete closes and deletes the sqlite.Blob at index
	Delete(index int) error

	// Count returns the number of sqlite.Blobs in the collection
	Count() int
}

var _ BlobList = (*tmpStore)(nil)

// tmpStore implements BlobList to tmpStore
type tmpStore struct {
	err    error
	conn   *sqlite.Conn
	blobs  []*sqlite.Blob
	rowids []int64
}

// NewstoreSize returns an initialized tmpStore
func NewstoreSize(conn *sqlite.Conn, initSize int) (BlobList, error) {
	st := &tmpStore{conn: conn}

	stmt := conn.Prep("CREATE TEMP TABLE IF NOT EXISTS BlobBuffer (blob BLOB);")
	if _, err := stmt.Step(); err != nil {
		return nil, err
	}
	if _, err := st.Add(int64(initSize)); err != nil {
		return nil, err
	}
	return st, nil
}

// Add appends a blob of a certain size
func (st *tmpStore) Add(size int64) (*sqlite.Blob, error) {
	stmt := st.conn.Prep("INSERT INTO BlobBuffer (blob) VALUES ($blob);")
	stmt.SetZeroBlob("$blob", size)
	if _, err := stmt.Step(); err != nil {
		return nil, err
	}
	rowid := st.conn.LastInsertRowID()
	blob, err := st.conn.OpenBlob("temp", "BlobBuffer", "blob", rowid, true)
	if err != nil {
		return nil, err
	}
	st.blobs = append(st.blobs, blob)
	st.rowids = append(st.rowids, rowid)
	return blob, nil
}

// Delete deletes a blob
func (st *tmpStore) Delete(index int) error {
	err := st.blobs[index].Close()
	if err != nil {
		return err
	}
	stmt := st.conn.Prep("DELETE FROM BlobBuffer WHERE rowid = $rowid;")
	stmt.SetInt64("$rowid", st.rowids[index])
	if _, err := stmt.Step(); err != nil {
		return err
	}
	return nil
}

// Count returns the number of allocated blobs
func (st *tmpStore) Count() int {
	return len(st.blobs)
}

// Get returns the blob
func (st *tmpStore) Get(index int) *sqlite.Blob {
	return st.blobs[index]
}

// File is a readable, writable, and seekable series of temporary SQLite blobs.
type File struct {
	io.Reader
	io.Writer
	io.Seeker
	st  BlobList
	err error
	off bbpos
	len bbpos
}

func NewFile(conn *sqlite.Conn) (*File, error) {
	return NewFileSize(conn, 16*1024)
}

func NewFileSize(conn *sqlite.Conn, initSize int) (*File, error) {
	st, err := NewstoreSize(conn, initSize)
	if err != nil {
		return nil, err
	}
	return &File{st: st}, nil
}

// grow adds an sqlite blob if the buffer is out of space.
func (bb *File) grow() error {
	lastSize := bb.st.Get(bb.st.Count() - 1).Size()
	size := lastSize * 2
	if _, err := bb.st.Add(size); err != nil {
		return err
	}
	return nil
}

// rem reports the remaining available bytes in the pointed-to blob
func (bb *File) rem(pos bbpos) int64 {
	return bb.st.Get(pos.index).Size() - pos.pos
}

func (bb *File) eq(p1, p2 bbpos) bool {
	if p1 == p2 {
		return true
	}
	if p1.index == p2.index+1 && bb.rem(p1) == 0 && p2.pos == 0 {
		return true
	}
	if p2.index == p1.index+1 && bb.rem(p2) == 0 && p1.pos == 0 {
		return true
	}
	return false
}

func (bb *File) gt(p1, p2 bbpos) bool {
	if bb.eq(p1, p2) {
		return false
	}
	if p1.index != p2.index {
		return p1.index > p2.index
	}
	return p1.pos > p2.pos
}

func (bb *File) zero(p1, p2 bbpos) error {
	var zeros [4096]byte
	for bb.gt(p2, p1) {
		w := bb.rem(p1)
		if w == 0 {
			p1.index++
			p1.pos = 0
			w = bb.rem(p1)
		}
		if p1.index == p2.index {
			w = p2.pos
		}
		if w > int64(len(zeros)) {
			w = int64(len(zeros))
		}
		nn, err := bb.st.Get(p1.index).WriteAt(zeros[:w], p1.pos)
		if err != nil {
			return err
		}
		p1.pos += int64(nn)
	}
	return nil
}

func (bb *File) Write(p []byte) (n int, err error) {
	if bb.err != nil {
		return 0, bb.err
	}

	if bb.gt(bb.off, bb.len) {
		if err := bb.zero(bb.len, bb.off); err != nil {
			bb.err = err
			return 0, err
		}
	}

	for len(p) > 0 {
		w := bb.rem(bb.off)
		if w == 0 {
			if bb.off.index == bb.st.Count()-1 {
				if bb.err = bb.grow(); bb.err != nil {
					return n, bb.err
				}
			}
			bb.off.index++
			bb.off.pos = 0
			w = bb.rem(bb.off)
		}
		if int64(len(p)) < w {
			w = int64(len(p))
		}
		nn, err := bb.st.Get(bb.off.index).WriteAt(p[:w], bb.off.pos)
		n += nn
		p = p[nn:]
		bb.off.pos += int64(nn)
		if bb.gt(bb.off, bb.len) {
			bb.len = bb.off
		}
		if err != nil {
			bb.err = err
			break
		}
	}

	return n, bb.err
}

func (bb *File) Read(p []byte) (n int, err error) {
	if bb.err != nil {
		return 0, bb.err
	}

	for len(p) > 0 && bb.gt(bb.len, bb.off) {
		if bb.rem(bb.off) == 0 {
			bb.off.index++
			bb.off.pos = 0
		}

		var bsize int64
		if bb.len.index == bb.off.index {
			bsize = bb.len.pos
		} else {
			bsize = bb.st.Get(bb.off.index).Size()
		}
		w := bsize - bb.off.pos
		if int64(len(p)) < w {
			w = int64(len(p))
		}
		nn, err := bb.st.Get(bb.off.index).ReadAt(p[:w], bb.off.pos)
		n += nn
		p = p[nn:]
		bb.off.pos += int64(nn)
		if err != nil {
			bb.err = err
			return n, err
		}
	}

	if n == 0 && (bb.eq(bb.off, bb.len) || bb.gt(bb.off, bb.len)) {
		return 0, io.EOF
	}

	return n, nil
}

func (bb *File) Seek(offset int64, whence int) (int64, error) {
	if bb.err != nil {
		return 0, bb.err
	}

	const (
		SeekStart   = 0
		SeekCurrent = 1
		SeekEnd     = 2
	)
	switch whence {
	case SeekStart:
		// use offset directly
	case SeekCurrent:
		for i := 0; i < bb.off.index; i++ {
			offset += bb.st.Get(i).Size()
		}
		offset += bb.off.pos
	case SeekEnd:
		offset += bb.Len()
	}
	if offset < 0 {
		return -1, fmt.Errorf("sqlitex.File: attempting to seek before beginning of blob (%d)", offset)
	}

	rem := offset
	bb.off.index = 0
	for i := 0; rem > bb.st.Get(i).Size(); i++ {
		bb.off.index = i + 1
		rem -= bb.st.Get(i).Size()

		if i == bb.st.Count()-1 {
			if err := bb.grow(); err != nil {
				return offset - rem, err
			}
		}
	}
	bb.off.pos = rem

	return offset, nil
}

func (bb *File) Truncate(size int64) error {
	for {
		for i := 0; i < bb.st.Count(); i++ {
			bsize := bb.st.Get(i).Size()
			if bsize > size {
				newlen := bbpos{index: i, pos: size}
				if err := bb.zero(bb.len, newlen); err != nil {
					return err
				}
				bb.len = newlen
				return nil
			}
			size -= bsize
		}
		if err := bb.grow(); err != nil {
			return err
		}
	}
}

func (bb *File) Len() (n int64) {
	for i := 0; i < bb.len.index; i++ {
		n += bb.st.Get(i).Size()
	}
	n += bb.len.pos
	return n
}

func (bb *File) Cap() (n int64) {
	for i := 0; i < bb.st.Count(); i++ {
		n += bb.st.Get(i).Size()
	}
	return n
}

func (bb *File) Close() error {
	if bb.err != nil {
		return bb.err
	}
	for i := 0; i < bb.st.Count(); i++ {
		err := bb.st.Delete(i)
		if bb.err == nil {
			bb.err = err
		}
	}
	return bb.err
}

type bbpos struct {
	index int   // bb.blobs[index]
	pos   int64 // point inside that blob
}
