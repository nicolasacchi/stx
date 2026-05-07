# stx

Read-and-write CLI for the [Stape API](https://api.app.eu.stape.io/api/doc). Go, single binary, JSON output, gjson `--jq` filters, multi-region (EU / global), multi-project (`~/.config/stx/config.toml`), multi-workspace (`X-WORKSPACE` header).

**Status (v0.1)**: vertical slice — `containers list`, `containers get`, `config {add,list,use,remove,current,doctor}`. Full 83-endpoint coverage in progress.

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
stx containers list --jq '#.identifier'        # all container IDs
stx containers get <id>
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

### `config`
- `stx config add <name> --api-key ... [--region eu|global] [--workspace UUID] [--default-container ID]`
- `stx config list` — show configured projects (`*` = default)
- `stx config use <name>` — set default project
- `stx config remove <name>`
- `stx config current` — show resolved credentials (key redacted)
- `stx config doctor` — verify auth + region + workspace by calling `GET /api/v2/users`

## Exit Codes

| Code | Meaning |
|---|---|
| 0 | success |
| 1 | generic error |
| 2 | unauthorized (401/403) — wrong key, wrong region, wrong workspace |
| 4 | not found (404) |

## Roadmap

- [x] M1 vertical: `containers list/get`, full `config` group
- [ ] M1 rest: `analytics`, `monitoring logs`, `--timeout` flag, table output
- [ ] M2: `containers create/update/delete`, `domains list/get/delete`, `power-ups get`, `containers transfer`
- [ ] M3: write paths (custom-loader, domains create/validate/revalidate, schedules, proxy-files, analytics enable)
- [ ] M4: 20 typed power-ups subcommands, monitoring rules + emails CRUD
- [ ] M5: users (incl. CSV exception), api-keys, partner-checker, resources
- [ ] M6: `stx overview` parallel-fetch dashboard, polish, goreleaser

## License

MIT
