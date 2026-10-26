# gh-jq

A jq wrapper for GitHub CLI that processes JSON using jq expressions.

## Installation

```sh
gh extension install srz-zumix/gh-jq
```

## Shell Completion

**Workaround Available!** While gh CLI doesn't natively support extension
completion, we provide a patch script that enables it.

**Prerequisites:** Before setting up gh-jq completion, ensure gh CLI completion
is configured for your shell.
See [gh completion documentation][gh-completion] for setup instructions.

For detailed installation instructions and setup for each shell,
see the [Shell Completion Guide][shell-completion-guide].

[gh-completion]: https://cli.github.com/manual/gh_completion
[shell-completion-guide]: https://github.com/srz-zumix/go-gh-extension/blob/main/docs/shell-completion.md

## Commands

### jq

#### Evaluate a jq expression

```sh
gh jq <expression> [file...] [flags]
gh jq --clear[=<channel>]
gh jq --list
```

Evaluate a jq expression against JSON input and print the results.
The expression is evaluated by the jq engine bundled with GitHub CLI, so no
external `jq` binary is required.

- `<expression>` (required unless `--clear` or `--list` is given): jq expression to evaluate
- `[file...]` (optional): JSON files to read; when omitted, input is read from
  stdin

Input is taken from the first matching source: `--null-input`, `--load`,
`[file...]`, then stdin.
When stdout is a terminal, output is indented and colorized; otherwise it is
printed as compact JSON.

`--save` stores the raw input (before `--slurp` is applied) so it can be reused
later with `--load` without fetching it again.
Saved input is organized into channels. A channel is a name you choose, such as
`work` or `issues`, so that several inputs can be kept side by side.
Omitting the channel name uses the `default` channel.
Channel names may contain only lowercase letters, digits, `-` and `_`.
Saved input is placed in the `gh-jq` directory under the user configuration
directory (for example, `~/.config/gh-jq` on Linux or
`~/Library/Application Support/gh-jq` on macOS). The `default` channel is stored
as `last-input.json`; other channels are stored in the `channels` subdirectory.

`--save`, `--load`, `--clear` and `--list` are mutually exclusive.
`--save` cannot be combined with `--null-input`.

Channel names are passed as values, so use `--save=<channel>` (or `-S=<channel>`)
rather than `--save <channel>`. A bare `--save` or `--load` uses `default`.

| Flag | Short | Default | Description |
| ---- | ----- | ------- | ----------- |
| `--clear[=<channel>]` | `-C` | | Remove the saved JSON input of `<channel>` and exit. Without a channel name, remove all channels. Succeeds even if nothing is saved |
| `--list` | | `false` | List the names of channels that have saved JSON input and exit |
| `--load[=<channel>]` | `-L` | | Use the saved JSON input of `<channel>` instead of stdin or `[file...]`. Without a channel name, use `default` |
| `--null-input` | `-n` | `false` | Use `null` as the single input value and ignore all other input (takes precedence over `--slurp`) |
| `--save[=<channel>]` | `-S` | | Save the input JSON to `<channel>` for later use with `--load`. Without a channel name, use `default` |
| `--slurp` | `-s` | `false` | Read all input values into a single array |
| `--version` | `-v` | | Print the version |

Examples:

```sh
# Extract a field from GitHub CLI output
gh api repos/cli/cli | gh jq '.full_name'

# Read from files
gh jq '.name' package.json

# Save the response once, then query it repeatedly
gh api repos/cli/cli/releases | gh jq --save '.[0].tag_name'
gh jq --load '.[].name'

# Keep separate channels for different responses
gh api repos/cli/cli/releases | gh jq --save=releases '.[0].tag_name'
gh api repos/cli/cli/issues | gh jq --save=issues 'length'
gh jq --load=releases '.[].name'
gh jq --list
gh jq --clear=issues
gh jq --clear

# Combine multiple JSON values into an array
gh jq --slurp 'map(.name)' a.json b.json

# Evaluate an expression without input
gh jq --null-input '{a: 1, b: 2}'
```

---

### completion

#### Generate a shell completion script

```sh
gh jq completion --shell <shell>
```

Generate a shell completion script for gh-jq.
When invoked as `gh jq`, a patch for gh completion is generated; when invoked as
`gh-jq`, a standard completion script is generated.
See [Shell Completion](#shell-completion) for setup instructions.

| Flag | Short | Default | Description |
| ---- | ----- | ------- | ----------- |
| `--shell` | `-s` | (required) | Shell type: `bash`, `zsh`, `fish` or `powershell` |

---

### skills

#### Manage the bundled agent skills

```sh
gh jq skills list
gh jq skills install
gh jq skills install -scope repo -dry-run
gh jq skills status
```

Install and manage the agent skills bundled with gh-jq.
Subcommands: `list`, `install`, `update`, `reinstall`, `uninstall`, `status`.

| Flag | Default | Description |
| ---- | ------- | ----------- |
| `-dry-run` | `false` | Print what would happen without making changes |
| `-force` | `false` | Overwrite unmanaged skills |
| `-prefix` | - | Skill installation directory (overrides `-scope`) |
| `-scope` | `user` | `user` (`~/.agents/skills`) or `repo` (`<repo-root>/.agents/skills`) |
