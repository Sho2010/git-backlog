# git-backlog

git subcommand for [Backlog](https://backlog.com/) ticket-driven development.
Resolves the Backlog issue for the current topic branch and exposes it via
short subcommands.

## Install

```sh
go install github.com/Sho2010/git-backlog@latest
```

Ensure `$GOPATH/bin` (or `$GOBIN`) is on `$PATH` so that `git backlog` resolves
as a git subcommand.

## Settings

Precedence: **env var > `git config --local` > `git config --global` > built-in default**.

| Git config key | Env var override | Default | Description |
|---|---|---|---|
| `backlog.baseUrl` | `BACKLOG_BASE_URL` | _(required)_ | Backlog space URL, e.g. `https://your-space.backlog.jp` |
| `backlog.apiKeyEnv` | `BACKLOG_API_KEY_ENV` | `BACKLOG_API_KEY` | Name of the env var holding the PAT |
| `backlog.issuePattern` | `BACKLOG_ISSUE_PATTERN` | `[A-Z][A-Z0-9_]*-[0-9]+` | Regex applied to branch names to extract the issue key. First match wins. |
| `backlog.cacheTTL` | _(none)_ | `24h` | Cache lifetime in Go duration format (`1m`, `2h`, …) |

The PAT itself is read from the env var named by `backlog.apiKeyEnv` and is
never stored in git config. Cache files live under
`$XDG_CACHE_HOME/git-backlog/` (defaults to `~/.cache/git-backlog/`).

### Example

```sh
git config --global backlog.baseUrl https://your-space.backlog.jp
export BACKLOG_API_KEY=xxxxxxxx
```

## Commands

| Command | Description |
|---|---|
| `git backlog` / `git backlog current` | Print the current branch's issue title |
| `git backlog list [--no-fetch] [--concurrency N]` | List local branches with their issue titles (auto-fetches missing entries by default) |
| `git backlog open` | Open the current branch's issue in the browser |
| `git backlog sync [--force] [--concurrency N]` | Fetch all branches' issues into cache |

Branches without a matching issue key are skipped silently (exit 0).

`--format` accepts a Go template for machine-readable output:

```sh
git backlog --format='{{.IssueKey}}: {{.Summary}}'
git backlog list --format='{{.Branch}} {{.IssueKey}} {{.Summary}}' | fzf
```

Available template fields:

- `current`: `IssueKey`, `Summary`, `ID`, `Description`
- `list`: `Branch`, `IssueKey`, `Summary`, `Resolved`

## Development

```sh
task            # list all tasks
task check      # vet + test
task sandbox    # (re)create tmp/sandbox-repo for manual testing
task sandbox:run -- list   # run the built binary inside the sandbox
```
