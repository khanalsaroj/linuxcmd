package commands

import (
	"os"
	"path/filepath"

	"linuxcmd/internal/command"
)

// externalTools is the table of Linux command names that linuxcmd
// provides as a handoff to a real installed Windows program rather than
// as a native implementation. See passthrough.go for why these are
// wrappers and how self-recursion is avoided.
//
// Everything here has an official, well-maintained Windows build. The
// value linuxcmd adds is name compatibility (python3, make, fd),
// resolution of the Windows-specific filename (mingw32-make.exe,
// npm.cmd), lookup in install locations that are commonly missing from
// PATH (System32\OpenSSH, Program Files\CMake\bin), and an actionable
// message when the tool simply is not installed.
var externalTools = []externalTool{
	// --- Remote access. Windows ships OpenSSH as an optional feature
	// that is enabled by default on current releases, but it lives in
	// System32\OpenSSH, which is on PATH only sometimes.
	{
		name:       "ssh",
		summary:    "run the installed OpenSSH client",
		candidates: exeCandidates("ssh.exe"),
		extraDirs:  opensshDirs,
		hint:       "enable OpenSSH Client in Settings > System > Optional features",
	},
	{
		name:       "scp",
		summary:    "copy files over SSH via the installed OpenSSH client",
		candidates: exeCandidates("scp.exe"),
		extraDirs:  opensshDirs,
		hint:       "enable OpenSSH Client in Settings > System > Optional features",
	},
	{
		name:       "sftp",
		summary:    "transfer files over SSH via the installed OpenSSH client",
		candidates: exeCandidates("sftp.exe"),
		extraDirs:  opensshDirs,
		hint:       "enable OpenSSH Client in Settings > System > Optional features",
	},

	// --- Packet capture and scanning.
	{
		name: "tcpdump",
		// WinDump is the Windows port of tcpdump and keeps its command
		// line; Windows' own pktmon is unrelated and not a substitute.
		summary:    "capture packets via the installed tcpdump or WinDump",
		candidates: exeCandidates("tcpdump.exe", "windump.exe"),
		extraDirs:  npcapDirs,
		hint:       "install Npcap and WinDump, or use Windows pktmon",
	},
	{
		name:       "nmap",
		summary:    "scan hosts and services via the installed Nmap",
		candidates: exeCandidates("nmap.exe"),
		extraDirs:  func() []string { return programFilesDirs("Nmap") },
		hint:       "install Nmap for Windows from nmap.org",
	},

	// --- Build tooling. nmake.exe is deliberately NOT a make candidate:
	// it shares the name but not the syntax, so silently running it
	// against a GNU Makefile would fail in confusing ways.
	{
		name:       "make",
		summary:    "run Makefiles via the installed GNU make",
		candidates: exeCandidates("make.exe", "mingw32-make.exe", "gnumake.exe"),
		extraDirs:  toolchainDirs,
		hint:       "install MSYS2, MinGW-w64, or Git for Windows",
	},
	{
		name:       "cmake",
		summary:    "generate build files via the installed CMake",
		candidates: exeCandidates("cmake.exe"),
		extraDirs:  func() []string { return programFilesDirs("CMake", "bin") },
		hint:       "install CMake from cmake.org or via winget",
	},
	{
		name:       "gcc",
		summary:    "compile C via the installed GCC toolchain",
		candidates: exeCandidates("gcc.exe"),
		extraDirs:  toolchainDirs,
		hint:       "install MSYS2 or MinGW-w64",
	},
	{
		name:       "g++",
		summary:    "compile C++ via the installed GCC toolchain",
		candidates: exeCandidates("g++.exe"),
		extraDirs:  toolchainDirs,
		hint:       "install MSYS2 or MinGW-w64",
	},

	// --- Language runtimes. py.exe is the Python launcher, which needs
	// an explicit "-3" to select Python 3.
	{
		name:    "python3",
		summary: "run the installed Python 3 interpreter",
		candidates: []externalCandidate{
			{File: "python3.exe"},
			{File: "python.exe"},
			{File: "py.exe", Args: []string{"-3"}},
		},
		hint: "install Python from python.org or via winget",
	},
	{
		name:       "perl",
		summary:    "run the installed Perl interpreter",
		candidates: exeCandidates("perl.exe"),
		extraDirs:  toolchainDirs,
		hint:       "install Strawberry Perl, or use the Perl in Git for Windows",
	},
	{
		name:       "ruby",
		summary:    "run the installed Ruby interpreter",
		candidates: exeCandidates("ruby.exe"),
		hint:       "install RubyInstaller for Windows",
	},
	{
		name:       "node",
		summary:    "run the installed Node.js runtime",
		candidates: exeCandidates("node.exe"),
		extraDirs:  func() []string { return programFilesDirs("nodejs") },
		hint:       "install Node.js from nodejs.org or via winget",
	},
	{
		name: "npm",
		// npm ships as a .cmd shim on Windows, which CreateProcess
		// cannot launch directly; runExternal routes it through COMSPEC.
		summary:    "manage Node.js packages via the installed npm",
		candidates: exeCandidates("npm.cmd", "npm.exe"),
		extraDirs:  func() []string { return programFilesDirs("nodejs") },
		hint:       "install Node.js, which bundles npm",
	},

	// --- Modern CLI utilities.
	{
		name:       "jq",
		summary:    "query and transform JSON via the installed jq",
		candidates: exeCandidates("jq.exe", "jq-win64.exe"),
		hint:       "install jq via winget, scoop, or jqlang.github.io",
	},
	{
		name:       "rg",
		summary:    "search recursively via the installed ripgrep",
		candidates: exeCandidates("rg.exe"),
		hint:       "install ripgrep via winget or scoop",
	},
	{
		name:       "fd",
		summary:    "find files via the installed fd",
		candidates: exeCandidates("fd.exe", "fdfind.exe"),
		hint:       "install fd via winget or scoop",
	},

	// --- Terminal editors and multiplexers. These are interactive, so
	// runExternal hands them the caller's real console handles.
	{
		name:       "vim",
		summary:    "edit files in the installed Vim",
		candidates: exeCandidates("vim.exe", "gvim.exe"),
		extraDirs:  vimDirs,
		hint:       "install Vim for Windows, or use the Vim in Git for Windows",
	},
	{
		name:       "nano",
		summary:    "edit files in the installed nano",
		candidates: exeCandidates("nano.exe"),
		extraDirs:  toolchainDirs,
		hint:       "install MSYS2, or use the nano in Git for Windows",
	},
	{
		name:       "tmux",
		summary:    "run persistent terminal sessions via the installed tmux",
		candidates: exeCandidates("tmux.exe"),
		extraDirs:  msys2Dirs,
		hint:       "install tmux under MSYS2 or Cygwin",
	},
	{
		name:       "screen",
		summary:    "run persistent terminal sessions via the installed screen",
		candidates: exeCandidates("screen.exe"),
		extraDirs:  msys2Dirs,
		hint:       "install screen under Cygwin or MSYS2",
	},
}

// opensshDirs covers the Windows OpenSSH feature's install directory,
// which is frequently absent from PATH, plus the copy Git for Windows
// bundles.
func opensshDirs() []string {
	dirs := []string{filepath.Join(systemRoot(), "System32", "OpenSSH")}
	return append(dirs, gitUnixToolDirs()...)
}

// toolchainDirs covers the usual MSYS2/MinGW locations plus the POSIX
// tools Git for Windows installs, which together account for most
// working GNU toolchains on a Windows box.
func toolchainDirs() []string {
	return append(msys2Dirs(), gitUnixToolDirs()...)
}

// msys2Dirs lists MSYS2's and Cygwin's default install roots. Neither
// exports an environment variable pointing at itself, so the
// conventional locations are checked directly, honoring an MSYS2_ROOT
// override when one is set.
func msys2Dirs() []string {
	roots := []string{`C:\msys64`, `C:\msys32`, `C:\cygwin64`, `C:\cygwin`}
	if v := os.Getenv("MSYS2_ROOT"); v != "" {
		roots = append([]string{v}, roots...)
	}
	var dirs []string
	for _, r := range roots {
		dirs = append(dirs,
			filepath.Join(r, "usr", "bin"),
			filepath.Join(r, "mingw64", "bin"),
		)
	}
	return dirs
}

// gitUnixToolDirs points at the GNU userland Git for Windows ships,
// which is the most commonly installed POSIX toolset on Windows.
func gitUnixToolDirs() []string {
	var dirs []string
	for _, sub := range [][]string{
		{"Git", "usr", "bin"},
		{"Git", "mingw64", "bin"},
		{"Git", "bin"},
	} {
		dirs = append(dirs, programFilesDirs(sub...)...)
	}
	return dirs
}

func npcapDirs() []string {
	return append(programFilesDirs("Npcap"), programFilesDirs("WinDump")...)
}

func vimDirs() []string {
	return append(programFilesDirs("Vim"), gitUnixToolDirs()...)
}

func init() {
	for _, t := range externalTools {
		command.Register(t)
	}
}
