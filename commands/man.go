package commands

import (
	"fmt"
	"sort"
	"strings"

	"linuxcmd/internal/command"
	"linuxcmd/internal/parser"
	"linuxcmd/internal/version"
)

// manCommand renders documentation for linuxcmd's own commands. It does
// not read Linux man pages -- there are none on a Windows box -- and it
// deliberately does not try to restate the GNU manuals. What it
// documents is the part a user cannot look up anywhere else: which
// options this implementation actually accepts, and where its behavior
// diverges from Linux because Windows works differently.
//
// Pages come from three places, in order: a curated entry in manPages, a
// shared note for the divergence group the command belongs to, and
// finally a generated stub built from the registry so that every
// registered command has a page rather than a "no entry" error.
type manCommand struct{}

func (manCommand) Name() string    { return "man" }
func (manCommand) Summary() string { return "show documentation for a linuxcmd command" }

var manSpec = parser.Spec{
	{Short: 'k', Long: "apropos"},
	{Short: 'f', Long: "whatis"},
	{Short: 'w', Long: "where"},
}

// manPage is one command's documentation. Every field is optional; empty
// sections are skipped when rendering.
type manPage struct {
	Synopsis    string
	Description string
	Options     []string
	// Windows describes behavior specific to this implementation. This
	// is the section worth reading and the reason this command exists.
	Windows string
	SeeAlso []string
}

func (manCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, manSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "man: %s\n", err)
		return command.ExitUsage
	}

	if res.Bool('k', "apropos") {
		return manApropos(ctx, res.Positional)
	}
	if len(res.Positional) == 0 {
		fmt.Fprintln(ctx.Stderr, "What manual page do you want?")
		fmt.Fprintln(ctx.Stderr, "For example, try 'man ls'. Use 'man -k KEYWORD' to search.")
		return command.ExitUsage
	}

	exit := command.ExitSuccess
	for i, name := range res.Positional {
		c, ok := command.Lookup(name)
		if !ok {
			fmt.Fprintf(ctx.Stderr, "No manual entry for %s\n", name)
			exit = command.ExitFailure
			continue
		}
		switch {
		case res.Bool('f', "whatis"):
			fmt.Fprintf(ctx.Stdout, "%s (1) - %s\n", c.Name(), c.Summary())
		case res.Bool('w', "where"):
			fmt.Fprintf(ctx.Stdout, "built in to linuxcmd %s\n", version.String())
		default:
			if i > 0 {
				fmt.Fprintln(ctx.Stdout)
			}
			manRender(ctx, c)
		}
	}
	return exit
}

// manApropos searches names and summaries, which is how a user finds a
// command among the full registry without scrolling the whole list.
func manApropos(ctx *command.Context, terms []string) int {
	if len(terms) == 0 {
		fmt.Fprintln(ctx.Stderr, "man: -k requires a search term")
		return command.ExitUsage
	}
	needle := strings.ToLower(strings.Join(terms, " "))

	found := false
	for _, n := range command.Names() {
		c, ok := command.Lookup(n)
		if !ok {
			continue
		}
		if strings.Contains(strings.ToLower(n), needle) ||
			strings.Contains(strings.ToLower(c.Summary()), needle) {
			fmt.Fprintf(ctx.Stdout, "%-12s (1) - %s\n", n, c.Summary())
			found = true
		}
	}
	if !found {
		fmt.Fprintf(ctx.Stderr, "%s: nothing appropriate.\n", needle)
		return command.ExitFailure
	}
	return command.ExitSuccess
}

func manRender(ctx *command.Context, c command.Command) {
	name := c.Name()
	page := manPages[name]

	manSection(ctx, "NAME", fmt.Sprintf("%s - %s", name, c.Summary()))

	synopsis := page.Synopsis
	if synopsis == "" {
		synopsis = name + " [OPTION]... [ARGUMENT]..."
	}
	manSection(ctx, "SYNOPSIS", synopsis)

	if page.Description != "" {
		manSection(ctx, "DESCRIPTION", page.Description)
	}
	if len(page.Options) > 0 {
		manSection(ctx, "OPTIONS", strings.Join(page.Options, "\n"))
	}

	// Prefer a command-specific note, then the shared note for whichever
	// divergence group the command falls into.
	windows := page.Windows
	if windows == "" {
		windows = manGroupNote(name)
	}
	if windows != "" {
		manSection(ctx, "ON WINDOWS", windows)
	}

	if len(page.SeeAlso) > 0 {
		manSection(ctx, "SEE ALSO", strings.Join(page.SeeAlso, ", "))
	}
	manSection(ctx, "LINUXCMD", fmt.Sprintf(
		"Part of linuxcmd %s. Run 'linuxcmd --list-commands' for the full command list,\nor 'man -k KEYWORD' to search by topic.", version.String()))
}

// manSection prints a man-style heading with its body indented, wrapping
// nothing: the bodies here are already written to a readable width.
func manSection(ctx *command.Context, heading, body string) {
	fmt.Fprintln(ctx.Stdout, heading)
	for _, line := range strings.Split(body, "\n") {
		if line == "" {
			fmt.Fprintln(ctx.Stdout)
			continue
		}
		fmt.Fprintf(ctx.Stdout, "       %s\n", line)
	}
	fmt.Fprintln(ctx.Stdout)
}

// manGroupNote returns the shared ON WINDOWS text for commands that
// diverge from Linux for the same underlying reason. Keeping one copy
// per reason means the explanation cannot drift between related
// commands.
func manGroupNote(name string) string {
	for _, g := range manGroups {
		for _, member := range g.commands {
			if member == name {
				return g.note
			}
		}
	}
	return ""
}

type manGroup struct {
	commands []string
	note     string
}

var manGroups = []manGroup{
	{
		commands: []string{"chmod", "chown", "chgrp", "umask", "id", "groups", "who", "users", "w"},
		note: "Windows uses ACLs, not POSIX permission bits, and has no numeric uid/gid.\n" +
			"Mode bits are approximated: the read-only attribute maps to the write bit and\n" +
			"all three owner/group/other triplets are reported identically. Ownership\n" +
			"changes are accepted but cannot be represented faithfully. Use getfacl to see\n" +
			"the real Windows access-control entries.",
	},
	{
		commands: []string{"kill", "pkill", "pgrep"},
		note: "Windows has no signal mechanism. Any signal that terminates on Linux maps to\n" +
			"TerminateProcess; signals that a process could catch or that mean something\n" +
			"other than termination (HUP, USR1, USR2, STOP, CONT) have no equivalent and\n" +
			"are rejected rather than silently treated as a kill.",
	},
	{
		commands: []string{"cd", "mkdir", "rmdir", "echo"},
		note: "cmd.exe resolves these names to its own builtin before searching PATH, so the\n" +
			"linuxcmd version does not run unless it is invoked explicitly (mkdir.exe) or\n" +
			"the optional DOSKEY overrides are installed. See the README section on\n" +
			"cmd.exe builtins for the tradeoffs.",
	},
	{
		commands: []string{
			"ssh", "scp", "sftp", "tcpdump", "nmap", "make", "cmake", "gcc", "g++",
			"python3", "perl", "ruby", "node", "npm", "jq", "rg", "fd", "vim", "nano",
			"tmux", "screen", "git",
		},
		note: "This command is a wrapper, not a reimplementation: it locates the real,\n" +
			"separately installed Windows build and hands off to it unchanged. linuxcmd\n" +
			"adds the Linux name, resolution of the Windows-specific filename, lookup in\n" +
			"install directories that are often missing from PATH, and a message naming\n" +
			"what to install when the tool is absent. All arguments, streams and the exit\n" +
			"code pass through untouched.",
	},
}

// manPages holds hand-written documentation. It intentionally covers the
// commands whose Windows behavior is surprising plus the ones with no
// close Linux-on-Windows precedent, rather than attempting a full manual
// for every registered command.
var manPages = map[string]manPage{
	"namei": {
		Synopsis: "namei [-l] [-m] [-o] [-x] PATH...",
		Description: "Walk a pathname one component at a time, printing the type of each component\n" +
			"as it is resolved, and stop at the first component that cannot be read.",
		Options: []string{
			"-m, --modes       show the full mode string instead of a single type character",
			"-o, --owners      show the owner of each component",
			"-l, --long        equivalent to -m -o",
			"-x, --mountpoints mark volume roots, the closest Windows equivalent of a mount point",
		},
		Windows: "This is the fastest way to see what linuxcmd did with a path. Arguments pass\n" +
			"through the same translation as every other command, so the first line of\n" +
			"output shows the argument as typed and the first component shows the Windows\n" +
			"volume it resolved to: ~ becomes the profile drive, /c/Users becomes C:\\Users,\n" +
			"/tmp becomes %TEMP%, and a bare /etc resolves against the current drive.\n" +
			"Junctions and symlinks are reported with their target, since a reparse point is\n" +
			"the usual explanation for a path behaving unexpectedly.",
		SeeAlso: []string{"realpath", "readlink", "stat"},
	},
	"dos2unix": {
		Synopsis: "dos2unix [-k] [-q] [-f] [-n INFILE OUTFILE]... [FILE]...",
		Description: "Convert DOS line endings (CRLF) to Unix line endings (LF). Files are converted\n" +
			"in place. With no file operands, standard input is converted to standard output.",
		Options: []string{
			"-k, --keepdate  keep the original modification time",
			"-q, --quiet     suppress the per-file conversion messages",
			"-f, --force     convert even a file that looks binary",
			"-n, --newfile   take INFILE OUTFILE pairs and leave the input untouched",
		},
		Windows: "Files containing a NUL byte are skipped unless -f is given, because rewriting\n" +
			"CR bytes inside a binary would corrupt it. In-place conversion writes to a\n" +
			"temporary file in the same directory and renames it over the original, so an\n" +
			"interrupted run cannot leave a half-converted file behind. A file already in\n" +
			"the target format is left untouched, including its timestamp.",
		SeeAlso: []string{"unix2dos", "file", "tr"},
	},
	"unix2dos": {
		Synopsis: "unix2dos [-k] [-q] [-f] [-n INFILE OUTFILE]... [FILE]...",
		Description: "Convert Unix line endings (LF) to DOS line endings (CRLF). Files are converted\n" +
			"in place. With no file operands, standard input is converted to standard output.",
		Options: []string{
			"-k, --keepdate  keep the original modification time",
			"-q, --quiet     suppress the per-file conversion messages",
			"-f, --force     convert even a file that looks binary",
			"-n, --newfile   take INFILE OUTFILE pairs and leave the input untouched",
		},
		Windows: "Conversion normalizes to LF first, so running unix2dos on a file that already\n" +
			"uses CRLF leaves it unchanged instead of producing CR CR LF.",
		SeeAlso: []string{"dos2unix", "file"},
	},
	"od": {
		Synopsis: "od [-A RADIX] [-t TYPE] [-j SKIP] [-N COUNT] [-v] [FILE]...",
		Description: "Dump files in octal and other formats. Multiple operands are concatenated and\n" +
			"treated as one stream. Runs of identical lines collapse to a single '*' unless\n" +
			"-v is given.",
		Options: []string{
			"-A RADIX  address radix: d, o, x, or n for no addresses",
			"-t TYPE   output type: a, c, d, f, o, u, x, optionally followed by a byte size",
			"-j SKIP   skip SKIP bytes of input first",
			"-N COUNT  dump at most COUNT bytes",
			"-v        do not collapse repeated lines",
			"-b -c -d -o -x -s -i -l    shorthands for the common -t types",
		},
		Windows: "Multi-byte values are read little-endian, matching the host architecture.\n" +
			"Only one -t format may be given per run; GNU od accepts several and prints one\n" +
			"line per format.",
		SeeAlso: []string{"hexdump", "xxd", "strings"},
	},
	"hexdump": {
		Synopsis: "hexdump [-C] [-b] [-c] [-d] [-o] [-x] [-n LENGTH] [-s SKIP] [-v] [FILE]...",
		Description: "Display file contents in hexadecimal, decimal, octal or character form. With no\n" +
			"format flag, output is two-byte hexadecimal words.",
		Options: []string{
			"-C  canonical: hex bytes alongside a printable-character gutter",
			"-b  one-byte octal",
			"-c  one-byte character",
			"-d  two-byte decimal",
			"-o  two-byte octal",
			"-x  two-byte hexadecimal (the default)",
			"-n  limit the dump to LENGTH bytes",
			"-s  skip SKIP bytes first",
			"-v  do not collapse repeated lines",
		},
		Windows: "The -e format-string language is not implemented. Use od for typed output that\n" +
			"hexdump's flags do not cover, or xxd when a reversible dump is needed.",
		SeeAlso: []string{"od", "xxd", "strings"},
	},
	"xclip": {
		Synopsis: "xclip [-i] [-o] [-r] [-selection SELECTION] [FILE]...",
		Description: "Copy data to the Windows clipboard, or with -o print the clipboard's contents.\n" +
			"With no file operands, input is read from standard input.",
		Options: []string{
			"-i, --in         read into the clipboard (the default)",
			"-o, --out        write the clipboard to standard output",
			"-r, --rmlastnl   strip the trailing newline before copying",
			"-selection NAME  accepted for compatibility; see below",
		},
		Windows: "X11 has three independent selections (PRIMARY, SECONDARY, CLIPBOARD) and\n" +
			"Windows has exactly one clipboard, so -selection is accepted but every value\n" +
			"addresses the same storage. Scripts that keep two different values in PRIMARY\n" +
			"and CLIPBOARD will see them collapse into one. Only Unicode text is handled;\n" +
			"a clipboard holding an image reads as empty.",
		SeeAlso: []string{"xsel"},
	},
	"xsel": {
		Synopsis: "xsel [-i] [-o] [-c] [-b] [-p] [-s]",
		Description: "Read or set the Windows clipboard. With no direction flag, the clipboard is\n" +
			"printed to standard output.",
		Options: []string{
			"-i, --input      read standard input into the clipboard",
			"-o, --output     print the clipboard (the default)",
			"-c, --clear      empty the clipboard",
			"-b -p -s         select CLIPBOARD, PRIMARY or SECONDARY; see below",
		},
		Windows: "Windows has a single clipboard, so -b, -p and -s all address the same storage.",
		SeeAlso: []string{"xclip"},
	},
	"xdg-open": {
		Synopsis:    "xdg-open FILE|URL...",
		Description: "Open each file or URL with the application registered to handle it.",
		Windows: "Implemented with ShellExecute, the same mechanism Explorer uses for a\n" +
			"double-click, so file associations and URL protocol handlers behave exactly as\n" +
			"they do in the shell. Also registered as 'open'. Exit codes follow the xdg-open\n" +
			"specification: 1 syntax error, 2 file not found, 3 no associated application,\n" +
			"4 the action failed.",
		SeeAlso: []string{"open"},
	},
	"open": {
		Synopsis:    "open FILE|URL...",
		Description: "Open each file or URL with the application registered to handle it.",
		Windows:     "An alias for xdg-open, provided because it is the name macOS users reach for.",
		SeeAlso:     []string{"xdg-open"},
	},
	"lsof": {
		Synopsis:    "lsof [-i[:PORT]] [-n] [-P] [-p PID]",
		Description: "List open network endpoints and the processes that own them.",
		Options: []string{
			"-i[:PORT]  show network connections, optionally filtered to a port",
			"-p PID     show only the given process",
			"-n         do not resolve addresses to hostnames (always the case here)",
			"-P         do not resolve port numbers to service names (always the case here)",
		},
		Windows: "Only the network side of lsof is implemented, via the documented TCP and UDP\n" +
			"connection tables. Listing open *files* per process requires enumerating kernel\n" +
			"handles through undocumented interfaces and administrator rights, so it is not\n" +
			"attempted rather than half-supported. For open file handles use Sysinternals\n" +
			"handle.exe or Process Explorer.",
		SeeAlso: []string{"ss", "netstat", "ps"},
	},
	"mount": {
		Synopsis: "mount [-t TYPE] [SOURCE DIRECTORY]",
		Description: "With no arguments, list mounted volumes and connected network shares. With a\n" +
			"source and target, connect a network share to a drive letter.",
		Windows: "Windows has no unified namespace to mount into: local volumes get drive\n" +
			"letters, and network shares are connected with 'net use'. Mounting a share on a\n" +
			"drive letter (mount //server/share Z:) maps onto that directly. Mounting disk\n" +
			"images or arbitrary filesystems at arbitrary directories is not supported;\n" +
			"use PowerShell's Mount-DiskImage or mountvol for those.",
		SeeAlso: []string{"umount", "df", "lsblk"},
	},
	"umount": {
		Synopsis:    "umount TARGET",
		Description: "Disconnect a network share or remove a volume's drive-letter assignment.",
		Windows: "Disconnects mapped network drives via the same mechanism as 'net use /delete'.\n" +
			"Local volumes are detached with mountvol, which requires administrator rights.",
		SeeAlso: []string{"mount", "df"},
	},
	"rsync": {
		Synopsis: "rsync [-a] [-r] [-v] [-n] [--delete] SOURCE... DEST",
		Description: "Synchronize files and directories, copying only what differs by size or\n" +
			"modification time.",
		Options: []string{
			"-a, --archive   recurse and preserve modification times",
			"-r, --recursive recurse into directories",
			"-v, --verbose   list each file transferred",
			"-n, --dry-run   report what would be transferred without writing anything",
			"    --delete    remove files from DEST that no longer exist in SOURCE",
		},
		Windows: "Local paths only. The rsync network protocol and rsync-over-ssh transfers are\n" +
			"not implemented, and a SOURCE or DEST containing a host specification is\n" +
			"rejected rather than silently treated as a filename. Files are compared by size\n" +
			"and modification time; the rolling-checksum delta algorithm is not used, so a\n" +
			"changed file is copied in full. Permissions are not preserved because Windows\n" +
			"ACLs have no POSIX equivalent.",
		SeeAlso: []string{"cp", "install", "tar"},
	},
	"getfacl": {
		Synopsis:    "getfacl [-a] [FILE]...",
		Description: "Display the access-control entries attached to each file.",
		Windows: "Reports real Windows ACL entries rather than the POSIX approximation that\n" +
			"chmod and ls -l show, so this is the accurate view of who can do what. Windows\n" +
			"access rights are richer than rwx: entries can allow or deny, can be inherited\n" +
			"from a parent directory, and distinguish rights that POSIX collapses together.\n" +
			"Output maps them onto the closest rwx spelling and marks entries that cannot\n" +
			"be represented. setfacl is deliberately not provided: a lossy translation in\n" +
			"the write direction would silently destroy access-control information.",
		SeeAlso: []string{"ls", "chmod", "stat"},
	},
	"ifconfig": {
		Synopsis:    "ifconfig [-a] [INTERFACE]",
		Description: "Display network interface configuration.",
		Windows: "Read-only. Interface configuration changes go through netsh or the\n" +
			"Set-NetIPAddress cmdlets, which need administrator rights and have no faithful\n" +
			"ifconfig spelling. Interface names are the Windows adapter names, which are\n" +
			"descriptive ('Ethernet', 'Wi-Fi') rather than eth0/wlan0.",
		SeeAlso: []string{"ip", "netstat", "ss", "route"},
	},
	"lsblk": {
		Synopsis:    "lsblk [-b]",
		Description: "List block devices: physical disks and the volumes on them.",
		Windows: "Volumes are reported by drive letter, since that is how Windows addresses\n" +
			"them. A volume with no drive letter (a recovery or EFI system partition) is\n" +
			"listed by its volume GUID path.",
		SeeAlso: []string{"blkid", "df", "mount"},
	},
	"blkid": {
		Synopsis:    "blkid [DEVICE]...",
		Description: "Show volume identifiers and filesystem types.",
		Windows: "UUID is the NTFS/FAT volume serial number, which is shorter than a Linux\n" +
			"filesystem UUID and is regenerated when a volume is reformatted. TYPE is the\n" +
			"Windows filesystem name (NTFS, FAT32, exFAT, ReFS).",
		SeeAlso: []string{"lsblk", "df"},
	},
	"vmstat": {
		Synopsis:    "vmstat [DELAY [COUNT]]",
		Description: "Report memory, paging and CPU activity, sampled every DELAY seconds.",
		Windows: "Columns that have no Windows counterpart are reported as zero rather than\n" +
			"invented: there is no separate buffer cache figure, and swap in/out counts\n" +
			"page-file activity, which is not the same accounting Linux uses. The first\n" +
			"sample covers the interval since boot, matching vmstat.",
		SeeAlso: []string{"free", "top", "iostat"},
	},
	"iostat": {
		Synopsis:    "iostat [DELAY [COUNT]]",
		Description: "Report CPU utilization and per-disk I/O activity.",
		Windows: "Per-disk figures come from the Windows disk performance counters. The 'iowait'\n" +
			"column has no Windows equivalent and is reported as zero; Windows accounts for\n" +
			"I/O waiting inside idle time instead of as a separate CPU state.",
		SeeAlso: []string{"vmstat", "df", "top"},
	},
	"sudo": {
		Synopsis:    "sudo COMMAND [ARGUMENT]...",
		Description: "Run a command with administrator rights.",
		Windows: "Windows 11 ships a native sudo; when it is present this hands off to it so\n" +
			"behavior matches the system tool exactly. Otherwise the command is launched\n" +
			"elevated via the UAC consent prompt, which starts it in a NEW console window.\n" +
			"That is the important difference from Linux: an elevated process cannot\n" +
			"inherit the current console's handles, so 'sudo cmd > file' and pipelines\n" +
			"through sudo will not capture output. There is no sudoers file, no credential\n" +
			"caching, and no -u to run as another user.",
		SeeAlso: []string{"su", "id"},
	},
	"su": {
		Synopsis:    "su [-] [USER]",
		Description: "Start a shell as another user.",
		Windows: "Maps onto the Windows 'runas' mechanism, which always prompts for the target\n" +
			"account's password interactively and opens a new console window. Passwords\n" +
			"cannot be supplied on the command line or piped in.",
		SeeAlso: []string{"sudo", "id", "whoami"},
	},
	"man": {
		Synopsis:    "man [-k] [-f] [-w] [COMMAND]...",
		Description: "Show documentation for a linuxcmd command.",
		Options: []string{
			"-k, --apropos  search command names and summaries for a keyword",
			"-f, --whatis   print the one-line summary only",
			"-w, --where    report where the page comes from",
		},
		Windows: "There are no Linux man pages on a Windows system, so this documents linuxcmd's\n" +
			"own implementations rather than GNU coreutils. The section worth reading is ON\n" +
			"WINDOWS, which records where this implementation diverges from Linux and why.\n" +
			"Commands without a hand-written page still get one generated from the registry.",
		SeeAlso: []string{"help", "whereis", "which"},
	},
}

// manPageNames is used by tests to check that every curated page names a
// registered command.
func manPageNames() []string {
	names := make([]string, 0, len(manPages))
	for n := range manPages {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

func init() { command.Register(manCommand{}) }
