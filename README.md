# stx

Read-and-write CLI for the [Stape API](https://api.app.eu.stape.io/api/doc). Go, single binary, JSON output, gjson `--jq` filters, multi-region (EU / global), multi-project (`~/.config/stx/config.toml`), multi-workspace (`X-WORKSPACE` header).

**Status (v0.4)**: M1 + M2 complete — read-only paths (`containers list/get`, `analytics info/browsers/clients`, `monitoring logs aggregated/detailed/trace`, `domains list/get`, `power-ups get`) + cleanup tooling (`containers create/update/delete/transfer`, `domains delete`). Destructive commands gated by `--yes`. Table renderer for `monitoring logs detailed` (replaces manual log.csv panel download). Full 83-endpoint coverage in progress.

## Install

```bash
export GOPRIVATE=github.com/nicolasacchi/*
go install github.com/nicolasacchi/stx/cmd/stx@latest
```

Or build from source:

```bash
git clone https://github.com/nicolasacchi/stx.git
cd stx
make install   # installs bin/stx into $GOBIN
```

## Quick Start

```bash
# 1. Generate an API key in Stape panel: Account settings -> API Keys
# 2. Configure stx
stx config add production --api-key 'stp_live_...' --region eu

# 3. Verify auth + region + workspace
stx config doctor

# 4. Use it
stx containers list
stx containers list --jq 'body.items.#.identifier'   # all container IDs

# Replaces manual log.csv panel download — last 5 minutes of outgoing traffic
stx monitoring logs detailed <container> --since 5m

# Filter by GTM event
stx monitoring logs detailed <container> --since 1h --event-type Purchase

# Trace one request end-to-end
stx monitoring logs trace <container> --trace-id <uuid>
```

## Authentication

Resolution order, **per field independently** (so `STAPE_API_KEY` env can pair with `region` from config):

1. CLI flag (`--api-key`, `--region`, `--workspace`)
2. Env var (`STAPE_API_KEY`, `STAPE_REGION`, `STAPE_WORKSPACE`)
3. Project block from `--project <name>` flag → from config file
4. `default_project` block in config file

### Multi-project config

`~/.config/stx/config.toml`:

```toml
default_project = "production"

[projects.production]
api_key   = "stp_live_..."
region    = "eu"                          # or "global"
workspace = "550e8400-e29b-..."           # optional, required for multi-workspace accounts
default_container = "abc123"              # optional, skips positional arg

[projects.staging]
api_key   = "stp_test_..."
region    = "eu"
```

## Global Flags

| Flag | Default | Description |
|---|---|---|
| `--api-key` | env / config | Stape API key |
| `--region` | `eu` | `eu` or `global` |
| `--workspace` | env / config / none | Workspace UUID, sent as `X-WORKSPACE` |
| `--project` | `default_project` | Named project block |
| `--json` | (auto) | Force JSON output (default on pipe) |
| `--jq <expr>` | none | gjson filter (NOT real jq) |
| `--limit <n>` | API default | Cap result count |
| `--timeout <dur>` | `30s` | HTTP timeout (e.g. `60s`, `2m`) |
| `--yes` | off | Skip confirmation prompt for destructive commands |
| `--verbose / -v` | off | Log HTTP requests to stderr |

## Output

JSON by default. Output is pretty-printed on a TTY, single-line on a pipe. Examples:

```bash
stx containers list                          # JSON to stdout
stx containers list --jq '#.identifier'      # all IDs as JSON array
stx containers list --jq '#.{id:identifier,name:name}'   # project per element
```

`--jq` uses **gjson** syntax, not real jq:
- Array all: `#.field`
- First: `0.field`
- Object project: `#.{a:a,b:b}`

## Commands

### `containers`
- `stx containers list` — list all containers in the workspace
- `stx containers get <id>` — single container
- `stx containers create --from-file <path> --yes` — create from a JSON body (ContainerCreateForm: name, code≥64chars, codeSettings, anonymizeOptions, zone)
- `stx containers update <id> {--from-file <path> | --name <new>} --yes` — full-document PUT (or GET-then-PUT mutate-name)
- `stx containers delete <id> --yes [--reason-setup ga,other --reason-cancel use,price]` — tries empty body first; populate reasons if Stape rejects with 400
- `stx containers transfer <id> --email <new-owner> --yes` — transfer ownership

### `domains`
- `stx domains list <container>` — list domains
- `stx domains get <container> <domain-uuid>` — note: takes the domain's UUID identifier (from `domains list .identifier`), NOT the hostname
- `stx domains delete <container> <domain-uuid> --yes`

### `analytics`
Subscription usage breakdowns. Requires the analytics module enabled on the container (HTTP 409 if disabled).
- `stx analytics info <id>` — current analytics state
- `stx analytics browsers <id> [--since 7d | --from --to]` — browser breakdown
- `stx analytics clients <id> [--since 7d | --from --to]` — client breakdown

### `power-ups`
Read-only this slice. The 20 typed PATCH subcommands (toggle / configure) land in M4.
- `stx power-ups get <container> <type>` — type names match the spec (kebab-case): `ad-blocker`, `anonymizer`, `cookie-keeper`, `custom-loader`, `dedicated-ip`, `enricher`, …. CLI alias: `header-config` → `preview-header-config`.
  - Note: returns 404 for power-ups that have no individual configuration record, even when the container's summary `powerUps.<name>` boolean is `true`. Use `containers get <id> --jq 'body.powerUps'` for the boolean summary.

### `monitoring logs`
Outgoing request logs — replaces the manual `log.csv` panel download.
- `stx monitoring logs aggregated <id> [--since 1h | --from --to] [--platform] [--event-type]`
- `stx monitoring logs detailed <id> [--since 5m | --from --to] [--platform] [--event-type]` — TTY renders a table; pipe emits JSON
- `stx monitoring logs trace <id> --trace-id <uuid> [--date now]` — single trace by ID (date defaults to now)

### `config`
- `stx config add <name> --api-key ... [--region eu|global] [--workspace UUID] [--default-container ID]`
- `stx config list` — show configured projects (`*` = default)
- `stx config use <name>` — set default project
- `stx config remove <name>`
- `stx config current` — show resolved credentials (key redacted)
- `stx config doctor` — verify auth + region + workspace by calling `GET /api/v2/containers`

## Exit Codes

| Code | Meaning |
|---|---|
| 0 | success |
| 1 | generic error |
| 2 | unauthorized (401/403) — wrong key, wrong region, wrong workspace |
| 4 | not found (404) |

## Roadmap

- [x] M1: `containers list/get`, `config` group, `analytics info/browsers/clients`, `monitoring logs aggregated/detailed/trace`, `--timeout` flag, table output for `monitoring logs detailed`
- [x] M2: `containers create/update/delete/transfer`, `domains list/get/delete`, `power-ups get`, `--yes` confirmation gate, `--from-file` body loader
- [ ] M3: `custom-loader generate`, `domains create/update/validate/revalidate`, `analytics enable`, `proxy-files list/update`, `schedules list/update`
- [ ] M3: write paths (custom-loader, domains create/validate/revalidate, schedules, proxy-files, analytics enable)
- [ ] M4: 20 typed power-ups subcommands, monitoring rules + emails CRUD
- [ ] M5: users (incl. CSV exception), api-keys, partner-checker, resources
- [ ] M6: `stx overview` parallel-fetch dashboard, polish, goreleaser

## License

MIT
