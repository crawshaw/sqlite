## Copyright (c) 2018 David Crawshaw <david@zentus.com>
##
## Permission to use, copy, modify, and distribute this software for any
## purpose with or without fee is hereby granted, provided that the above
## copyright notice and this permission notice appear in all copies.
##
## THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
## WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
## MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
## ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
## WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
## ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
## OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.

# This Makefile is simply for development purposes. Normally, when this package
# is imported, Go will build the ./c/sqlite.c file that is included directly by
# static.go. However this is pretty slow ~30s. When developing this is very
# annoying. Use this Makefile to pre-build the sqlite3.o object and then build
# the package with the build tag link, which will ignore static.go and
# use link.go instead to link against sqlite.o. This reduces compilation times
# down to < 3 sec!
#
# If you are using an editor that builds the project as you work on it, you'll
# want to build the sqlite3.o object and tell your editor to use the
# link go build tag when working on this project.
# For vim-go, use the command `GoBuildTags link` or
# 	`let g:go_build_tags = # 'link'`

.PHONY: clean all env test sqleet test-all
all: sqlite3.o
	go build -tags=link ./...

test-all: test test-sqleet

test: sqlite3.o
	go test -tags=link ./...

sqleet: sqleet.o
	go build -tags=link,sqleet ./...

test-sqleet: sqleet.o
	go test -tags=link,sqleet ./

# Paths to look for source files
VPATH = ./c/sqlite ./c/sqleet

# !!! THESE DEFINES SHOULD MATCH sqlite.go for linux !!!
CFLAGS += -std=c99
CFLAGS += -DSQLITE_THREADSAFE=2
CFLAGS += -DSQLITE_DEFAULT_WAL_SYNCHRONOUS=1
CFLAGS += -DSQLITE_ENABLE_UNLOCK_NOTIFY
CFLAGS += -DSQLITE_ENABLE_FTS5
CFLAGS += -DSQLITE_ENABLE_RTREE
CFLAGS += -DSQLITE_LIKE_DOESNT_MATCH_BLOBS
CFLAGS += -DSQLITE_OMIT_DEPRECATED
CFLAGS += -DSQLITE_ENABLE_JSON1
CFLAGS += -DSQLITE_ENABLE_SESSION
CFLAGS += -DSQLITE_ENABLE_SNAPSHOT
CFLAGS += -DSQLITE_ENABLE_PREUPDATE_HOOK
CFLAGS += -DSQLITE_USE_ALLOCA
CFLAGS += -DSQLITE_ENABLE_COLUMN_METADATA
CFLAGS += -DHAVE_USLEEP=1
CFLAGS += -DSQLITE_DQS=0
CFLAGS += -DSQLITE_ENABLE_GEOPOLY
LDFLAGS = -ldl -lm
# !!! THESE DEFINES SHOULD MATCH sqlite.go !!!

clean:
	rm -f sqlite3.o sqleet.o
