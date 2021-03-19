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

package sqlite

import (
	"testing"
)

func TestIncrementor(t *testing.T) {
	start := 5
	i := NewIncrementor(start)
	if i == nil {
		t.Fatal("Incrementor returned nil")
	}
	if i() != start {
		t.Fatalf("first call did not start at %v", start)
	}
	for j := 1; j < 10; j++ {
		if i() != start+j {
			t.Fatalf("%v call did not return %v+%v", j, start, j)
		}
	}

	b := BindIncrementor()
	if b() != 1 {
		t.Fatal("BindIncrementor does not start at 1")
	}

	c := ColumnIncrementor()
	if c() != 0 {
		t.Fatal("ColumnIncrementor does not start at 0")
	}
}
