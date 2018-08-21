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
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"crawshaw.io/sqlite"
)

const (
	extC = `
/*
** Copyright (c) 2018 David Crawshaw <david@zentus.com>
**
** Permission to use, copy, modify, and distribute this software for any
** purpose with or without fee is hereby granted, provided that the above
** copyright notice and this permission notice appear in all copies.
**
** THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
** WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
** MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
** ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
** WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
** ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
** OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
*/

#include <sqlite3ext.h>
SQLITE_EXTENSION_INIT1

#include <stdlib.h>

extern void hellofunc(sqlite3_context *context, int argc, sqlite3_value **argv);

#ifdef _WIN32
__declspec(dllexport)
#endif
int sqlite3_hello_init(
 sqlite3 *db,
 char **pzErrMsg,
 const sqlite3_api_routines *pApi
){
   int rc = SQLITE_OK;
   SQLITE_EXTENSION_INIT2(pApi);
   (void)pzErrMsg;  /* Unused parameter */
   return sqlite3_create_function(db, "hello", 0, SQLITE_UTF8, 0, hellofunc, 0, 0);
}`

	extGo = `
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

package main

// #cgo CFLAGS: -I../
// #cgo LDFLAGS: -ldl -lsqlite3
// #include "sqlite3.h"
// #include <stdlib.h>
import "C"

//export hellofunc
func hellofunc(context *C.sqlite3_context, argc C.int, argv **C.sqlite3_value) {
	C.sqlite3_result_text(context, C.CString("Hello, World!"), -1, (*[0]byte)(C.free))
}

func main() {}`
)

func libext(t *testing.T) string {
	t.Helper()
	switch runtime.GOOS {
	case "darwin":
		return "dylib"
	case "linux":
		return "so"
	case "windows":
		return "dll"
	}
	t.Skip("os not supported")
	return ""
}

func TestLoadExtension(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "sqlite")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)
	exts := []string{"c", "go"}
	for i, ext := range []string{extC, extGo} {
		fout, err := os.Create(filepath.Join(tmpdir, "ext.") + exts[i])
		if err != nil {
			t.Fatal(err)
		}
		io.Copy(fout, strings.NewReader(ext))
		fout.Close()
	}
	cmd := exec.Command(
		"go",
		"build",
		"-buildmode=c-shared",
		"-o", "libhello."+libext(t),
	)
	cmd.Dir = tmpdir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatal(string(out), err)
	}
	c, err := sqlite.OpenConn(":memory:", 0)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := c.Close()
		if err != nil {
			t.Error(nil)
		}
	}()
	libPath := filepath.Join(tmpdir, "libhello."+libext(t))
	err = c.LoadExtension(libPath, "")
	if err == nil {
		t.Error("loaded extension without enabling load extension")
	}
	err = c.EnableLoadExtension(true)
	if err != nil {
		t.Fatal(err)
	}
	err = c.LoadExtension(libPath, "")
	if err != nil {
		t.Fatal(err)
	}
	stmt := c.Prep("SELECT hello();")
	if _, err := stmt.Step(); err != nil {
		t.Fatal(err)
	}
	if got, want := stmt.ColumnText(0), "Hello, World!"; got != want {
		t.Error("failed to load extension, got: %s, want: %s", got, want)
	}
}
