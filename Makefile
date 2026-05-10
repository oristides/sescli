# sescli — local dev and release helpers
#
# Auto tag + push (CI runs GoReleaser on the tag):
#   make release                 → next patch after latest v* (e.g. v0.3.1 → v0.3.2)
#   make release BUMP=minor      → next minor (v0.3.1 → v0.4.0)
#   make release TAG=v1.0.0      → explicit tag (must sort after latest v*)
#
# Local snapshot (no tag):
#   make snapshot

SHELL := /bin/bash
.ONESHELL:

BUMP ?= patch

# If TAG is empty: next sequential tag from scripts/next-release-tag.sh; else use TAG.
# Use := so the auto tag is computed once per invocation (only when TAG is unset).
RESOLVED_TAG := $(if $(strip $(TAG)),$(TAG),$(shell "$(CURDIR)/scripts/next-release-tag.sh" $(BUMP)))

.PHONY: default help check test vet fmt snapshot release list-tags tags install-tools verify-release-tag

default: help

help:
	@echo "sescli Makefile"
	@echo ""
	@echo "  make check                 gofmt (check) + vet + test ./..."
	@echo "  make snapshot              GoReleaser snapshot → dist/ (no git tag)"
	@echo "  make tags                  v* tags newest-first with dates (see release history)"
	@echo "  make list-tags             all v* tags oldest → newest"
	@echo "  make release               auto next PATCH tag, show recent tags, verify, test, push tag"
	@echo "  make release BUMP=minor    auto next MINOR (… X.Y.0)"
	@echo "  make release BUMP=major    auto next MAJOR (… X.0.0)"
	@echo "  make release TAG=vX.Y.Z    same flow with explicit tag"
	@echo "  make verify-release-tag    dry-run: is RESOLVED_TAG ok? (uses TAG= or auto)"
	@echo "  make install-tools         install goreleaser (for snapshot)"
	@echo ""
	@echo "Sequential semver: auto tag is always one bump after the latest v* (git sort -V)."

check: fmt vet test

test:
	go test ./...

vet:
	go vet ./...

fmt:
	@test -z "$$(gofmt -l . 2>/dev/null | grep -v '^vendor/' || true)" || (echo "run: gofmt -w ."; exit 1)

list-tags:
	@git tag -l 'v*' | sort -V

tags:
	@echo "v* tags (newest first, date = tag creation):"
	@git for-each-ref refs/tags/v* --sort=-version:refname --format '  %(refname:short)  %(creatordate:iso8601)' 2>/dev/null | head -20 || true

snapshot: install-tools
	goreleaser release --snapshot --clean --skip=publish

install-tools:
	@command -v goreleaser >/dev/null 2>&1 || { echo "Installing goreleaser..."; go install github.com/goreleaser/goreleaser/v2@latest; }
	@command -v goreleaser >/dev/null 2>&1

verify-release-tag:
	@./scripts/verify-release-tag.sh "$(RESOLVED_TAG)"

release: check
	@echo ""
	@echo "==> Recent v* tags (newest first):"
	@git for-each-ref refs/tags/v* --sort=-version:refname --format '    %(refname:short)  %(creatordate:iso8601)' 2>/dev/null | head -15 || true
	@echo "==> Planned tag: $(RESOLVED_TAG)"
	@if [ -z "$(strip $(TAG))" ]; then echo "    (automatic next tag; BUMP=$(BUMP); override with TAG=vX.Y.Z or BUMP=minor|major)"; else echo "    (explicit TAG= from command line)"; fi
	@echo ""
	@./scripts/verify-release-tag.sh "$(RESOLVED_TAG)"
	./scripts/release.sh "$(RESOLVED_TAG)"
