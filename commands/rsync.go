package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"linuxcmd/internal/command"
	"linuxcmd/internal/fsutil"
	"linuxcmd/internal/output"
	"linuxcmd/internal/parser"
	"linuxcmd/internal/paths"
)

// rsyncCommand synchronizes local directories. It implements the part of
// rsync people use on their own machine -- mirror this tree onto that
// one, skip what has not changed, optionally delete what no longer
// exists -- including the trailing-slash rule on the source, which is
// rsync's most consequential piece of syntax.
//
// It is local-only, by design rather than by omission. rsync's network
// side is a wire protocol plus an ssh transport, and a half-working
// implementation of a sync protocol is a good way to lose data
// silently. A source or destination naming a remote host is rejected
// with a clear message instead of being mistaken for a filename.
//
// Files are compared by size and modification time. The rolling-checksum
// delta transfer that rsync is famous for is not implemented, so a
// changed file is copied whole; over a local filesystem that is usually
// faster anyway, and it is never wrong, only less clever.
type rsyncCommand struct{}

func (rsyncCommand) Name() string    { return "rsync" }
func (rsyncCommand) Summary() string { return "synchronize files and directories locally" }

var rsyncSpec = parser.Spec{
	{Short: 'a', Long: "archive"},
	{Short: 'r', Long: "recursive"},
	{Short: 'v', Long: "verbose"},
	{Short: 'q', Long: "quiet"},
	{Short: 'n', Long: "dry-run"},
	{Short: 'u', Long: "update"},
	{Short: 't', Long: "times"},
	{Short: 'h', Long: "human-readable"},
	{Long: "delete"},
	{Long: "exclude", HasArg: true},
	{Long: "progress"},
	{Long: "stats"},
}

// rsyncOptions is the resolved behavior for one run.
type rsyncOptions struct {
	recursive bool
	verbose   bool
	dryRun    bool
	update    bool
	times     bool
	del       bool
	human     bool
	exclude   string
}

// rsyncStats accumulates what the run did, for the closing summary.
type rsyncStats struct {
	filesConsidered int
	filesCopied     int
	dirsCreated     int
	filesDeleted    int
	bytesCopied     int64
}

func (rsyncCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, rsyncSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "rsync: %s\n", err)
		return command.ExitUsage
	}
	if len(res.Positional) < 2 {
		fmt.Fprintln(ctx.Stderr, "usage: rsync [-a] [-r] [-v] [-n] [--delete] SOURCE... DEST")
		return command.ExitUsage
	}

	archive := res.Bool('a', "archive")
	opts := rsyncOptions{
		recursive: archive || res.Bool('r', "recursive"),
		verbose:   res.Bool('v', "verbose") && !res.Bool('q', "quiet"),
		dryRun:    res.Bool('n', "dry-run"),
		update:    res.Bool('u', "update"),
		times:     archive || res.Bool('t', "times"),
		del:       res.Bool(0, "delete"),
		human:     res.Bool('h', "human-readable"),
	}
	if v, ok := res.Value(0, "exclude"); ok {
		opts.exclude = v
	}

	sources := res.Positional[:len(res.Positional)-1]
	destSpec := res.Positional[len(res.Positional)-1]

	for _, spec := range append(append([]string{}, sources...), destSpec) {
		if isRemoteSpec(spec) {
			fmt.Fprintf(ctx.Stderr, "rsync: %s: remote transfers are not supported\n", spec)
			fmt.Fprintln(ctx.Stderr, "rsync: this build synchronizes local paths only; use scp or a real rsync for network copies")
			return command.ExitFailure
		}
	}

	dest, err := paths.Resolve(destSpec)
	if err != nil {
		output.SimpleErrorf(ctx.Stderr, "rsync", destSpec, err)
		return command.ExitFailure
	}

	if opts.verbose {
		fmt.Fprintln(ctx.Stdout, "sending incremental file list")
	}

	var stats rsyncStats
	exit := command.ExitSuccess
	for _, srcSpec := range sources {
		if !syncOne(ctx, srcSpec, dest, destSpec, &opts, &stats) {
			exit = command.ExitFailure
		}
	}

	if opts.verbose || res.Bool(0, "stats") {
		printRsyncSummary(ctx, &stats, &opts)
	}
	return exit
}

func syncOne(ctx *command.Context, srcSpec, dest, destSpec string, opts *rsyncOptions, stats *rsyncStats) bool {
	// The trailing slash is significant and must be read from the
	// argument as typed, before path resolution cleans it away:
	// "src/" copies the contents of src, "src" copies src itself.
	copyContents := strings.HasSuffix(srcSpec, "/") || strings.HasSuffix(srcSpec, `\`)

	src, err := paths.Resolve(srcSpec)
	if err != nil {
		output.SimpleErrorf(ctx.Stderr, "rsync", srcSpec, err)
		return false
	}
	info, err := os.Stat(src)
	if err != nil {
		output.SimpleErrorf(ctx.Stderr, "rsync", srcSpec, err)
		return false
	}

	if !info.IsDir() {
		target := dest
		// A destination that is an existing directory, or that was
		// written with a trailing slash, receives the file inside it.
		if destInfo, err := os.Stat(dest); err == nil && destInfo.IsDir() {
			target = filepath.Join(dest, info.Name())
		} else if strings.HasSuffix(destSpec, "/") || strings.HasSuffix(destSpec, `\`) {
			target = filepath.Join(dest, info.Name())
		}
		return transferFile(ctx, src, target, info.Name(), info, opts, stats)
	}

	if !opts.recursive {
		fmt.Fprintf(ctx.Stderr, "rsync: skipping directory %s (use -r or -a)\n", srcSpec)
		return true
	}

	destRoot := dest
	if !copyContents {
		destRoot = filepath.Join(dest, info.Name())
	}
	return syncTree(ctx, src, destRoot, opts, stats)
}

func syncTree(ctx *command.Context, srcRoot, destRoot string, opts *rsyncOptions, stats *rsyncStats) bool {
	ok := true
	// present records every relative path the source contributes, so
	// --delete knows what in the destination is extraneous.
	present := map[string]bool{}

	err := filepath.Walk(srcRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "rsync", path, err)
			ok = false
			return nil
		}
		rel, relErr := filepath.Rel(srcRoot, path)
		if relErr != nil {
			return nil
		}
		if rel == "." {
			if !opts.dryRun {
				if mkErr := os.MkdirAll(destRoot, 0o755); mkErr != nil {
					output.Errorf(ctx.Stderr, "rsync", "cannot create", destRoot, mkErr)
					ok = false
					return filepath.SkipAll
				}
			}
			return nil
		}
		if opts.exclude != "" && matchesExclude(rel, opts.exclude) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		present[filepath.ToSlash(rel)] = true
		target := filepath.Join(destRoot, rel)

		if info.IsDir() {
			if _, statErr := os.Stat(target); os.IsNotExist(statErr) {
				stats.dirsCreated++
				if opts.verbose {
					fmt.Fprintf(ctx.Stdout, "%s/\n", filepath.ToSlash(rel))
				}
				if !opts.dryRun {
					if mkErr := os.MkdirAll(target, 0o755); mkErr != nil {
						output.Errorf(ctx.Stderr, "rsync", "cannot create", target, mkErr)
						ok = false
						return filepath.SkipDir
					}
				}
			}
			return nil
		}

		if !transferFile(ctx, path, target, filepath.ToSlash(rel), info, opts, stats) {
			ok = false
		}
		return nil
	})
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "rsync: %s\n", err)
		return false
	}

	if opts.del {
		if !deleteExtraneous(ctx, destRoot, present, opts, stats) {
			ok = false
		}
	}
	return ok
}

// transferFile copies one file if it differs from the destination.
func transferFile(ctx *command.Context, src, target, label string, info os.FileInfo, opts *rsyncOptions, stats *rsyncStats) bool {
	stats.filesConsidered++

	if !needsTransfer(src, target, info, opts) {
		return true
	}
	if opts.verbose {
		fmt.Fprintln(ctx.Stdout, label)
	}
	stats.filesCopied++
	stats.bytesCopied += info.Size()
	if opts.dryRun {
		return true
	}

	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		output.Errorf(ctx.Stderr, "rsync", "cannot create", filepath.Dir(target), err)
		return false
	}
	if err := fsutil.CopyFile(src, target); err != nil {
		output.Errorf(ctx.Stderr, "rsync", "cannot copy", label, err)
		return false
	}
	if opts.times {
		if err := os.Chtimes(target, info.ModTime(), info.ModTime()); err != nil {
			output.Errorf(ctx.Stderr, "rsync", "cannot set times on", label, err)
			return false
		}
	}
	return true
}

// needsTransfer applies rsync's quick check: transfer when the
// destination is missing, differs in size, or differs in modification
// time. With -u, a destination that is newer than the source is kept.
func needsTransfer(src, target string, info os.FileInfo, opts *rsyncOptions) bool {
	destInfo, err := os.Stat(target)
	if err != nil {
		return true
	}
	if opts.update && destInfo.ModTime().After(info.ModTime()) {
		return false
	}
	if destInfo.Size() != info.Size() {
		return true
	}
	// Filesystem timestamp resolution differs between NTFS and FAT, so
	// times within a second of each other count as equal, the same
	// tolerance rsync applies.
	diff := destInfo.ModTime().Sub(info.ModTime())
	if diff < 0 {
		diff = -diff
	}
	return diff.Seconds() >= 1
}

// deleteExtraneous removes destination entries with no counterpart in
// the source, deepest first so directories are empty by the time they
// are removed.
func deleteExtraneous(ctx *command.Context, destRoot string, present map[string]bool, opts *rsyncOptions, stats *rsyncStats) bool {
	var extraneous []string
	err := filepath.Walk(destRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		rel, relErr := filepath.Rel(destRoot, path)
		if relErr != nil || rel == "." {
			return nil
		}
		if !present[filepath.ToSlash(rel)] {
			extraneous = append(extraneous, path)
		}
		return nil
	})
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "rsync: %s\n", err)
		return false
	}

	// Longest path first puts children ahead of their parents.
	sort.Slice(extraneous, func(i, j int) bool { return len(extraneous[i]) > len(extraneous[j]) })

	ok := true
	for _, path := range extraneous {
		rel, _ := filepath.Rel(destRoot, path)
		if opts.verbose {
			fmt.Fprintf(ctx.Stdout, "deleting %s\n", filepath.ToSlash(rel))
		}
		stats.filesDeleted++
		if opts.dryRun {
			continue
		}
		if err := os.RemoveAll(path); err != nil {
			output.Errorf(ctx.Stderr, "rsync", "cannot delete", filepath.ToSlash(rel), err)
			ok = false
		}
	}
	return ok
}

// matchesExclude applies a single --exclude pattern against the relative
// path and each of its components, which is how rsync's simplest
// patterns behave.
func matchesExclude(rel, pattern string) bool {
	slashed := filepath.ToSlash(rel)
	if ok, _ := filepath.Match(pattern, slashed); ok {
		return true
	}
	for _, part := range strings.Split(slashed, "/") {
		if ok, _ := filepath.Match(pattern, part); ok {
			return true
		}
	}
	return false
}

func printRsyncSummary(ctx *command.Context, stats *rsyncStats, opts *rsyncOptions) {
	size := fmt.Sprintf("%d", stats.bytesCopied)
	if opts.human {
		size = output.HumanSize(stats.bytesCopied)
	}
	prefix := ""
	if opts.dryRun {
		prefix = "(dry run) "
	}
	fmt.Fprintf(ctx.Stdout, "\n%s%d file(s) considered, %d transferred, %d director(ies) created, %d deleted\n",
		prefix, stats.filesConsidered, stats.filesCopied, stats.dirsCreated, stats.filesDeleted)
	fmt.Fprintf(ctx.Stdout, "total transferred size: %s bytes\n", size)
}

// isRemoteSpec reports whether a path names something on another host.
// A Windows drive letter is not a host: the single character before the
// colon distinguishes "C:\src" from "server:/src", and a colon appearing
// after a path separator belongs to the path.
func isRemoteSpec(s string) bool {
	if strings.HasPrefix(s, "rsync://") || strings.Contains(s, "::") {
		return true
	}
	colon := strings.IndexByte(s, ':')
	if colon <= 1 {
		return false
	}
	return !strings.ContainsAny(s[:colon], `/\`)
}

func init() { command.Register(rsyncCommand{}) }
