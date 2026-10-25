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
gh jq --clear
```

Evaluate a jq expression against JSON input and print the results.
The expression is evaluated by the jq engine bundled with GitHub CLI, so no
external `jq` binary is required.

- `<expression>` (required unless `--clear` is given): jq expression to evaluate
- `[file...]` (optional): JSON files to read; when omitted, input is read from
  stdin

Input is taken from the first matching source: `--null-input`, `--load`,
`[file...]`, then stdin.
When stdout is a terminal, output is indented and colorized; otherwise it is
printed as compact JSON.

`--save` stores the raw input (before `--slurp` is applied) so it can be reused
later with `--load` without fetching it again.
The saved input is placed in the `gh-jq` directory under the user configuration
directory (for example, `~/.config/gh-jq` on Linux or
`~/Library/Application Support/gh-jq` on macOS).
`--save`, `--load` and `--clear` are mutually exclusive.

| Flag | Short | Default | Description |
| ---- | ----- | ------- | ----------- |
| `--clear` | `-C` | `false` | Remove the saved JSON input and exit; succeeds even if nothing is saved |
| `--load` | `-L` | `false` | Use the previously saved JSON input instead of stdin or `[file...]` |
| `--null-input` | `-n` | `false` | Use `null` as the single input value and ignore all other input (takes precedence over `--slurp`) |
| `--save` | `-S` | `false` | Save the input JSON for later use with `--load` |
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
