# linuXwin

<p align="center">
  <a href="https://github.com/khanalsaroj/linuXwin/actions/workflows/release.yml"><img src="https://github.com/khanalsaroj/linuXwin/actions/workflows/release.yml/badge.svg" /></a>
  <a href="https://github.com/khanalsaroj/linuXwin/releases"><img src="https://img.shields.io/github/v/release/khanalsaroj/linuXwin?sort=semver" /></a>
  <img src="https://img.shields.io/badge/go-1.21+-00ADD8?logo=go&logoColor=white" />
  <img src="https://img.shields.io/badge/platform-windows-0078D6?logo=windows&logoColor=white" />
  <a href="LICENSE"><img src="https://img.shields.io/github/license/khanalsaroj/linuXwin" /></a>
</p>

A Linux command compatibility layer for the native Windows `cmd.exe` prompt, written in Go.

Install it, open a normal Command Prompt window, and type `ls`, `pwd`, `cp`, `grep`, `cat`... They run as real Windows
executables backed by the Go standard library and Windows APIs — no WSL, no Cygwin, no MSYS2, no Git Bash, no Linux VM.

```text
C:\Users\you> ls -la
C:\Users\you> cp file1.txt file2.txt
C:\Users\you> grep -n "TODO" *.go
```

## Contents

- [Installation](#installation)
- [Supported commands](#supported-commands)
- [Examples](#examples)
- [Architecture](#architecture)
- [Path handling](#path-handling)
- [Environment variables](#environment-variables)
- [Windows/Linux behavior differences](#windowslinux-behavior-differences)
- [Why cd/mkdir/rmdir/echo need the DOSKEY layer](#why-cdmkdirrmdirecho-need-the-doskey-layer)
- [Limitations](#limitations)
- [Command compatibility table](#command-compatibility-table)
- [Development](#development)
- [Building](#building)
- [Testing](#testing)
- [Packaging](#packaging)
- [Releases](#releases)
- [License](#license)

## Installation

### 🪟 Windows (PowerShell) — one-liner

```powershell
iwr -useb https://raw.githubusercontent.com/khanalsaroj/linuXwin/main/main/install.ps1 | iex
```

Downloads the latest release for your architecture, verifies its checksum, and installs it to
`%LOCALAPPDATA%\Programs\LinuxCmd`, adding it to your user PATH. No admin rights, no Go toolchain required.

> Open a **new** Command Prompt window afterwards (PATH changes never apply to windows already open) and try `ls -la`.

Since piping into `iex` can't take script parameters, configure it with environment variables instead:

```powershell
$env:LINUXCMD_VERSION = "v1.2.3"                                   # pin a version instead of latest
$env:LINUXCMD_INSTALL_DIR = "D:\Tools\LinuxCmd"                    # custom install location
$env:LINUXCMD_ENABLE_DOSKEY = "1"                                  # also enable cd/mkdir/rmdir/echo bare-word support
iwr -useb https://raw.githubusercontent.com/khanalsaroj/linuXwin/main/main/install.ps1 | iex
```

See [why the DOSKEY layer is opt-in](#why-cdmkdirrmdirecho-need-the-doskey-layer).

### 📦 Prebuilt binaries

Grab the zip for your architecture (`amd64` for almost everyone) from
the [Releases](https://github.com/khanalsaroj/linuXwin/releases) page (each release ships a `checksums.txt`), extract it
anywhere, and run `installer\install.ps1` from inside the extracted folder — optionally with `-EnableDoskeyOverrides` or
`-InstallDir`.

### 🛠️ From source

Requires [Go](https://go.dev) 1.21+.

```powershell
git clone https://github.com/khanalsaroj/linuXwin.git
cd linuXwin
.\scripts\build.ps1                # builds dist\linuxcmd.exe + one launcher per command
.\installer\install.ps1            # installs to %LOCALAPPDATA%\Programs\LinuxCmd, adds it to PATH
```

### Verify

```text
linuxcmd --version
```

### Uninstalling

```powershell
.\installer\uninstall.ps1
```

Reverses everything install.ps1 did: removes the PATH entry, the optional AutoRun/DOSKEY hook, the `LINUXCMD_HOME`
variable, and the installed files. Nothing outside `HKCU` (the current user's own registry hive) is ever touched, and
cmd.exe itself is never modified. If you installed to a custom directory, pass the same `-InstallDir`.

## Supported commands

Filesystem: `ls`, `pwd`, `cd`, `mkdir`, `rmdir`, `cp`, `mv`, `rm`, `cat`, `touch`
Shell utilities: `echo`, `clear`, `whoami`, `hostname`
Search/text: `grep`, `find`
System/network: `ping`, `ip`, `ps`, `kill`

Run `linuxcmd` with no arguments (or `linuxcmd --help`) for a live list with one-line summaries, generated from the
actual command registry. `linuxcmd --version` (or `-v`) prints the installed version.

## Examples

```text
ls -la                          long listing, hidden files included
ls -1 /tmp                      one entry per line
cp -r src dst                   recursive copy
mv old.txt new.txt              rename
rm -rf build                    recursive, ignore missing
cat -n file.txt                 numbered lines
grep -in "error" *.log          case-insensitive, line numbers, glob-expanded
find . -name "*.go" -type f
echo "home is $HOME"
touch -c maybe-exists.txt       update mtime, don't create
ping -c 4 example.com
ip link
ps | findstr node.exe
```

## Architecture

```text
linuxcmd/
├── cmd/linuxcmd/          entrypoint — the ONLY compiled program
├── internal/
│   ├── command/           Command interface + registry + exit codes
│   ├── parser/             POSIX-style flag parser (combined short flags, --long)
│   ├── paths/               Linux → Windows path normalization + glob expansion
│   ├── environment/     $VAR → Windows env var mapping/expansion
│   ├── fsutil/               shared recursive copy/move/remove helpers
│   └── output/            Linux-style error text, ls -l formatting, sizes
├── commands/                one file per command, self-registers via init()
├── installer/               install.ps1 / uninstall.ps1 / common.ps1 / linuxcmd.doskey
├── scripts/build.ps1     dev build → dist/
└── tests/                    integration tests (build + run the real exe)
```

### Shared "multicall" binary

There is exactly **one** compiled Go program: `linuxcmd.exe`. Every per-command executable on PATH (`ls.exe`, `cp.exe`,
`grep.exe`, ...) is a **hardlink** to that same file — not a separate build, not a wrapper that shells out. At startup,
`cmd/linuxcmd/main.go` looks at its own invoked filename (`os.Args[0]`) to decide which command to run; when invoked as
`linuxcmd` itself, the command name is instead taken from the first argument (`linuxcmd ls -la`). This is the same
technique BusyBox uses on Linux.

Consequences:

- **Small install size**: hardlinks share the same on-disk data, so 20 "separate" executables cost the disk space of one
  (~3.7 MB total, not 20×3.7 MB).
- **Fast startup**: every command is a native Windows exe launch, no interpreter, no shim process spawning another
  process.
- **Easy to extend**: adding a command means adding one file to `commands/` that calls `command.Register(...)` in an
  `init()` function. `scripts/build.ps1` and `installer/install.ps1` discover it automatically via
  `linuxcmd --list-commands` — nothing else needs editing.

### Command flow

```text
ls -la /tmp
  ↓ argv[0] = "ls.exe"  →  command name "ls"
  ↓ internal/parser      →  {l:true, a:true}, positional=["/tmp"]
  ↓ internal/paths        →  "/tmp" resolved to a real Windows path
  ↓ os.ReadDir / os.Stat →  Windows filesystem
  ↓ internal/output       →  Linux-style "drwxr-xr-x ... Aug 17 12:30 name"
```

Commands use the Go standard library and Windows syscalls directly (`os`, `path/filepath`, `syscall` for `ps`/`kill`
/console mode, `net` for `ip`); they do **not** shell out to `cmd.exe /c <string>`. The one deliberate exception is
`ping`, which wraps Windows' own `ping.exe` by explicit `System32` path with separated `os/exec` arguments (never a
shell string) — implementing raw ICMP without it would need elevated privileges.

## Path handling

`internal/paths` normalizes Linux-style syntax into a real Windows path — it never assumes `/` is the Windows root,
since Windows has no single root shared across drives:

| You type                       | Resolves to                                                                         |
|--------------------------------|-------------------------------------------------------------------------------------|
| `~` or `~/Documents`           | `%USERPROFILE%` (or `...\Documents`)                                                |
| `/tmp`, `/tmp/foo`             | `%TEMP%` (or `...\foo`)                                                             |
| `/dev/null`                    | `NUL`                                                                               |
| `/c/Users/x`                   | `C:\Users\x` (Git-Bash-style drive shorthand)                                       |
| `C:\x`, `C:/x`                 | `C:\x` (slash direction normalized)                                                 |
| `/etc` (no Windows equivalent) | `<current drive>:\etc` — closest meaningful mapping, documented as a simplification |
| `.`, `..`, relative paths      | resolved against the real current working directory                                 |

Output paths (`pwd`, `ls`, error messages) show the **real Windows path**, not a virtual Linux-style one — this is the
one consistent, documented choice per the project's own requirement to prefer real paths by default.

### Wildcard expansion

Unlike bash, `cmd.exe` never expands `*`/`?`/`[...]` before invoking a program — each command receives the literal
string `*.txt`. `internal/paths.ExpandGlobs` is applied to file-path arguments of `rm`, `cp` (sources), `mv` (sources),
`cat`, `ls`, and `grep` (files) so that `rm *.txt` behaves the way Linux users expect. An argument that matches nothing
is passed through unchanged so normal "No such file or directory" handling still applies to genuine typos.

## Environment variables

`internal/environment` maps common Linux-style names onto their Windows equivalents (`$HOME`→`%USERPROFILE%`, `$USER`→
`%USERNAME%`, `$TMP`/`$TMPDIR`→`%TEMP%`, `$SHELL`→`%COMSPEC%`, `$HOSTNAME`→`%COMPUTERNAME%`, `$PWD`→ computed live).
Anything else falls back to a direct lookup, so existing Windows variables like `$USERNAME` or `$APPDATA` also work.
`echo` (and only `echo`, for now) expands `$VAR`/`${VAR}` references.

## Windows/Linux behavior differences

| Concept                  | Linux                                | This project's Windows mapping                                                          |
|--------------------------|--------------------------------------|-----------------------------------------------------------------------------------------|
| Filesystem root          | single `/`                           | per-drive; bare `/x` maps to the *current* drive's root                                 |
| `~`                      | `$HOME`                              | `%USERPROFILE%`                                                                         |
| `/tmp`                   | tmpfs/disk temp dir                  | `%TEMP%`                                                                                |
| `/dev/null`              | null device                          | `NUL`                                                                                   |
| Permission bits          | full POSIX rwxrwxrwx                 | approximated — see [Limitations](#limitations)                                          |
| Signals (`kill -9` etc.) | distinct signals, some catchable     | none — every `kill` is an unconditional `TerminateProcess`                              |
| Shell wildcard expansion | done by bash before the program runs | done by each command itself (`internal/paths.ExpandGlobs`), since cmd.exe never does it |
| `cd`                     | shell builtin                        | see next section                                                                        |

## Why cd/mkdir/rmdir/echo need the DOSKEY layer

`cmd.exe` has its own **builtin** commands — `cd`, `md`/`mkdir`, `rd`/`rmdir`, `echo`, `cls`, `copy`, `del`, `type`,
`move`, `ren` — and a builtin **always** wins over a same-named `.exe` on PATH, with no way to override this via PATH
ordering. This is hard-wired into `cmd.exe`'s parser.

So `cd.exe`, `mkdir.exe`, `rmdir.exe` and `echo.exe` are fully working programs — usable via `linuxcmd cd ...`, from
scripts, or by typing the extension explicitly (`cd.exe ~`) — but **typing `cd` bare at the prompt still runs cmd.exe's
own builtin**, unless you opt in:

```powershell
.\installer\install.ps1 -EnableDoskeyOverrides
```

This registers `installer\linuxcmd.doskey` via the standard, documented
`HKCU\Software\Microsoft\Command Processor\AutoRun` extension point (not a modification of cmd.exe itself — the same
mechanism people use for custom prompts). It's opt-in by default because it's a small, persistent behavior change
applied to every new interactive CMD session, and turning it on should be an explicit choice, not a side effect of
installing a tool. The installer never overwrites an existing `AutoRun` value outright — it appends its own hook
alongside whatever you already had there, and removes only that segment on uninstall.

**`cd` is special even with the DOSKEY layer enabled**, because a child process can never change its parent shell's
working directory — that's true on every OS (it's exactly why every real Unix shell implements `cd` as a builtin too,
not an external program). `cd.exe` only resolves Linux-style syntax (`~`, `/tmp`, ...) to a real Windows path and prints
it; the DOSKEY macro captures that output with `for /f` and feeds it to cmd.exe's own builtin `cd`, in the same process,
which is what actually changes the directory:

```text
cd=for /f "usebackq delims=" %i in (`"%LINUXCMD_HOME%\cd.exe" $*`) do @cd /d "%i"
```

`mkdir` and `rmdir` have no such limitation — creating/removing a directory is a real filesystem action a child process
can perform directly — so their macros are plain passthroughs to `mkdir.exe`/`rmdir.exe` for Linux-style flags (`-p`,
`-v`) and error text. Same for `echo.exe` (`$VAR` expansion, `-n`, `-e`).

Note: DOSKEY macro expansion only applies to lines typed at an **interactive** prompt — it does not affect `.bat`/`.cmd`
script files at all (batch files are read directly by cmd.exe's interpreter, bypassing the console input layer doskey
hooks into), so this layer cannot change the behavior of existing scripts, including ones that use `@echo off`.

## Limitations

- **Permission bits are approximated, not faked.** Windows has no POSIX rwxrwxrwx model. `ls -l` shows `d`/`-`/`l` for
  type, and derives a single rwx-style triplet (repeated three times, since Windows has no owner/group/other
  distinction) from the read-only attribute and whether the entry is a directory. This is documented here rather than
  hidden.
- **Hardlink counts, owner, and group are simplified.** The link count column is always `1`; owner and group both show
  the current Windows username (Windows files don't carry a POSIX-style group).
- **`cd` cannot change your shell's directory without the opt-in DOSKEY layer** (see above) — this is an OS-level
  constraint, not a bug.
- **`kill` has no signal semantics.** Every invocation maps onto `TerminateProcess`; there's no SIGTERM-vs-SIGKILL
  distinction, since Windows doesn't have one for arbitrary processes.
- **`ping` wraps the real Windows `ping.exe`** rather than reimplementing ICMP (which needs elevated privileges on
  Windows); only `-c` (count) is translated, other flags pass through as-is.
- **`ip` is read-only** and covers `ip addr`/`ip link`; there's no iproute2 on Windows to wrap, so it's a native
  implementation over Go's `net` package. Route/address modification subcommands are out of scope for this MVP.
- **`grep`'s regex dialect is RE2** (Go's `regexp`), which is close to POSIX extended but has no backreferences or
  lookaround.
- **No pipelines or redirection yet** (`cat f | grep x`, `> out.txt`) — `cmd.exe`'s own `|`/`>`/`>>` already work with
  any of these commands today since they're normal executables using real stdin/stdout, so this mostly matters for
  parity with typing habits, not missing functionality. The `Context{Stdin,Stdout,Stderr}` design in `internal/command`
  exists specifically so this can be added later without reworking commands.

## Command compatibility table

| Command    | Supported | Linux behavior                        | Windows implementation                                | Limitations                                                      |
|------------|-----------|---------------------------------------|-------------------------------------------------------|------------------------------------------------------------------|
| `ls`       | Yes       | list directory, `-l -a -h -1`         | `os.ReadDir` + custom formatter                       | permission bits approximated                                     |
| `pwd`      | Yes       | print cwd                             | `os.Getwd`                                            | prints real Windows path, not virtual `/home/...`                |
| `cd`       | Partial   | change shell directory                | resolves path, prints it                              | can't change parent shell's cwd without `-EnableDoskeyOverrides` |
| `mkdir`    | Yes       | create dirs, `-p -v`                  | `os.Mkdir`/`MkdirAll`                                 | shadowed by cmd builtin unless DOSKEY layer enabled              |
| `rmdir`    | Yes       | remove empty dirs, `-p -v`            | `os.Remove` (+ parent walk for `-p`)                  | same shadowing note                                              |
| `cp`       | Yes       | copy, `-r -f -n -v`                   | `io.Copy` based, recursive walk                       | no `-p` (preserve attrs) yet                                     |
| `mv`       | Yes       | move/rename, `-f -n -v`               | `os.Rename`, falls back to copy+delete across volumes | —                                                                |
| `rm`       | Yes       | remove, `-r -f -v`                    | `os.Remove`/`RemoveAll`                               | —                                                                |
| `cat`      | Yes       | concatenate/print, `-n`               | buffered I/O                                          | no `-A`/`-E` etc.                                                |
| `touch`    | Yes       | create/update mtime, `-c`             | `os.Create`/`os.Chtimes`                              | no `-t TIMESTAMP`                                                |
| `echo`     | Yes       | print args, `-n -e`, `$VAR` expansion | —                                                     | shadowed by cmd builtin unless DOSKEY layer enabled              |
| `clear`    | Yes       | clear screen                          | ANSI escape + enables VT mode                         | needs a VT-capable console (default on Win10 1511+)              |
| `whoami`   | Yes       | print username                        | `os/user`, domain prefix stripped                     | —                                                                |
| `hostname` | Yes       | print hostname                        | `os.Hostname`                                         | —                                                                |
| `grep`     | Yes       | pattern search, `-i -n -v -r -l -c`   | Go `regexp` (RE2)                                     | no backreferences/lookaround                                     |
| `find`     | Partial   | search tree                           | `filepath.Walk` + `-name -iname -type`                | small predicate subset, no `-exec`                               |
| `ping`     | Yes       | ICMP echo                             | wraps real `ping.exe`, `-c`→`-n`                      | most flags pass through untranslated                             |
| `ip`       | Partial   | show/manage networking                | native via Go `net` package, `addr`/`link`            | read-only, no `iproute2` route/tunnel features                   |
| `ps`       | Yes       | list processes                        | `syscall.CreateToolhelp32Snapshot`                    | columns limited to PID/PPID/CMD                                  |
| `kill`     | Yes       | terminate by PID                      | `syscall.TerminateProcess`                            | no signal distinction                                            |

## Development

- Go 1.21+, Windows (this project targets `GOOS=windows` throughout — several commands use the `syscall` package's
  Windows-only APIs).
- Add a command: create `commands/<name>.go`, implement `command.Command` (`Name() string`, `Summary() string`,
  `Run(ctx *command.Context) int`), register it in an `init()` function. It's picked up automatically everywhere (help
  text, build script, installer) — no other files need editing.
- Keep path handling going through `internal/paths`, error messages through `internal/output`, and flag parsing through
  `internal/parser` rather than reimplementing ad hoc versions per command.

## Building

```powershell
.\scripts\build.ps1                              # → dist\linuxcmd.exe + one launcher per command
.\scripts\build.ps1 -OutDir "C:\some\other\dir"
```

Or directly with Go, for just the shared engine:

```powershell
go build -o dist\linuxcmd.exe .\cmd\linuxcmd
```

## Testing

```powershell
go vet ./...
go test ./...          # unit tests (internal/*, commands/*) + integration tests (tests/)
go test ./... -v       # verbose
```

- `internal/paths`, `internal/parser`, `internal/output`: package-level unit tests.
- `commands/*_test.go`: exercise each command through the real `command.Lookup` registry with in-memory stdout/stderr,
  covering normal operation, missing files, bad arguments, spaces/Unicode filenames, relative/absolute paths, `.`/`..`,
  empty directories, and exit codes.
- `tests/integration_test.go`: builds the actual `linuxcmd.exe` and runs it as a real subprocess (both via
  `linuxcmd <cmd>` and via a hardlinked per-command launcher, mirroring how `cmd.exe` actually invokes it), covering
  exit-code propagation and command-not-found handling.

## Packaging

For a distributable release (so end users don't need Go installed):

1. `.\scripts\build.ps1 -Version v1.2.3` to produce `dist\` with the version embedded (`linuxcmd --version` reflects
   it).
2. Ship the repo's `installer\` directory alongside `dist\`, preserving the relative layout (`install.ps1` looks for
   `..\dist\linuxcmd.exe` next to itself).
3. End users run `installer\install.ps1` (optionally with `-EnableDoskeyOverrides` or `-InstallDir`) — no admin rights
   or Go installation required.

`.github/workflows/release.yml` automates exactly this: every push to `main` that passes `go vet`/`build`/`test` bumps a
semantic version from the commit message prefix (`feat:` → minor, `fix:` → patch, `BREAKING CHANGE:`/`!:` → major),
cross-compiles `windows/amd64`, `windows/arm64` and `windows/386` with the version embedded via `-ldflags`, packages
each into a zip with the matching `dist/` + `installer/` layout above, tags the commit, and publishes a GitHub release
with generated notes and checksums.

## Releases

Prebuilt zips for each Windows
architecture: [github.com/khanalsaroj/linuXwin/releases](https://github.com/khanalsaroj/linuXwin/releases). Each release
includes `checksums.txt` (SHA-256) for the archives.

## License

[MIT](LICENSE) © Saroj Khanal
