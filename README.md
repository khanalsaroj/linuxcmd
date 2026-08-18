# linuxcmd

<p align="center">
  <a href="https://github.com/khanalsaroj/linuxcmd/actions/workflows/release.yml"><img src="https://github.com/khanalsaroj/linuxcmd/actions/workflows/release.yml/badge.svg" /></a>
  <a href="https://github.com/khanalsaroj/linuxcmd/releases"><img src="https://img.shields.io/github/v/release/khanalsaroj/linuxcmd?sort=semver" /></a>
  <img src="https://img.shields.io/badge/go-1.21+-00ADD8?logo=go&logoColor=white" />
  <img src="https://img.shields.io/badge/platform-windows-0078D6?logo=windows&logoColor=white" />
  <a href="LICENSE"><img src="https://img.shields.io/github/license/khanalsaroj/linuxcmd" /></a>
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
iwr -useb https://raw.githubusercontent.com/khanalsaroj/linuxcmd/main/main/install.ps1 | iex
```

Downloads the latest release for your architecture, verifies its checksum, and installs it to
`%LOCALAPPDATA%\Programs\LinuxCmd`, adding it to your user PATH. No admin rights, no Go toolchain required.

> Open a **new** Command Prompt window afterwards (PATH changes never apply to windows already open) and try `ls -la`.

Since piping into `iex` can't take script parameters, configure it with environment variables instead:

```powershell
$env:LINUXCMD_VERSION = "v1.2.3"                                   # pin a version instead of latest
$env:LINUXCMD_INSTALL_DIR = "D:\Tools\LinuxCmd"                    # custom install location
$env:LINUXCMD_ENABLE_DOSKEY = "1"                                  # also enable cd/mkdir/rmdir/echo bare-word support
iwr -useb https://raw.githubusercontent.com/khanalsaroj/linuxcmd/main/main/install.ps1 | iex
```

See [why the DOSKEY layer is opt-in](#why-cdmkdirrmdirecho-need-the-doskey-layer).

### 📦 Prebuilt binaries

Grab the zip for your architecture (`amd64` for almost everyone) from
the [Releases](https://github.com/khanalsaroj/linuxcmd/releases) page (each release ships a `checksums.txt`), extract it
anywhere, and run `installer\install.ps1` from inside the extracted folder — optionally with `-EnableDoskeyOverrides` or
`-InstallDir`.

### 🛠️ From source

Requires [Go](https://go.dev) 1.21+.

```powershell
git clone https://github.com/khanalsaroj/linuxcmd.git
cd linuxcmd
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

193 commands. `linuxcmd --list-commands` prints the authoritative list; the groupings below are just a map.

- **Files and directories** — `ls` `cp` `mv` `rm` `mkdir` `rmdir` `ln` `find` `tree` `stat` `du` `df` `touch` `install`
  `rsync` `shred` `truncate` `readlink` `realpath` `namei`
- **Text processing** — `grep` `sed` `awk` `cut` `paste` `sort` `uniq` `tr` `head` `tail` `wc` `diff` `comm` `join`
  `fold` `fmt` `column` `nl` `rev` `dos2unix` `unix2dos`
- **Archives, encoding, digests** — `tar` `zip` `unzip` `gzip` `bzip2` `base64` `xxd` `od` `hexdump` `md5sum` `sha1sum`
  `sha256sum` `cksum`
- **System and processes** — `ps` `top` `pstree` `kill` `pgrep` `pkill` `free` `uptime` `uname` `lscpu` `lsmem` `lsblk`
  `blkid` `vmstat` `iostat` `mount` `umount` `getfacl` `sudo` `su` `systemctl` `journalctl` `crontab`
- **Networking** — `ping` `curl` `wget` `ip` `ifconfig` `ss` `netstat` `lsof` `dig` `host` `nslookup` `nc` `traceroute`
  `arp` `route`
- **Desktop integration** — `xclip` `xsel` `xdg-open` `open`
- **Wrappers for installed tools** — `ssh` `scp` `sftp` `git` `make` `cmake` `gcc` `g++` `python3` `perl` `ruby` `node`
  `npm` `jq` `rg` `fd` `vim` `nano` `tmux` `screen` `tcpdump` `nmap`

Run `linuxcmd` with no arguments (or `linuxcmd --help`) for the live list with one-line summaries, generated from the
actual command registry. `man COMMAND` gives the details for any one of them, including an **ON WINDOWS** section
recording where that command's behavior diverges from Linux and why; `man -k KEYWORD` searches names and summaries.
`linuxcmd --version` (or `-v`) prints the installed version.

Per-command limitations are in the [command compatibility table](#command-compatibility-table).

### Native commands vs. wrappers

Almost every command is a native Go implementation built directly on Win32 APIs, so it works on a stock Windows machine
with nothing else installed. A small, explicitly listed set are **wrappers**: they locate a separately installed
Windows build of the real tool and hand off to it unchanged.

Those tools — `ssh`, `make`, `gcc`, `python3`, `node`, `jq`, `vim` and the rest of the group above — already ship
first-class Windows builds, and shipping a half-working clone of any of them would be worse than useless. What linuxcmd
adds is:

- the Linux name, so `python3` works where only `python.exe` exists;
- resolution of the Windows-specific filename (`mingw32-make.exe`, `npm.cmd`, `py.exe -3`);
- lookup in install directories that are routinely missing from `PATH` (`System32\OpenSSH`, `Program Files\CMake\bin`,
  MSYS2 and Git for Windows' `usr\bin`);
- an actionable message naming what to install when the tool is absent, instead of
  `'ssh' is not recognized as an internal or external command`.

Arguments, stdin/stdout/stderr and the exit code all pass through untouched, so interactive programs like `vim` and
`ssh` get the real console.

One consequence of the multicall design matters here. The installer creates one hardlink per command, so after
installing there is an `ssh.exe` in `LINUXCMD_HOME` that *is* linuxcmd. If that directory precedes `System32` on
`PATH`, a naive lookup would resolve straight back to this process and fork forever. The lookup therefore rejects any
candidate that is the same file as the running executable, compared by file identity rather than by directory — a
directory check alone would miss a hardlink placed elsewhere on `PATH`.

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

- **`sudo` starts a new console unless Windows' own `sudo` is installed.** Elevation crosses a security boundary, so an
  elevated process cannot inherit the current console's handles; Windows gives it a fresh console instead. That means
  `sudo ls > out.txt` leaves `out.txt` empty. Windows 11 ships a native `sudo.exe` that solves this properly with an
  inline mode, and when it is present linuxcmd hands off to it so the behavior matches the system tool exactly.
- **`rsync` is local-only and copies whole files.** The network protocol and the ssh transport are not implemented, and
  a source or destination naming a remote host is rejected rather than mistaken for a filename. Files are compared by
  size and modification time; the rolling-checksum delta transfer is not implemented, so a changed file is copied in
  full.
- **`lsof` reports network endpoints, not open files.** Enumerating a process's file handles on Windows requires
  undocumented kernel interfaces and administrator rights; use Sysinternals `handle.exe` or Process Explorer for that.
- **Wrapper commands need the real tool installed.** `ssh`, `make`, `python3`, `node` and the rest of that group locate
  and run a separately installed Windows build (see [Native commands vs. wrappers](#native-commands-vs-wrappers)).
- **Some Linux commands are deliberately absent.** These are not on a roadmap; each would be actively harmful as a
  translation:
  - `setfacl` — Windows ACLs carry deny entries, ordering and inheritance that `rwx` cannot express, so writing an ACL
    through a POSIX-shaped interface would silently destroy access-control information. `getfacl` reads them safely.
  - `fdisk`, `mkfs` — partition editing and formatting, where a translation bug destroys a disk. Use `diskpart` or
    `Format-Volume`.
  - `iptables` — Windows Firewall has no chains, tables or targets; a familiar name over an unfamiliar model would be a
    lie. Use `netsh advfirewall`.
  - `strace` — needs kernel ETW providers or a driver. Use Process Monitor.
  - `useradd`, `usermod`, `userdel`, `passwd`, `groupadd` — account mutation is destructive, admin-only, and lossy
    (Windows has SIDs, not uids). Use `net user` and `net localgroup`.

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
| `head`     | Yes       | first lines/bytes, `-n -c -q -v`      | buffered scanner/`io.CopyN`                           | no old-style `-NUM` shorthand                                    |
| `tail`     | Yes       | last lines/bytes, `-n -c -f -q -v`    | in-memory tail, `-f` polls for appended data          | `-f` supports one file at a time                                 |
| `wc`       | Yes       | line/word/byte/char counts            | streaming byte scan                                   | word-splitting is ASCII-whitespace only                          |
| `sort`     | Yes       | `-n -r -u -k`                         | `sort.SliceStable`                                    | no locale collation, single-field `-k`                           |
| `uniq`     | Yes       | `-c -d -u`, adjacent dedup            | line scan                                             | input must already be sorted for full dedup, like GNU uniq       |
| `cut`      | Yes       | `-f -c -d`                            | field/rune selection                                  | no `-b` (byte) mode                                               |
| `tr`       | Yes       | translate/delete/squeeze, `a-z` ranges| rune-by-rune stream                                   | no `\`-escapes or `[:class:]` sets                                |
| `tee`      | Yes       | `-a`                                  | `io.MultiWriter`                                      | —                                                                 |
| `printf`   | Yes       | `%s %d %f %x %%`, `\n\t\r` escapes    | small format parser                                   | no width/precision modifiers                                     |
| `test`     | Yes       | `-f -d -e -z -n`, string/int compare  | `os.Stat` + comparisons                               | no `-o`/`-a` combinators                                          |
| `true`     | Yes       | exit 0                                | —                                                      | —                                                                 |
| `false`    | Yes       | exit 1                                | —                                                      | —                                                                 |
| `which`    | Yes       | resolve on `PATH`                     | `PATH`/`PATHEXT` search                                | —                                                                 |
| `date`     | Yes       | `+FORMAT`, `-u`                       | Go `time` + `%`-token mapping                         | subset of strftime tokens                                        |
| `sleep`    | Yes       | fractional seconds, `s/m/h/d` suffix  | `time.Sleep`                                          | —                                                                 |
| `sha256sum`/`sha1sum`/`md5sum` | Yes | `-c` verify                  | Go `crypto/*` hashes                                  | —                                                                 |
| `cksum`    | Yes       | CRC + byte count                      | `hash/crc32` (IEEE)                                   | not the exact POSIX cksum polynomial                              |
| `zip`      | Yes       | `-r`                                  | Go `archive/zip`                                      | no password/encryption support                                   |
| `unzip`    | Yes       | `-d`                                  | Go `archive/zip`, zip-slip safe                       | —                                                                 |
| `tar`      | Yes       | `-c -x -t -v -z -f`                   | Go `archive/tar` + `compress/gzip`                    | no bzip2/xz                                                       |
| `stat`     | Yes       | file metadata                         | `os.Stat` + `FormatMode`                              | no `-c FORMAT`                                                    |
| `file`     | Yes       | magic-byte + text heuristics          | reads first 512 bytes                                 | small signature set                                               |
| `du`       | Yes       | `-s -h`                               | `filepath.Walk`, per-dir totals                       | no hard-link double-count avoidance                               |
| `df`       | Yes       | `-h`                                  | `GetDiskFreeSpaceExW`                                 | lists drive letters, not Unix-style mount points                  |
| `tree`     | Yes       | `-L -a`                               | recursive `os.ReadDir`                                | —                                                                 |
| `realpath` | Yes       | clean + resolve                       | `internal/paths` + `filepath.EvalSymlinks`             | —                                                                 |
| `readlink` | Yes       | `-f`                                  | `os.Readlink`/`filepath.EvalSymlinks`                  | —                                                                 |
| `pgrep`    | Yes       | `-l`, substring match                 | Tool Help snapshot                                     | matches executable filename only, not full command line          |
| `pkill`    | Yes       | substring match                       | Tool Help snapshot + `TerminateProcess`                | matches executable filename only, not full command line          |
| `pstree`   | Yes       | `-p`                                  | Tool Help snapshot, PID/PPID tree                       | —                                                                 |
| `nl`       | Yes       | `-b a\|t`                             | line scan                                              | no `-i`/`-s`/`-w` customization                                   |
| `tac`      | Yes       | reverse line order                    | buffers all lines                                      | —                                                                 |
| `rev`      | Yes       | reverse each line                     | rune-based                                              | —                                                                 |
| `seq`      | Yes       | `FIRST [STEP] LAST`                   | float loop                                              | no `-s`/`-f` format control                                       |
| `yes`      | Yes       | repeat text, honors broken pipe       | batched writes                                          | —                                                                 |
| `basename` | Yes       | strip dir + optional suffix           | `filepath.Base`                                         | —                                                                 |
| `dirname`  | Yes       | print directory component             | `filepath.Dir`                                          | —                                                                 |
| `expand`   | Yes       | `-t`                                  | column-tracked tab expansion                            | —                                                                 |
| `unexpand` | Yes       | `-t -a`                               | column-tracked tabify                                   | `-a` converts every space run, not just runs of 2+               |
| `fold`     | Yes       | `-w`                                  | rune-count wrapping                                     | no `-s` (break at word boundaries)                                |
| `base64`   | Yes       | `-d`                                  | `encoding/base64`                                        | —                                                                 |
| `xxd`      | Yes       | `-g -r`                               | `encoding/hex`                                           | no `-l`/`-s` offset controls                                      |
| `strings`  | Yes       | `-n`                                  | printable-run scan                                       | —                                                                 |
| `fmt`      | Yes       | `-w`                                  | greedy paragraph fill                                    | no hyphenation, no `-s`/`-u` modes                                |
| `paste`    | Yes       | `-d`                                  | parallel line scan                                       | —                                                                 |
| `join`     | Yes       | join on field 1                       | hash-join                                                | no `-1`/`-2`/`-t` field/delimiter selection                       |
| `comm`     | Yes       | `-1 -2 -3`                            | merge of two sorted line lists                           | —                                                                 |
| `cmp`      | Yes       | first-difference report               | byte-by-byte scan                                        | no `-l` (list all differing bytes)                                |
| `diff`     | Yes       | normal and `-u` unified               | O(n·m) LCS dynamic-programming diff                       | not O(ND); slow on very large files                               |
| `split`    | Yes       | `-l -b`                               | line/byte chunking, `aa`/`ab`/... suffixes                | —                                                                 |
| `csplit`   | Yes       | line-number and `/regex/` boundaries  | line scan + `regexp`                                      | no `{N}` repeat-count syntax                                      |
| `shuf`     | Yes       | `-n`                                  | `math/rand.Shuffle`                                       | —                                                                 |
| `gzip`     | Yes       | `-d -k`                               | `compress/gzip`                                            | —                                                                 |
| `ln`       | Yes       | `-s -f`                               | `os.Link`/`os.Symlink`                                     | symlinks need Developer Mode or admin rights on Windows           |
| `mktemp`   | Yes       | `-d`, `XXXX` templates                | `os.CreateTemp`/`os.MkdirTemp`                              | —                                                                 |
| `truncate` | Yes       | `-s`                                  | `os.Truncate`                                               | —                                                                 |
| `sync`     | Yes       | no-op                                 | —                                                            | nothing to flush across separate per-command processes            |
| `umask`    | Yes       | report/validate mask                  | —                                                            | cannot persist to the parent shell, like `cd`                     |
| `uname`    | Yes       | `-a -s -n -r -m`                      | static Windows values + `runtime.GOARCH`                     | kernel-name/release are fixed strings, not real kernel version   |
| `arch`     | Yes       | prints machine arch                   | `runtime.GOARCH` mapping                                     | —                                                                 |
| `id`       | Yes       | uid/gid/groups                        | `os/user`                                                     | Windows has no separate primary-group SID the way Linux does     |
| `groups`   | Yes       | group memberships                     | `os/user`                                                     | —                                                                 |
| `uptime`   | Yes       | boot duration                         | `GetTickCount64`                                              | no load average (not exposed on Windows)                          |
| `free`     | Yes       | `-h`                                  | `GlobalMemoryStatusEx`                                        | —                                                                 |
| `lscpu`    | Yes       | arch, core count, model name          | `runtime.NumCPU` + `PROCESSOR_IDENTIFIER` env var             | far fewer fields than real lscpu                                  |
| `lsmem`    | Yes       | memory totals                         | `GlobalMemoryStatusEx`                                        | no per-range breakdown                                            |
| `nproc`    | Yes       | logical processor count               | `runtime.NumCPU`                                               | —                                                                 |
| `getconf`  | Yes       | `PAGE_SIZE`, `_NPROCESSORS_ONLN`, `ARCH` | small lookup table                                          | tiny fixed variable set                                           |
| `hostid`   | Yes       | machine-derived identifier            | FNV hash of hostname                                           | not Linux's actual hostid algorithm                               |
| `tty`      | Yes       | reports console/not-console           | `os.ModeCharDevice` check                                     | —                                                                 |
| `stty`     | Yes       | reports console mode bits             | `GetConsoleMode`                                               | read-only; no mode-changing support yet                           |
| `getent`   | Yes       | `hosts`, `passwd`, `group`            | DNS lookup + `os/user`                                        | `passwd`/`group` limited to the current user                      |
| `last`     | Yes       | recent sign-in events                 | wraps `wevtutil.exe` (Security log, event 4624)                | usually needs administrator rights                                 |
| `users`    | Yes       | interactive session usernames         | Terminal Services (WTS) API                                    | —                                                                 |
| `who`      | Yes       | active sessions                       | Terminal Services (WTS) API                                    | single fixed "console" TTY column                                 |
| `w`        | Yes       | sessions + uptime header              | WTS API + `GetTickCount64`                                     | no idle time or per-session process info                          |
| `top`      | Yes       | `-n -d`, refreshing snapshot          | Tool Help snapshot on a timer                                   | no CPU%/memory columns                                            |
| `timeout`  | Yes       | kills the whole child process tree    | `taskkill /T /F` on expiry                                       | —                                                                 |
| `time`     | Yes       | reports elapsed wall-clock time       | wraps a child process                                            | no user/sys CPU time breakdown                                    |
| `watch`    | Yes       | `-n` interval, `-c` bounded count     | re-runs a child + ANSI clear                                     | `-c` is a linuxcmd-only extension for bounded/scripted use         |
| `env`      | Yes       | `NAME=value` overrides                | modifies `cmd.Env`                                                | —                                                                 |
| `printenv` | Yes       | one or all variables                  | `os.Environ`/`os.LookupEnv`                                       | —                                                                 |
| `whereis`  | Yes       | binary + linuxcmd-registration check  | `PATH`/`PATHEXT` search + `command.Lookup`                        | no source/man-page location (linuxcmd has neither)                |
| `nice`     | Yes       | `-n`                                  | maps nice value to a Win32 priority class                          | 6 priority classes, not a continuous -20..19 scale                |
| `renice`   | Yes       | `-p`                                  | `OpenProcess` + `SetPriorityClass`                                  | same coarse priority-class mapping as `nice`                      |
| `nohup`    | Yes       | detach + `nohup.out`                  | `CREATE_NEW_PROCESS_GROUP`                                          | —                                                                 |
| `xargs`    | Yes       | `-n`                                  | whitespace-split stdin tokens, batched exec                        | no `-0`/quoting-aware splitting                                   |
| `less`     | Yes       | pages on a real console               | shares `more`'s implementation                                     | no backward scroll or `/search` yet (a `more`-equivalent subset)  |
| `more`     | Yes       | pages on a real console               | "-- More --" prompt every 24 lines                                 | —                                                                 |
| `column`   | Yes       | `-t -s`                               | field width scan                                                   | no true multi-column fill mode without `-t`                        |
| `expr`     | Yes       | `+ - * / % = != < <= > >= :`          | small recursive evaluator                                          | no `&`/`|` boolean operators                                       |
| `factor`   | Yes       | trial division                        | —                                                                    | slow for very large primes                                        |
| `bc`       | Yes       | `+ - * / % ^ ()`                      | small recursive-descent parser over `float64`                       | not arbitrary precision like real bc; no variables/functions      |
| `help`     | Yes       | lists commands / one summary line     | `command.Names()` + `Summary()`                                     | no per-flag usage detail (not tracked by the command registry)    |
| `curl`     | Yes       | `-o -O -L -H -X -I`                   | `net/http`                                                          | no upload (`-T`/`-d`) flags yet                                   |
| `wget`     | Yes       | `-O`                                  | `net/http`                                                          | no recursive/mirroring modes                                      |
| `nc`       | Yes       | `-l -u -w`                            | `net.Dial`/`net.Listen`                                             | exits once either direction stops, not on explicit close          |
| `host`     | Yes       | A/AAAA lookup                          | `net.LookupHost`                                                    | no MX/TXT/NS (use `dig` for those)                                 |
| `nslookup` | Yes       | A/AAAA lookup                          | `net.LookupHost`                                                    | no interactive mode, no server selection                          |
| `dig`      | Yes       | A/AAAA/MX/TXT/NS                       | `net` package resolver functions                                    | uses the system resolver, not a raw query to a chosen server      |
| `traceroute` | Yes     | wraps `tracert.exe`                    | `os/exec`                                                            | output format matches `tracert`, not GNU traceroute                |
| `netstat`  | Yes       | wraps `netstat.exe`                   | `os/exec`                                                            | flags are Windows netstat's own, not GNU netstat's                |
| `arp`      | Yes       | wraps `arp.exe`                        | `os/exec`                                                            | —                                                                 |
| `route`    | Yes       | `print` by default; passes others through | `os/exec`                                                        | verb-first syntax (`route add`), not flag-first like Linux route  |
| `ss`       | Yes       | `-t -u -l -n`                          | filters `netstat -an` output client-side                            | not a real IP Helper socket table; text-filtered approximation    |
| `hostnamectl` | Yes    | `status` only                          | `os.Hostname`                                                        | `set-hostname` needs a reboot + different API; not implemented    |
| `journalctl` | Yes     | `-n`                                   | wraps `wevtutil.exe` (Application log)                                | —                                                                 |
| `dmesg`    | Yes       | recent System log entries              | wraps `wevtutil.exe`                                                  | —                                                                 |
| `systemctl` | Yes      | `status/start/stop/restart`            | wraps `sc.exe`                                                       | start/stop/restart need administrator rights, enforced by `sc.exe`|
| `service`  | Yes       | `NAME status/start/stop/restart`       | reuses `systemctl`'s `sc.exe` wrapper                                 | —                                                                 |
| `crontab`  | Yes       | `-l`                                   | wraps `schtasks.exe /query`                                          | no install-from-file; use `at` for one-time tasks                 |
| `at`       | Yes       | `TIME COMMAND`                         | `schtasks.exe /create /sc once`                                       | —                                                                 |
| `shutdown` | Yes       | `-r/-h TIME` ("now"/"+MIN"/seconds)    | wraps `shutdown.exe`                                                  | no absolute clock-time (`HH:MM`) argument, only relative delays   |
| `reboot`   | Yes       | wraps `shutdown -r now`                | —                                                                     | —                                                                 |
| `poweroff` | Yes       | wraps `shutdown -h now`                | —                                                                     | —                                                                 |
| `chmod`    | Yes       | `-R`, octal and `+w/-w` modes         | `os.Chmod`                                                            | only tracks the read-only attribute; `+x/-x` accepted but no-op   |
| `chown`    | Yes       | sets owner                             | wraps `icacls.exe /setowner`                                          | usually needs admin/SeTakeOwnershipPrivilege                       |
| `chgrp`    | Yes       | grants group Modify access             | wraps `icacls.exe /grant`                                             | approximation only; Windows has no primary-group field             |
| `install`  | Yes       | `-D -m`                                | `fsutil.CopyFile` + `os.MkdirAll` + chmod's mode logic                 | —                                                                 |
| `mkfifo`   | Yes       | reports unsupported                    | —                                                                      | no on-disk FIFO object on Windows                                  |
| `dd`       | Yes       | `if= of= bs= skip= seek= count=`       | block-by-block copy                                                    | `K`/`M`/`G` size suffixes only, no `conv=`                          |
| `shred`    | Yes       | `-u -n`                                | random-data overwrite passes                                           | best-effort; SSD wear leveling can leave data recoverable          |
| `bzip2`    | Yes       | `-d -k` (decompress only)              | `compress/bzip2`                                                       | compression unsupported (Go has no bzip2 writer)                   |
| `xz`       | Yes       | reports unsupported                    | —                                                                      | no XZ codec in Go's standard library                                |
| `sed`      | Yes       | one `s///[gi]` or `d`, `/addr/`/line#  | small hand-written script parser                                       | one command per script; no ranges, `a/i/c/y`, or `;`-chaining      |
| `awk`      | Yes       | `-F`, `/pattern/`, `{print $N,...}`    | whitespace/`-F`-split fields                                            | no variables, arithmetic, user functions, or BEGIN/END              |
| `openssl`  | Yes       | `rand -hex`, `dgst -sha256/-sha1/-md5` | Go `crypto/*`                                                          | small safe subset only, not a full OpenSSL CLI                     |
| `git`      | Yes       | transparent wrapper                    | locates the real `git.exe` on `PATH` (skipping linuxcmd's own dir)     | requires Git for Windows to be separately installed                 |
| `dos2unix`                     | Yes         | `-k -q -f -n`, in place or as a filter    | atomic rewrite of CRLF to LF                                       | skips files containing NUL unless `-f`                             |
| `unix2dos`                     | Yes         | `-k -q -f -n`, in place or as a filter    | normalizes to LF first, then expands                               | same binary-file guard; idempotent on CRLF files                   |
| `od`                           | Yes         | `-A -t -j -N -v`, `-b -c -d -o -x -s -i -l` | native formatter, little-endian words                              | one `-t` format per run (GNU allows several)                       |
| `hexdump`                      | Yes         | `-C -b -c -d -o -x -n -s -v`              | native formatter                                                   | no `-e` format-string language                                     |
| `namei`                        | Yes         | `-l -m -o -x`                             | walks each component with `os.Lstat`                               | shows the Linux-to-Windows path translation                        |
| `man`                          | Partial     | `-k -f -w`                                | docs generated from the command registry                           | documents linuxcmd's own commands, not GNU manuals                 |
| `xclip`                        | Yes         | `-i -o -r -selection`                     | Win32 clipboard (`CF_UNICODETEXT`)                                 | Windows has one clipboard; X11 selections collapse into it         |
| `xsel`                         | Yes         | `-i -o -c -b -p -s`                       | same clipboard as `xclip`                                          | same single-selection note                                         |
| `xdg-open`                     | Yes         | open a file or URL in the default app     | `ShellExecuteW`                                                    | xdg-open's documented exit codes are preserved                     |
| `open`                         | Yes         | alias for `xdg-open`                      | `ShellExecuteW`                                                    | provided because macOS users reach for this name                   |
| `ifconfig`                     | Partial     | `-a -s`, per-interface detail             | `net.Interfaces` + `GetIfEntry` counters                           | read-only; configure with `netsh` or `Set-NetIPAddress`            |
| `lsof`                         | Partial     | `-i[:PORT] -p -t -n -P`                   | `GetExtendedTcpTable`/`GetExtendedUdpTable`                        | network endpoints only; no open-file handles                       |
| `lsblk`                        | Yes         | `-b -f`                                   | `GetLogicalDriveStrings` + storage ioctls                          | volumes with no drive letter are not listed                        |
| `blkid`                        | Yes         | `-s TAG -o value`                         | `GetVolumeInformationW`                                            | UUID is the 32-bit volume serial, not a filesystem UUID            |
| `vmstat`                       | Partial     | `DELAY [COUNT]`, `-S -n`                  | `GlobalMemoryStatusEx` + PDH counters                              | `b`, `buff`, `wa`, `st` have no Windows counterpart and read 0     |
| `iostat`                       | Partial     | `-c -d -k -m`, `DELAY [COUNT]`            | PDH counters + `IOCTL_DISK_PERFORMANCE`                            | `%iowait`/`%steal` read 0; Windows folds I/O wait into idle        |
| `getfacl`                      | Yes         | `-n -c`                                   | `GetNamedSecurityInfoW` + ACE walk                                 | read-only by design; `setfacl` is deliberately absent              |
| `mount`                        | Partial     | list mounts, attach a share               | volume enumeration + `WNetAddConnection2`                          | network shares only; no image or arbitrary-path mounts             |
| `umount`                       | Partial     | detach a mount                            | `WNetCancelConnection2`                                            | mapped drives only; local volumes need `mountvol` and admin        |
| `sudo`                         | Partial     | run a command elevated                    | Windows 11 `sudo.exe` if present, else UAC                         | without the system sudo the command gets a NEW console             |
| `su`                           | Partial     | `-c`, run as another user                 | wraps `runas.exe`                                                  | always prompts interactively; passwords cannot be piped            |
| `rsync`                        | Partial     | `-a -r -v -n -u -t --delete --exclude`    | size+mtime comparison, whole-file copy                             | local paths only; no delta algorithm and no ssh transport          |
| `md5sum`                       | Yes         | digest files, `-c` to verify              | `crypto/md5`                                                       | —                                                                  |
| `sha1sum`                      | Yes         | digest files, `-c` to verify              | `crypto/sha1`                                                      | —                                                                  |
| `ssh`                          | Yes         | transparent wrapper                       | finds `ssh.exe`, incl. `System32\OpenSSH`                          | needs the OpenSSH Client optional feature                          |
| `scp`                          | Yes         | transparent wrapper                       | same lookup as `ssh`                                               | needs the OpenSSH Client optional feature                          |
| `sftp`                         | Yes         | transparent wrapper                       | same lookup as `ssh`                                               | needs the OpenSSH Client optional feature                          |
| `make`                         | Yes         | transparent wrapper                       | finds `make.exe`/`mingw32-make.exe`/`gnumake.exe`                  | never falls back to `nmake.exe`, which is not GNU make             |
| `cmake`                        | Yes         | transparent wrapper                       | finds `cmake.exe`, incl. `Program Files\CMake\bin`                 | needs CMake installed                                              |
| `gcc`                          | Yes         | transparent wrapper                       | finds `gcc.exe` via PATH, MSYS2, MinGW, Git                        | needs a GCC toolchain installed                                    |
| `g++`                          | Yes         | transparent wrapper                       | finds `g++.exe` via PATH, MSYS2, MinGW, Git                        | needs a GCC toolchain installed                                    |
| `python3`                      | Yes         | transparent wrapper                       | finds `python3.exe`, `python.exe`, or `py.exe -3`                  | needs Python installed                                             |
| `perl`                         | Yes         | transparent wrapper                       | finds `perl.exe`, incl. Git for Windows' copy                      | needs Perl installed                                               |
| `ruby`                         | Yes         | transparent wrapper                       | finds `ruby.exe`                                                   | needs RubyInstaller                                                |
| `node`                         | Yes         | transparent wrapper                       | finds `node.exe`, incl. `Program Files\nodejs`                     | needs Node.js installed                                            |
| `npm`                          | Yes         | transparent wrapper                       | finds `npm.cmd`, run through `COMSPEC`                             | needs Node.js; `.cmd` shims cannot be launched directly            |
| `jq`                           | Yes         | transparent wrapper                       | finds `jq.exe` or `jq-win64.exe`                                   | needs jq installed                                                 |
| `rg`                           | Yes         | transparent wrapper                       | finds `rg.exe`                                                     | needs ripgrep installed                                            |
| `fd`                           | Yes         | transparent wrapper                       | finds `fd.exe`                                                     | needs fd installed                                                 |
| `vim`                          | Yes         | transparent wrapper                       | finds `vim.exe`/`gvim.exe`, incl. Git's copy                       | needs Vim installed; runs on the real console                      |
| `nano`                         | Yes         | transparent wrapper                       | finds `nano.exe` via MSYS2 or Git for Windows                      | needs nano installed                                               |
| `tmux`                         | Yes         | transparent wrapper                       | finds `tmux.exe` via MSYS2 or Cygwin                               | needs MSYS2/Cygwin; no native Windows tmux exists                  |
| `screen`                       | Yes         | transparent wrapper                       | finds `screen.exe` via Cygwin or MSYS2                             | needs Cygwin/MSYS2                                                 |
| `tcpdump`                      | Yes         | transparent wrapper                       | finds `tcpdump.exe` or `windump.exe`                               | needs Npcap + WinDump; `pktmon` is not a substitute                |
| `nmap`                         | Yes         | transparent wrapper                       | finds `nmap.exe`, incl. `Program Files\Nmap`                       | needs Nmap for Windows                                             |

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
architecture: [github.com/khanalsaroj/linuxcmd/releases](https://github.com/khanalsaroj/linuxcmd/releases). Each release
includes `checksums.txt` (SHA-256) for the archives.

## License

[MIT](LICENSE) © Saroj Khanal
