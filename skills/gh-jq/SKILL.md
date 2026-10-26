---
name: gh-jq
description: GitHub CLI extension (gh jq) that evaluates jq expressions against JSON using the jq engine bundled with GitHub CLI, so no external jq binary is required — filtering `gh api` and other gh JSON output, reading JSON from stdin or files, slurping multiple JSON values into an array, evaluating expressions with null input, and saving JSON responses to named channels to query them repeatedly with --save / --load / --clear / --list.
---

# gh-jq

Reference for gh-jq — a GitHub CLI extension that processes JSON with jq
expressions. It uses the jq engine bundled with GitHub CLI (gojq), so it works
anywhere `gh` works without installing the `jq` binary.

Version: 0.1.0

## Prerequisites

### Installation

```bash
gh extension install srz-zumix/gh-jq
```

### Authentication

gh-jq itself does not call the GitHub API, so no authentication is needed.
Commands that produce the input (for example `gh api`) need `gh auth login` as
usual.

## CLI Structure

```text
gh jq <expression> [file...]   # Root command: evaluate a jq expression
├── completion                  # Generate a shell completion script
└── skills                      # Manage the bundled agent skills
```

## Input Selection

Input is taken from the first matching source:

1. `--null-input` — a single `null` value; all other input is ignored
2. `--load` — the JSON saved by a previous `--save`
3. `[file...]` — the given files, concatenated in order
4. stdin

The input must be **a single JSON value** unless `--slurp` is given. Multiple
values (for example several files, newline-delimited JSON, or
`gh api --paginate` output) fail with `invalid character '{' after top-level
value`; pass `--slurp` to read them into one array.

## Output Format

- When stdout is a terminal, JSON results are indented with two spaces and
  colorized. Otherwise (pipes, redirects, `$(...)`) they are compact, one result
  per line.
- String results are printed raw, without quotes (like `jq -r`).
- `null` results are printed as an empty line.
- `@csv`, `@tsv`, `@base64`, `@json` and other jq format strings are available
  in the expression.

## Commands

### jq (root)

Evaluate a jq expression against JSON input and print the results.

```bash
gh jq <expression> [file...] [--slurp] [--null-input] [--save[=<channel>] | --load[=<channel>]]
gh jq --clear[=<channel>]
gh jq --list
```

- `<expression>` (required unless `--clear` or `--list` is given): jq expression to evaluate
- `[file...]` (optional): JSON files to read; when omitted, stdin is read

| Option | Default | Description |
| --- | --- | --- |
| `-C`, `--clear[=<channel>]` | (none) | Remove the saved JSON input of `<channel>` and exit. Without a channel name, remove all channels. Succeeds even if nothing is saved |
| `--list` | `false` | List the names of channels that have saved JSON input and exit |
| `-L`, `--load[=<channel>]` | (none) | Use the saved JSON input of `<channel>` instead of stdin or `[file...]`. Without a channel name, use `default` |
| `-n`, `--null-input` | `false` | Use `null` as the single input value and ignore all other input (takes precedence over `--slurp` and `--load`) |
| `-S`, `--save[=<channel>]` | (none) | Save the input JSON to `<channel>` for later use with `--load`. Without a channel name, use `default` |
| `-s`, `--slurp` | `false` | Read all input values into a single array |
| `-v`, `--version` | - | Print the version |

`--save`, `--load`, `--clear` and `--list` are mutually exclusive.
`--save` cannot be combined with `--null-input`.

Channel names are values, so write `--save=<channel>` (or `-S=<channel>`), not
`--save <channel>`. A channel name may contain only lowercase letters, digits,
`-` and `_`.

`--save` stores the raw input **before** `--slurp` is applied. When the saved
input contains multiple JSON values, pass `--slurp` again with `--load`.
Each channel is stored in the `gh-jq` directory under the user configuration
directory (for example `~/.config/gh-jq` on Linux or
`~/Library/Application Support/gh-jq` on macOS). The `default` channel is
`last-input.json`; other channels are in the `channels` subdirectory. Each
`--save` to a channel overwrites that channel.

### completion

Generate a shell completion script. When invoked as `gh jq`, a patch for gh
completion is generated; when invoked as `gh-jq`, a standard completion script
is generated.

```bash
gh jq completion --shell bash
```

| Option | Default | Description |
| --- | --- | --- |
| `-s`, `--shell` | (required) | Shell type: `bash`, `zsh`, `fish` or `powershell` |

See the
[Shell Completion Guide](https://github.com/srz-zumix/go-gh-extension/blob/main/docs/shell-completion.md)
for setup instructions.

### skills

Install and manage the agent skills bundled with gh-jq.

```bash
gh jq skills list
gh jq skills install
gh jq skills install -scope repo -dry-run
gh jq skills status
```

Subcommands: `list`, `install`, `update`, `reinstall`, `uninstall`, `status`.

| Option | Default | Description |
| --- | --- | --- |
| `-dry-run` | `false` | Print what would happen without making changes |
| `-force` | `false` | Overwrite unmanaged skills |
| `-prefix` | - | Skill installation directory (overrides `-scope`) |
| `-scope` | `user` | `user` (`~/.agents/skills`) or `repo` (`<repo-root>/.agents/skills`) |

## Common Workflows

### Filter GitHub CLI output

```bash
gh api repos/cli/cli | gh jq '.full_name'
gh pr list --json number,title,author \
  | gh jq '.[] | select(.author.login == "octocat") | .number'
```

Prefer `gh api --jq` / `gh pr list --json ... --jq` when a single filter is
enough. Use gh-jq when the same response is queried multiple times, when the
input comes from files, or when `--slurp` / `--null-input` is needed.

### Query a large response repeatedly without refetching

```bash
# Fetch once and save to a channel
gh api --paginate repos/my-org/my-repo/issues | gh jq --save=issues --slurp 'add | length'

# Explore the saved data
# (--slurp is needed again because --paginate emits one array per page)
gh jq --load=issues --slurp 'add | map(.title)'
gh jq --load=issues --slurp \
  'add | group_by(.user.login) | map({user: .[0].user.login, count: length})'

# List channels, then remove the data when done
gh jq --list
gh jq --clear=issues
```

### Combine multiple JSON files or values

```bash
gh jq --slurp 'map(.name)' a.json b.json
printf '{"a":1}\n{"a":2}\n' | gh jq --slurp 'map(.a) | add'
```

### Produce plain text for shell scripts

```bash
# Raw strings, one per line
gh api repos/cli/cli/releases | gh jq '.[].tag_name'

# Tab-separated rows
gh api repos/cli/cli/releases | gh jq '.[] | [.tag_name, .published_at] | @tsv'
```

### Evaluate an expression without input

```bash
gh jq --null-input '{a: 1, b: 2}'
gh jq -n '[range(5)] | map(. * 2)'
```

## Differences from jq

- Only the options listed above are supported. jq options such as `--arg`,
  `--argjson`, `-r`, `-c`, `-e`, `--tab` and `--raw-input` are rejected with
  `unknown flag`. Embed values directly in the expression instead, and note that
  strings are already printed raw.
- Indentation and color cannot be controlled; they depend only on whether
  stdout is a terminal.
- Expressions are evaluated by gojq, so a few jq features (for example
  `input_line_number`, `$__loc__`, or some date/regex edge cases) may behave
  differently.

## Troubleshooting

| Symptom | Cause / Resolution |
| --- | --- |
| `invalid character '{' after top-level value` (or `'['`) | The input contains more than one JSON value (several files, NDJSON, `gh api --paginate`). Add `--slurp`. |
| `no saved input found for channel` | `--load` was used for a channel that was never saved, or it was removed with `--clear`. Run the producing command again with `--save=<channel>`. |
| `invalid channel name` | Channel names may contain only lowercase letters, digits, `-` and `_`. |
| `unknown flag: --arg` / `unknown shorthand flag: 'r'` | gh-jq does not support these jq options. Embed values in the expression; strings are already printed raw. |
| `failed to parse jq expression` | The expression has a syntax error. Quote the whole expression with single quotes so the shell does not alter it. |
| `if any flags in the group [save load clear list] are set none of the others can be` | `--save`, `--load`, `--clear` and `--list` cannot be combined. Run them as separate commands. |
| `if any flags in the group [save null-input] are set none of the others can be` | `--save` cannot be combined with `--null-input`, because there is no input to save. |
| `cannot open file` | A `[file...]` argument does not exist or is not readable. |
| `@csv cannot be applied to: number` | `@csv` / `@tsv` need an array; wrap the values, for example `[.a, .b] \| @csv`. |
| Output has no colors or indentation | stdout is not a terminal (piped or redirected). This is expected. |
