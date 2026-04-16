# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

A Go CLI that updates Cloudflare A/AAAA records from an IP detected on a local network interface or a user-supplied shell command. Runs one-shot or as a daemon (`--run-every`), designed to be driven by the shipped systemd timer. Experimental multihost mode lets several hosts share one domain, using the Cloudflare record `comment` field as an ownership tag.

## Commands

Primary via `go-task` (`Taskfile.yaml`):

```sh
task go:build                  # goreleaser snapshot build → bin/cloudflare-dynamic-dns
task go:run -- <args>          # build and run with CLI args
task go:lint                   # golangci-lint + go vet + gosec (parallel)
task go:fmt                    # goimports-reviser then gofumpt -w .
task docs:update-readme        # regenerates README usage block from --help
task misc:build:goreleaser-config  # regenerates .goreleaser.yml from .goreleaser.ytt.yml
```

Plain Go:

```sh
go build -o cloudflare-dynamic-dns main.go
go test -v ./...   # no *_test.go files exist yet; this is the standard command for when they do
```

Dev environment: `nix develop` (or direnv `use flake`) provides go, go-task, gofumpt, goimports-reviser, golangci-lint, goreleaser, gosec, nix-update, ytt.

Pre-commit hooks via `lefthook`: `go vet`, `gofumpt -l` (check only), `golangci-lint` on staged `.go` files; commitlint on commit-msg.

## Architecture

Single cobra command (no subcommands). Entry: `main.go` → `cmd.NewRootCmd(version, commit, date).Execute()`.

- **`cmd/cmd.go`** — root command, flag registration, viper setup (`CFDDNS_` env prefix, `~/.cloudflare-dynamic-dns.yaml` default config), `collectConfiguration()` (validates + computes state-file name via FNV-1a hash of domains and DNS-affecting settings), and `run(cfg)` loop (repeats on `--run-every`).
- **`cmd/ip.go`** — IP selection. `ipStack` interface (`ipv4Stack`/`ipv6Stack`), interface enumeration or `--ipcmd` shell execution (via `internal/execext`). 64-bit scoring: `[reserved 16 | priority-subnet 32 | base-score 16]`, preferring GUA > ULA, EUI-64 > random (IPv6) and public > Shared-Address-Space > private (IPv4). `--priority-subnets` adjusts the middle bits.
- **`cmd/cloudflare.go`** — zone resolution via `publicsuffix.Domain` + `api.ZoneIDByName`, then create/update/delete records. Multihost mode stores `"<host-id> (managed by cloudflare-dynamic-dns)"` in the record `comment`.
- **`internal/execext/`** — vendored (via `vendir`) from `github.com/go-task/task`. Wraps `mvdan.cc/sh/v3` so `--ipcmd` runs in a portable shell interpreter.

**Config precedence:** CLI flag → env var (`CFDDNS_*`, `-`/`.` → `_`) → YAML config file → default.

**Logging:** `logrus` structured fields. Error-handling convention: most errors are fatal via `log.WithError(err).Fatal(...)` — appropriate for a systemd-timer-driven CLI.

**Deployment artifacts:** Linux binary + systemd template service `cloudflare-dynamic-dns@.service` (reads `/etc/cloudflare-dynamic-dns/config.d/%I.yaml`) + timer; Docker image `ghcr.io/zebradil/cloudflare-dynamic-dns`; DEB/RPM/APK; AUR `cloudflare-dynamic-dns-bin`; Nix flake/package. State directory respects `$STATE_DIRECTORY` from systemd.

## Non-obvious conventions

- **`.goreleaser.yml` is generated** — edit `.goreleaser.ytt.yml` and run `task misc:build:goreleaser-config` (`ytt`). Never edit `.goreleaser.yml` by hand.
- **README usage section is generated** — the block between `<!-- BEGIN CFDDNS_USAGE -->` and `<!-- END CFDDNS_USAGE -->` is injected by `scripts/update_readme` (`task docs:update-readme`). The source text is `longDescription` in `cmd/cmd.go`.
- **`internal/execext` is vendored, not a go-modules dep** — update via `vendir sync` from repo root; commit both the synced files and `vendir.lock.yml`. The `ref:` in `vendir.yml` is pinned to a specific SHA (not `main`) because upstream `nightly` has introduced imports from `go-task`-internal packages (`github.com/go-task/task/v3/internal/env`, `github.com/go-task/task/v3/errors`) that cannot be consumed cross-module. When bumping the pin, verify `go build ./...` still succeeds before committing.
- **Conventional Commits required** — enforced by lefthook (commit-msg) and CI PR-title linter. Semantic-release uses them to determine the next version.
- **Release pipeline** — push to `main` → semantic-release → updates `CHANGELOG.md` + bumps `nix/package.nix` (`nix-update` regenerates `vendorHash`) → tags without `v` prefix (`tagFormat: '${version}'`) → goreleaser publishes binaries, GHCR images, AUR, nfpm packages.
- **Formatting** — `gofumpt` (stricter than `gofmt`) + `goimports-reviser -set-alias`. Expect aliased imports: `log "github.com/sirupsen/logrus"`, `cloudflare "github.com/cloudflare/cloudflare-go"`.
- **`cfg.yaml`** at repo root is a dev fixture — do not overwrite it with real credentials.
