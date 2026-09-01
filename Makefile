# Verify targets. Hermes/Claude workers pick up build/lint/test/check from this
# file; keep the target names. See CLAUDE.md.
#
# Tests need only the Go toolchain: MySQL is go-sqlmock (db.NewTestDb) and Redis
# is redigomock (redis.NewRedisMock); no daemon, network or /etc/pram/pram.conf.
# `build` uses ./... so nothing is written to the tree (`go build -o eirka-admin`
# would leave an untracked binary; it is not in .gitignore).

GO ?= go

.PHONY: build lint test check

build:
	$(GO) build ./...

lint:
	@out="$$(gofmt -l $$($(GO) list -f '{{.Dir}}' ./...))"; if [ -n "$$out" ]; then echo "gofmt: unformatted files:"; echo "$$out"; exit 1; fi
	$(GO) vet ./...

test:
	$(GO) test -count=1 ./...

check: build lint test
